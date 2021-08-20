package jobs

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

type GarbageCollector struct {
	period uint64
	db *badger.DB

	count uint64
	Incoming chan struct{}
}

func NewGarbageCollector(db *badger.DB) *GarbageCollector {
	return &GarbageCollector{
		period:   1000000,
		db:       db,
		Incoming: make(chan struct{}, 100),
	}
}

func (gc *GarbageCollector) run(errs chan<- error) {
	for range gc.Incoming {
		gc.count++
		if gc.count % gc.period == 0 {
			err := RunGC(gc.db)
			if err != nil {
				errs <- err
			}
		}
	}
}

func RunGC(db *badger.DB) error {
	for {
		fmt.Println("Running Garbage Collection")
		err := db.RunValueLogGC(0.5)
		if err != nil {
			if !errors.Is(err, badger.ErrNoRewrite) && !errors.Is(err, badger.ErrRejected) {
				return err
			}
			break
		}
	}
	return nil
}
