package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/ristretto/z"
)

// Process runs process on each Value in the database.
// process must be thread-safe as it may be called simultaneously by multiple
// threads.
//
// Returns a WaitGroup which finishes after the last Value from the DB has been
// processed.
func (r *Runner) Process(ctx context.Context, cancel context.CancelCauseFunc, process Process) (*sync.WaitGroup, error) {
	dbOpts := badger.DefaultOptions(r.path).WithNumGoroutines(r.parallel)

	db, err := badger.Open(dbOpts)
	if err != nil {
		return nil, fmt.Errorf("opening Badger DB %q: %w", r.path, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			err2 := db.Close()
			if err2 != nil {
				cancel(err2)
			}

			wg.Done()
		}()

		err = r.process(ctx, db, process)
		if err != nil {
			cancel(err)
		}
	}()

	return &wg, nil
}

func (r *Runner) process(ctx context.Context, db *badger.DB, process Process) error {
	stream := db.NewStream()
	stream.NumGo = r.parallel
	stream.Send = send(process)

	err := stream.Orchestrate(ctx)
	if err != nil {
		return fmt.Errorf("stream.Orchestrate: %w", err)
	}

	return nil
}

func send(process Process) func(buf *z.Buffer) error {
	return func(buf *z.Buffer) error {
		list, err := badger.BufferToKVList(buf)
		if err != nil {
			return fmt.Errorf("badger.BufferToKVList: %w", err)
		}

		kvs := list.GetKv()
		for _, kv := range kvs {
			value := kv.GetValue()

			err = process(value)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
