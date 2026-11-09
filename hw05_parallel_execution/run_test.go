package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRunIgnoredErrors(t *testing.T) {
	defer goleak.VerifyNone(t)

	for _, errorsLimit := range []int{0, -1} {
		t.Run(fmt.Sprintf("errors limit %d", errorsLimit), func(t *testing.T) {
			tasksCount := 100
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32

			for i := 0; i < tasksCount; i++ {
				err := fmt.Errorf("task %d error", i)
				tasks = append(tasks, func() error {
					atomic.AddInt32(&runTasksCount, 1)
					return err
				})
			}

			err := Run(tasks, 5, errorsLimit)

			require.NoError(t, err)
			require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were run with ignored errors")
		})
	}
}

func TestRunErrorsBelowLimit(t *testing.T) {
	defer goleak.VerifyNone(t)

	const (
		tasksCount  = 20
		errorsCount = 3
	)

	errorsLimit := errorsCount + 2

	tasks := make([]Task, 0, tasksCount)
	var runTasksCount int32

	for i := 0; i < tasksCount; i++ {
		var taskErr error
		if i < errorsCount {
			taskErr = fmt.Errorf("task %d failed", i)
		}

		tasks = append(tasks, func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return taskErr
		})
	}

	err := Run(tasks, 4, errorsLimit)

	require.NoError(t, err)
	require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were run when errors were below the limit")
}

func TestRunConcurrencyWithoutSleep(t *testing.T) {
	defer goleak.VerifyNone(t)

	const (
		tasksCount   = 10
		workersCount = 5
	)

	var started atomic.Int32
	release := make(chan struct{})

	tasks := make([]Task, 0, tasksCount)
	for range tasksCount {
		tasks = append(tasks, func() error {
			started.Add(1)
			<-release
			return nil
		})
	}

	var runErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		runErr = Run(tasks, workersCount, 1)
	}()

	// Exactly workersCount tasks must start and stay blocked, tasks run concurrently.
	require.Eventually(t, func() bool {
		return started.Load() == int32(workersCount)
	}, time.Second, 10*time.Millisecond)

	// While the workers are blocked, no extra tasks will be started.
	require.Equal(t, int32(workersCount), started.Load())

	close(release)
	<-done

	require.NoError(t, runErr)
	require.Equal(t, int32(tasksCount), started.Load(), "not all tasks were completed after release")
}
