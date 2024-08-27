package jobs

import (
	"context"
	"sync"
)

type MapFn[IN, OUT any] func(<-chan IN, chan<- OUT) Job

// Map processes objects of type T to type U.
type Map[IN, OUT any] func(<-chan IN) (*sync.WaitGroup, Job, <-chan OUT)

func NewMap[IN, OUT any](mapFn MapFn[IN, OUT]) Map[IN, OUT] {
	return func(in <-chan IN) (*sync.WaitGroup, Job, <-chan OUT) {
		out := make(chan OUT, DefaultBuffer)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		first := &sync.Once{}

		job := mapFn(in, out)
		go func() {
			wg.Wait()
			close(out)
		}()

		return wg, newJob(wg, first, job), out
	}
}

// Convert constructs a MapFn where there is a 1:1 relationship between in and out.
func Convert[IN, OUT any](fn func(IN) (OUT, error)) MapFn[IN, OUT] {
	return func(in <-chan IN, out chan<- OUT) Job {
		return func(ctx context.Context, errs chan<- error) {
			for item := range in {
				result, err := fn(item)
				if err != nil {
					errs <- err
					continue
				}

				select {
				case <-ctx.Done():
					return
				case out <- result:
				}
			}
		}
	}
}

// ReduceToMany reduces a stream of IN to one or more OUT.
// Initializes a number of OUT equal to the number of times the Job is called.
func ReduceToMany[IN, OUT any](newOut func() OUT, fn func(IN, OUT) error) MapFn[IN, OUT] {
	return func(ins <-chan IN, outs chan<- OUT) Job {
		return func(ctx context.Context, errs chan<- error) {
			out := newOut()
			defer func() {
				select {
				case <-ctx.Done():
				case outs <- out:
				}
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case in, ok := <-ins:
					if !ok {
						return
					}
					err := fn(in, out)
					if err != nil {
						errs <- err
					}
				}
			}
		}
	}
}

// ReduceToOne reduces a stream of IN to a single OUT.
func ReduceToOne[IN, OUT any](newOut func() OUT, fn func(IN, OUT) error) MapFn[IN, OUT] {
	return func(ins <-chan IN, outs chan<- OUT) Job {
		return func(ctx context.Context, errs chan<- error) {
			out := newOut()
			defer func() {
				select {
				case <-ctx.Done():
				case outs <- out:
				}
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case in, ok := <-ins:
					if !ok {
						return
					}
					err := fn(in, out)
					if err != nil {
						errs <- err
					}
				}
			}
		}
	}
}

func MakeMap[K comparable, V any]() map[K]V {
	return make(map[K]V)
}
