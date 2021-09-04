package jobs

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/dgraph-io/badger/v3"
	"google.golang.org/protobuf/proto"
)

const WorkBuffer = 1000

func NewProtoWork() chan proto.Message {
	return make(chan proto.Message, WorkBuffer)
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
