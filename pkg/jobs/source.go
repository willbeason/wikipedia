package jobs

import (
	"sync"
)

type SourceFn[T any] func(chan<- T) Job

// A Source generates a channel of objects of some type T.
// A Source may be called multiple times concurrently, returning a new
// Job and channel each time.
type Source[T any] func() (*sync.WaitGroup, Job, <-chan T)

func NewSource[T any](sourceFn SourceFn[T]) Source[T] {
	return func() (*sync.WaitGroup, Job, <-chan T) {
		in := make(chan T)
		wg := &sync.WaitGroup{}
		// Ensure the job is initiated at least once before waiting succeeds.
		wg.Add(1)
		first := &sync.Once{}

		job := sourceFn(in)
		go func() {
			// Close once the Job has been initiated at least once and all outstanding
			// Jobs have completed.
			wg.Wait()
			close(in)
		}()

		return wg, newJob(wg, first, job), in
	}
}
