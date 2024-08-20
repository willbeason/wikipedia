package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

// ProcessIDs runs process on each id's corresponding value.
// process must be thread-safe as it may be called simultaneously by multiple
// threads.
//
// Returns a WaitGroup which finishes after the last id has been processed.
func (r *Runner) ProcessIDs(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	process Process,
	ids <-chan uint32,
) (*sync.WaitGroup, error) {
	dbOpts := badger.
		DefaultOptions(r.path).
		WithMetricsEnabled(false).
		WithLoggingLevel(badger.WARNING).
		WithReadOnly(true)

	db, err := badger.Open(dbOpts)
	if err != nil {
		return nil, fmt.Errorf("opening Badger DB %q: %w", r.path, err)
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

	for range r.parallel {
		go func() {
			processIDs(ctx, cancel, db, ids, process)
			wg.Done()
		}()
	}

	return &wg, nil
}

func processIDs(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	db *badger.DB,
	ids <-chan uint32,
	process Process,
) {
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
}

func processID(id uint32, process Process) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := toKey(id)

		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("getting ID %d: %w", id, err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			// Only returns an error if the item is an error.
			return fmt.Errorf("getting ID %d: %w", id, err)
		}

		return process(value)
	}
}
