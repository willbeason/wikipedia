package jobs

import (
	"context"
	"sync"
)

type Run func(ctx context.Context) error

type Runner struct {
	parallel int
}

func NewRunner(parallel int) *Runner {
	return &Runner{parallel: parallel}
}

func (r *Runner) Run(ctx context.Context, run Run, errs chan<- error) *sync.WaitGroup {
	ctx, cancel := context.WithCancel(ctx)

	wg := sync.WaitGroup{}
	wg.Add(r.parallel)

	go func() {
		defer cancel()
		wg.Wait()
	}()

	for i := 0; i < r.parallel; i++ {
		go func() {
			err := run(ctx)
			if err != nil {
				errs <- err

				cancel()
			}

			wg.Done()
		}()
	}

	return &wg
}
