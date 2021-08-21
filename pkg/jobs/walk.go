package jobs

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/ristretto/z"
	"google.golang.org/protobuf/proto"
)

const WorkBuffer = 100

func NewProtoWork() chan proto.Message {
	return make(chan proto.Message, WorkBuffer)
}

func Walk(ctx context.Context, db *badger.DB, newProto NewProto, parallel int, errs chan<- error) <-chan proto.Message {
	work := NewProtoWork()

	stream := db.NewStream()
	stream.NumGo = parallel
	stream.Send = walkStream(newProto, work)

	go func() {
		err := stream.Orchestrate(ctx)
		if err != nil {
			errs <- err
		}

		close(work)
	}()

	return work
}

func walkStream(newProto NewProto, work chan<- proto.Message) func(buf *z.Buffer) error {
	return func(buf *z.Buffer) error {
		list, err := badger.BufferToKVList(buf)
		if err != nil {
			return err
		}

		for _, kv := range list.GetKv() {
			value := kv.GetValue()
			p := newProto()

			err = proto.Unmarshal(value, p)
			if err != nil {
				return err
			}

			work <- p
		}

		return nil
	}
}

// IDs returns a channel of protos.
func IDs(db *badger.DB, newProto NewProto, ids []uint, errs chan<- error) <-chan proto.Message {
	work := NewProtoWork()

	go func() {
		for _, id := range ids {
			p := newProto()

			err := db.View(getProto(id, p))
			if err != nil {
				errs <- err
				continue
			}

			work <- p
		}

		close(work)
	}()

	return work
}

// getProto fills in dst with the proto with id using a badger.Txn.
// id must be a valid uint32.
func getProto(id uint, dst proto.Message) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		if id > math.MaxUint32 {
			return fmt.Errorf("id %d is greater than %d", id, math.MaxUint32)
		}

		key := make([]byte, 4)
		binary.LittleEndian.PutUint32(key, uint32(id))

		var valueBytes []byte

		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("%w: %d", err, id)
		}

		valueBytes, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return proto.Unmarshal(valueBytes, dst)
	}
}
