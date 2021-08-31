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
