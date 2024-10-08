package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/pb"
	"github.com/dgraph-io/ristretto/z"
)

// Process runs process on each Value in the database.
// process must be thread-safe as it may be called simultaneously by multiple
// threads.
//
// Returns a WaitGroup which finishes after the last Value from the DB has been
// processed.
func (r *Runner) Process(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	process Process,
) (*sync.WaitGroup, error) {
	dbOpts := badger.
		DefaultOptions(r.path).
		WithMetricsEnabled(false).
		WithLoggingLevel(badger.WARNING).
		WithReadOnly(true)

	db, err := badger.Open(dbOpts)
	if err != nil {
		fmt.Printf("Recovering DB %q from crash\n", r.path)
		// Open/close with not readonly. This allows Badger to replay any logs.
		dbOpts2 := badger.
			DefaultOptions(r.path)
		db2, err2 := badger.Open(dbOpts2)
		if err2 != nil {
			return nil, fmt.Errorf("opening dummy Badger DB %q: %w", r.path, err)
		}

		err2 = db2.Close()
		if err2 != nil {
			return nil, fmt.Errorf("closing dummy Badger DB %q: %w", r.path, err)
		}

		// Open properly with readonly.
		db, err = badger.Open(dbOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("opening Badger DB %q: %w", r.path, err)
	}

	lists := make(chan *pb.KVList)

	go func() {
		defer func() {
			close(lists)
			err2 := db.Close()
			if err2 != nil {
				cancel(err2)
			}
		}()

		err = r.process(ctx, db, lists)
		if err != nil {
			cancel(err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	// While technically threadsafe to parallelize, it does not result in increased performance, only increased memory
	// usage.
	go func() {
		for list := range lists {
			select {
			case <-ctx.Done():
				break
			default:
				kvs := list.GetKv()
				for _, kv := range kvs {
					value := kv.GetValue()

					processErr := process(value)
					if processErr != nil {
						cancel(processErr)
					}
				}
			}
		}
		wg.Done()
	}()

	return &wg, nil
}

func (r *Runner) process(ctx context.Context, db *badger.DB, lists chan<- *pb.KVList) error {
	stream := db.NewStream()
	stream.NumGo = r.parallel
	stream.Send = send(lists)

	err := stream.Orchestrate(ctx)
	if err != nil {
		return fmt.Errorf("stream.Orchestrate: %w", err)
	}

	return nil
}

func send(lists chan<- *pb.KVList) func(buf *z.Buffer) error {
	return func(buf *z.Buffer) error {
		list, err := badger.BufferToKVList(buf)
		if err != nil {
			return fmt.Errorf("badger.BufferToKVList: %w", err)
		}
		lists <- list

		return nil
	}
}
