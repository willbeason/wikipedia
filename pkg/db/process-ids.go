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
func (r *Runner) ProcessIDs(ctx context.Context, process Process, ids <-chan uint32, errs chan<- error) (*sync.WaitGroup, error) {
	db, err := badger.Open(badger.DefaultOptions(r.path))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	workWg := sync.WaitGroup{}
	workWg.Add(r.parallel)

	go func() {
		defer cancel()
		defer closeDB(db, errs)
		workWg.Wait()
	}()

	for i := 0; i < r.parallel; i++ {
		go func() {
			perr := processIDs(ctx, db, ids, process)
			if perr != nil {
				cancel()
				errs <- perr
			}

			workWg.Done()
		}()
	}

	return &workWg, nil
}

func processIDs(ctx context.Context, db *badger.DB, ids <-chan uint32, process Process) error {
	for id := range ids {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := db.View(processID(id, process))
			if err != nil {
				return err
			}
		}
	}

	return nil
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
