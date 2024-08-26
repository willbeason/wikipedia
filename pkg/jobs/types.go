package jobs

import (
	"context"
	"sync"
)

// A Job represents a unit of work which may be run concurrently with other Jobs.
// The caller is responsible for handling errors and cancelling the job's context
// if appropriate.
//
// Jobs are blocking and should generally be run in goroutines.
type Job func(context.Context, chan<- error)

// newJob creates a new Job which may be executed multiple times concurrently.
func newJob(wg *sync.WaitGroup, first *sync.Once, job Job) Job {
	return func(ctx context.Context, errs chan<- error) {
		wg.Add(1)
		first.Do(wg.Done)
		defer wg.Done()

		job(ctx, errs)
	}
}

// Map processes objects of type T to type U.
type Map[T, U any] func(<-chan T) (*sync.WaitGroup, Job, <-chan U)
