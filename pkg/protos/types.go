package protos

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// ID is a supertype for organizing information tied to particular articles.
// This allows having multiple sets of computed information about each article
// without needing to include it all in one object/database.
type ID interface {
	proto.Message
	ID() uint32
}

type Proto[T any] interface {
	*T
	proto.Message
}

type ProtoID[T any] interface {
	*T
	proto.Message
	ID
}

type Sink[IN any, PIN Proto[IN], OUT any] func(
	context.Context,
	context.CancelCauseFunc,
	<-chan PIN,
	chan<- error,
) (<-chan OUT, error)
