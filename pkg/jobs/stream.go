package jobs

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

var ErrStream = errors.New("executing workflow")

const WorkBuffer = 100

type Work func() error

type WorkQueue <-chan Work

type Runner struct {
	// parallel is the maximum number of goroutines launched for each call to Run().
	parallel int
}

type RunnerOpt func(*Runner)

func Parallel(parallel int) func(*Runner) {
	return func(runner *Runner) {
		runner.parallel = parallel
	}
}

func defaultOpts() []RunnerOpt {
	return []RunnerOpt{Parallel(runtime.NumCPU())}
}

func NewRunner(opts ...RunnerOpt) *Runner {
	runner := &Runner{}
	for _, opt := range defaultOpts() {
		opt(runner)
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

// Run launches goroutines that execute work in workQueue.
// Returns a WaitGroup that completes when all work is done.
// Halts and marks WaitGroup as done on error.
func (r *Runner) Run(ctx context.Context, cancel context.CancelCauseFunc, workQueue WorkQueue) *sync.WaitGroup {
	wg := sync.WaitGroup{}

	for range r.parallel {
		wg.Add(1)

		go func() {
		DONE:
			for work := range workQueue {
				select {
				case <-ctx.Done():
					break DONE
				default:
					err := work()
					if err != nil {
						cancel(err)
						break DONE
					}
				}
			}

			wg.Done()
		}()
	}

	return &wg
}

// Deprecated: MapOld mutates an input channel, optionally changing the type.
func MapOld[FROM, TO any](buffer int, in <-chan FROM, fn func(FROM) (TO, error)) (chan TO, WorkQueue) {
	out := make(chan TO, buffer)
	work := make(chan Work, buffer)

	go func() {
		for i := range in {
			x := i
			work <- func() error {
				o, err := fn(x)
				if err != nil {
					return err
				}

				out <- o
				return nil
			}
		}

		close(work)
	}()

	return out, work
}

// Deprecated: ForEachOld performs an action for each item, returning nothing.
func ForEachOld[FROM any](buffer int, in <-chan FROM, fn func(FROM) error) WorkQueue {
	work := make(chan Work, buffer)

	go func() {
		for i := range in {
			x := i
			work <- func() error {
				err := fn(x)
				if err != nil {
					return err
				}

				return nil
			}
		}

		close(work)
	}()

	return work
}

// Reduce consumes a channel.
func Reduce[IN any](ctx context.Context, buffer int, in <-chan IN, fn func(IN) error) WorkQueue {
	work := make(chan Work, buffer)

	go func() {
		for i := range in {
			select {
			case <-ctx.Done():
				break
			default:
				x := i
				work <- func() error {
					return fn(x)
				}
			}
		}

		close(work)
	}()

	return work
}
