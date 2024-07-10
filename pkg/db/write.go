package db

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/willbeason/wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func WriteProto(db *badger.DB) func(p protos.ID) error {
	return func(p protos.ID) error {
		return db.Update(write(p))
	}
}

func write(m protos.ID) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := toKey(m.ID())

		bytes, err := proto.Marshal(m)
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
