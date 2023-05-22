package db

import (
	"context"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

// ProcessIDs runs process on each id's corresponding value.
// process must be thread-safe as it may be called simultaneously by multiple
// threads.
//
// Returns a WaitGroup which finishes after the last id has been processed.
func (r *Runner) ProcessIDs(ctx context.Context, cancel context.CancelCauseFunc, process Process, ids <-chan uint32) (*sync.WaitGroup, error) {
	db, err := badger.Open(badger.DefaultOptions(r.path))
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(r.parallel)

	go func() {
		defer func() {
			err2 := db.Close()
			if err2 != nil {
				cancel(err2)
			}
		}()

		wg.Wait()
	}()

	for i := 0; i < r.parallel; i++ {
		go func() {
			processIDs(ctx, cancel, db, ids, process)
			wg.Done()
		}()
	}

	return &wg, nil
}

func processIDs(ctx context.Context, cancel context.CancelCauseFunc, db *badger.DB, ids <-chan uint32, process Process) {
	for id := range ids {
		select {
		case <-ctx.Done():
			return
		default:
			err := db.View(processID(id, process))
			if err != nil {
				cancel(err)
			}
		}
	}

	return
}

func processID(id uint32, process Process) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := toKey(id)

		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return process(value)
	}
}
