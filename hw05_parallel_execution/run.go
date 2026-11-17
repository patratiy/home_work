package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
// If m <= 0, errors are ignored and all tasks are executed.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 || n <= 0 {
		return nil
	}

	tasksCh := make(chan Task)

	var (
		wg       sync.WaitGroup
		errCount atomic.Int32
	)

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				if err := task(); err != nil && m > 0 {
					errCount.Add(1)
				}
			}
		}()
	}

	for _, task := range tasks {
		if m > 0 && errCount.Load() >= int32(m) {
			break
		}
		tasksCh <- task
	}
	close(tasksCh)

	wg.Wait()

	if m > 0 && errCount.Load() >= int32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
