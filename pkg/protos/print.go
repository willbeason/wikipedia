package protos

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"google.golang.org/protobuf/encoding/prototext"
)

// PrintProtos prints passed protos as JSON to the stdout.
// Returns a WaitGroup which finishes after all protos have been printed, or if
// an error is encountered.
func PrintProtos(_ context.Context, ps <-chan ID, errs chan<- error) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := printProtos(ps)
		if err != nil {
			errs <- err
		}
	}()

	return &wg, nil
}

var _ Sink = PrintProtos

func printProtos(ps <-chan ID) error {
	for p := range ps {
		err := printProto(p)
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrPrint = errors.New("printing proto")

func printProto(p ID) error {
	bytes, err := prototext.MarshalOptions{Indent: "  "}.Marshal(p)
	if err != nil {
		return fmt.Errorf("%w with ID %v", ErrPrint, p.ID())
	}

	fmt.Println(string(bytes))
	fmt.Println()

	return nil
}
