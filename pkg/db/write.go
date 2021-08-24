package db

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"google.golang.org/protobuf/proto"
)

// Write writes all protos to the DB.
//
// Returns a WaitGroup which finishes after the last proto has been written.
func (r *Runner) Write(protos <-chan MessageID, errs chan<- error) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}

	db, err := badger.Open(badger.DefaultOptions(r.path))
	if err != nil {
		return &wg, err
	}

	for i := 0; i < r.parallel; i++ {
		wg.Add(1)

		go func() {
			writeProtos(protos, db, errs)
			wg.Done()
		}()
	}

	go func() {
		defer closeDB(db, errs)

		wg.Wait()

		err = runGC(db)
		if err != nil {
			errs <- err
		}
	}()

	return &wg, nil
}

func writeProtos(protos <-chan MessageID, db *badger.DB, errs chan<- error) {
	for p := range protos {
		err := db.Update(write(p))
		if err != nil {
			errs <- err
		}
	}
}

func write(m MessageID) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := toKey(m.ID())

		bytes, err := proto.Marshal(m)
		if err != nil {
			return err
		}

		return txn.Set(key, bytes)
	}
}

func runGC(db *badger.DB) error {
	const discardRatio = 0.5

	for {
		fmt.Println("Running Garbage Collection")

		err := db.RunValueLogGC(discardRatio)
		if err != nil {
			if !errors.Is(err, badger.ErrNoRewrite) && !errors.Is(err, badger.ErrRejected) {
				return err
			}

			break
		}
	}

	return nil
}
