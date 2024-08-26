package jobs

import "sync"

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
