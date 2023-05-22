package protos

import (
	"context"
	"sync"
)

// NoSink does nothing with output.
func NoSink(_ context.Context, ps <-chan ID, errs chan<- error) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for range ps {
		}
	}()

	return &wg, nil
}
