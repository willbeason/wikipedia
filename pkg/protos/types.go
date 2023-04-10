package protos

import (
	"context"
	"sync"

	"google.golang.org/protobuf/proto"
)

type ID interface {
	proto.Message
	ID() uint32
}

type Sink func(context.Context, <-chan ID, chan<- error) (*sync.WaitGroup, error)

func NoSink(ctx context.Context, ps <-chan ID, _ chan<- error) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range ps {
			if <-ctx.Done(); true {
				return
			}
		}
	}()

	return &wg, nil
}
