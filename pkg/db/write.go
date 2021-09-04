package db

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/willbeason/wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

// Write writes all protos to the Runner's DB.
func (r *Runner) Write() protos.Sink {
	return func(ctx context.Context, ps <-chan protos.ID, errs chan<- error) (*sync.WaitGroup, error) {
		db, err := badger.Open(badger.DefaultOptions(r.path))
		if err != nil {
			return nil, err
		}

		wg := sync.WaitGroup{}
		wg.Add(r.parallel)
		ctx, cancel := context.WithCancel(ctx)

		go func() {
			defer cancel()
			defer closeDB(db, errs)
			wg.Wait()

			err = runGC(db)
			if err != nil {
				errs <- err
			}
		}()

		for i := 0; i < r.parallel; i++ {
			go func() {
				perr := writeProtos(ctx, ps, db)
				if perr != nil {
					errs <- perr

					cancel()
				}

				wg.Done()
			}()
		}

		return &wg, nil
	}
}

func writeProtos(ctx context.Context, ps <-chan protos.ID, db *badger.DB) error {
	for p := range ps {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := db.Update(write(p))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func write(m protos.ID) func(txn *badger.Txn) error {
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
		// No error indicates garbage was collected, so run garbage collection again
		// until we get ErrNoRewrite.
	}

	return nil
}
