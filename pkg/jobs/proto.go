package jobs

import (
	"sync"

	"google.golang.org/protobuf/proto"
)

type NewProto func() proto.Message

type Proto func(proto proto.Message) error


func RunProto(parallel int, job Proto, work <-chan proto.Message, errs chan<- error) *sync.WaitGroup {
	workWg := sync.WaitGroup{}

	for i := 0; i < parallel; i++ {
		workWg.Add(1)

		go func() {
			runProtoWorker(job, work, errs)
			workWg.Done()
		}()
	}

	return &workWg
}

func runProtoWorker(job Proto, work <-chan proto.Message, errs chan<- error) {
	for pb := range work {
		err := job(pb)
		if err != nil {
			errs <- err
		}
	}
}
