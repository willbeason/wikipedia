package protos

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/encoding/prototext"
)

// PrintProtos prints passed protos as JSON to the stdout.
// Returns a WaitGroup which finishes after all protos have been printed, or if
// an error is encountered.
func PrintProtos[IN any, PIN Proto[IN]](
	ctx context.Context,
	cancel context.CancelCauseFunc,
	ps <-chan PIN,
) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		printProtos[IN, PIN](ctx, cancel, ps)

		wg.Done()
	}()

	return &wg, nil
}

func printProtos[IN any, PIN Proto[IN]](ctx context.Context, cancel context.CancelCauseFunc, ps <-chan PIN) {
	for p := range ps {
		select {
		case <-ctx.Done():
			return
		default:
			err := printProto(p)
			if err != nil {
				cancel(err)
			}
		}
	}
}

var ErrPrint = errors.New("printing proto")

func printProto(p proto.Message) error {
	bytes, err := prototext.MarshalOptions{Indent: "  "}.Marshal(p)
	if err != nil {
		if pid, isID := p.(ID); isID {
			return fmt.Errorf("%w with ID %v", ErrPrint, pid.ID())
		} else {
			return fmt.Errorf("%w with proto %+v", ErrPrint, p)
		}
	}

	fmt.Println(string(bytes))
	fmt.Println()

	return nil
}
