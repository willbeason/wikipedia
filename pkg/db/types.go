package db

import (
	"google.golang.org/protobuf/proto"
)

type Process func(value []byte) error

type MessageID interface {
	proto.Message
	ID() uint32
}
