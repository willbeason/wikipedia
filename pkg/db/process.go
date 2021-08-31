package db

import (
	"context"
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
func (r *Runner) Process(ctx context.Context, process Process, errs chan<- error) (*sync.WaitGroup, error) {
	dbOpts := badger.DefaultOptions(r.path).WithNumGoroutines(r.parallel)

	db, err := badger.Open(dbOpts)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer closeDB(db, errs)

		err = r.process(ctx, db, process)
		if err != nil {
			errs <- err
		}
	}()

	return &wg, nil
}

func (r *Runner) process(ctx context.Context, db *badger.DB, process Process) error {
	stream := db.NewStream()
	stream.NumGo = r.parallel
	stream.Send = send(process)

	return stream.Orchestrate(ctx)
}

func send(process Process) func(buf *z.Buffer) error {
	return func(buf *z.Buffer) error {
		list, err := badger.BufferToKVList(buf)
		if err != nil {
			return err
		}

		for _, kv := range list.GetKv() {
			value := kv.GetValue()

			err = process(value)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
