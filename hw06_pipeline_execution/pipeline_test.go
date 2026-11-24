package hw06pipelineexecution

import (
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestPipelineEdgeCases(t *testing.T) {
	t.Run("empty stages list", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]interface{}, 0)
		for s := range ExecutePipeline(in, nil) {
			result = append(result, s)
		}

		require.Equal(t, []interface{}{1, 2, 3}, result)
	})

	t.Run("empty input channel", func(t *testing.T) {
		in := make(Bi)
		close(in)

		stages := []Stage{
			func(in In) Out {
				out := make(Bi)
				go func() {
					defer close(out)
					for v := range in {
						out <- v
					}
				}()
				return out
			},
		}

		result := make([]interface{}, 0)
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s)
		}

		require.Len(t, result, 0)
	})

	t.Run("done already closed", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		close(done)

		stages := []Stage{
			func(in In) Out {
				out := make(Bi)
				go func() {
					defer close(out)
					for v := range in {
						out <- v
					}
				}()
				return out
			},
		}

		go func() {
			in <- 1
			close(in)
		}()

		result := make([]interface{}, 0)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s)
		}

		require.Len(t, result, 0)
	})
}

func TestPipelineStopsGoroutinesWithinTimeout(t *testing.T) {
	const (
		waitFor = time.Second
		tick    = 10 * time.Millisecond
	)

	tests := []struct {
		name              string
		stagesCount       int
		doneAlreadyClosed bool
	}{
		{
			name:              "cancel after stages started",
			stagesCount:       3,
			doneAlreadyClosed: false,
		},
		{
			name:              "done already closed",
			stagesCount:       1,
			doneAlreadyClosed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := make(Bi, 1)
			done := make(Bi)
			var running atomic.Int32
			var stagesStarted atomic.Int32
			var received atomic.Int32
			baseline := runtime.NumGoroutine()

			startGoroutine := func(body func()) {
				running.Add(1)
				go func() {
					defer running.Add(-1)
					body()
				}()
			}

			var closeDoneOnce sync.Once
			closeDone := func() {
				closeDoneOnce.Do(func() {
					close(done)
				})
			}
			t.Cleanup(closeDone)

			if tc.doneAlreadyClosed {
				closeDone()
			}

			// Keep the source open until cancellation so stage goroutines are active.
			startGoroutine(func() {
				in <- 1
				<-done
				close(in)
			})

			stages := make([]Stage, 0, tc.stagesCount)
			for i := 0; i < tc.stagesCount; i++ {
				stages = append(stages, func(in In) Out {
					out := make(Bi, 1)
					startGoroutine(func() {
						stagesStarted.Add(1)
						defer close(out)
						for v := range in {
							select {
							case out <- v:
							case <-done:
								return
							}
						}
					})
					return out
				})
			}

			output := ExecutePipeline(in, done, stages...)
			outputClosed := make(chan struct{})
			// Drain the output and report when ExecutePipeline closes it.
			startGoroutine(func() {
				defer close(outputClosed)
				for range output {
					received.Add(1)
				}
			})

			require.Eventually(t, func() bool {
				return stagesStarted.Load() == int32(tc.stagesCount)
			}, waitFor, tick, "not all stage goroutines started")

			// Close done after all stage goroutines have started unless it is already closed.
			closeDone()

			timer := time.NewTimer(waitFor)
			select {
			case <-outputClosed:
				timer.Stop()
			case <-timer.C:
				require.Fail(t, "pipeline output did not close within timeout")
			}

			requireNoPipelineGoroutines(t, baseline, &running)
			require.LessOrEqual(t, received.Load(), int32(1), "pipeline emitted more values than supplied")
		})
	}
}

func requireNoPipelineGoroutines(t *testing.T, baseline int, running *atomic.Int32) {
	t.Helper()

	waitFor := time.Second
	tick := 10 * time.Millisecond
	timer := time.NewTimer(waitFor)
	defer timer.Stop()
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		if running.Load() == 0 && runtime.NumGoroutine() <= baseline {
			return
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			require.Fail(t, "pipeline goroutines did not stop within timeout",
				"running=%d, goroutines=%d, baseline=%d",
				running.Load(), runtime.NumGoroutine(), baseline)
		}
	}
}
