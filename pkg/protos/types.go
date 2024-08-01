package protos

import (
	"context"
	"sync"

	"google.golang.org/protobuf/proto"
)

// ID is a supertype for organizing information tied to particular articles.
// This allows having multiple sets of computed information about each article
// without needing to include it all in one object/database.
type ID interface {
	proto.Message
	ID() uint32
}

type Sink func(context.Context, <-chan ID, chan<- error) (*sync.WaitGroup, error)
