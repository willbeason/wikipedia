package jobs

import (
	"context"
	"sync"
)

type SourceFn[T any] func(chan<- T) Job

// A Source generates a channel of objects of some type T.
// A Source may be called multiple times concurrently, returning a new
// Job and channel each time.
type Source[T any] func() (*sync.WaitGroup, Job, <-chan T)

func NewSource[T any](sourceFn SourceFn[T]) Source[T] {
	return func() (*sync.WaitGroup, Job, <-chan T) {
		in := make(chan T, DefaultBuffer)
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

type KV[K comparable, V any] struct {
	Key   K
	Value V
}

func MapSourceFn[K comparable, V any](m map[K]V) SourceFn[KV[K, V]] {
	return func(kvs chan<- KV[K, V]) Job {
		return func(ctx context.Context, errors chan<- error) {
			for k, v := range m {
				if ctx.Err() != nil {
					break
				}
				kvs <- KV[K, V]{Key: k, Value: v}
			}
		}
	}
}

func SliceSourceFn[V any](m []V) SourceFn[V] {
	return func(vs chan<- V) Job {
		return func(ctx context.Context, errors chan<- error) {
			for _, v := range m {
				if ctx.Err() != nil {
					break
				}
				vs <- v
			}
		}
	}
}
