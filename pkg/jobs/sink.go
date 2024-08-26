package jobs

import "sync"

type SinkFn[T any] func(<-chan T) Job

// A Sink consumes a channel of objects.
type Sink[T any] func(<-chan T) (*sync.WaitGroup, Job)

func NewSink[T any](sinkFn SinkFn[T]) Sink[T] {
	return func(out <-chan T) (*sync.WaitGroup, Job) {
		wg := &sync.WaitGroup{}
		// Ensure the job is initiated at least once before waiting succeeds.
		wg.Add(1)
		first := &sync.Once{}

		job := sinkFn(out)

		return wg, newJob(wg, first, job)
	}
}
