package jobs

import (
	"encoding/binary"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"google.golang.org/protobuf/proto"
)

type MessageID interface {
	proto.Message
	ID() uint32
}

func WriteProtos(db *badger.DB, parallel int, protos <-chan MessageID, errs chan<- error) *sync.WaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(parallel)

	gc := NewGarbageCollector(db)

	for i := 0; i < parallel; i++ {
		go func() {
			runProtoWriter(db, gc, protos, errs)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(gc.Incoming)
	}()

	gcWg := sync.WaitGroup{}
	gcWg.Add(1)

	go func() {
		gc.run(errs)
		gcWg.Done()
	}()

	return &gcWg
}

func runProtoWriter(db *badger.DB, gc *GarbageCollector, protos <-chan MessageID, errs chan<- error) {
	for p := range protos {
		err := db.Update(WriteProto(p))
		if err != nil {
			errs <- err
			continue
		}

		gc.Incoming <- struct{}{}
	}
}

func WriteProto(m MessageID) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		key := make([]byte, 4)
		binary.LittleEndian.PutUint32(key, m.ID())

		pageBytes, err := proto.Marshal(m)
		if err != nil {
			return err
		}

		return txn.Set(key, pageBytes)
	}
}
