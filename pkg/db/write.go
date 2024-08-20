package db

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/willbeason/wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func WriteProto[PROTO protos.ID](db *badger.DB) func(p PROTO) error {
	return func(p PROTO) error {
		return db.Update(write(p))
	}
}

func write(m protos.ID) func(txn *badger.Txn) error {
	key := toKey(m.ID())
	bytes, err := proto.Marshal(m)

	return func(txn *badger.Txn) error {
		if err != nil {
			return fmt.Errorf("marshalling %d: %w", fromKey(key), err)
		}

		return txn.Set(key, bytes)
	}
}

func RunGC(db *badger.DB) error {
	const discardRatio = 0.5

	for {
		fmt.Println("Running Garbage Collection")

		err := db.RunValueLogGC(discardRatio)
		if err != nil {
			if !errors.Is(err, badger.ErrNoRewrite) && !errors.Is(err, badger.ErrRejected) {
				return fmt.Errorf("running garbage collection: %w", err)
			}

			break
		}
		// No error indicates garbage was collected, so run garbage collection again
		// until we get ErrNoRewrite.
	}

	return nil
}
