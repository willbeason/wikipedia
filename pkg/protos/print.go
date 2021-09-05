package protos

import (
	"context"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/documents"
	"sync"
)

// PrintProtos prints passed protos as JSON to the stdout.
// Returns a WaitGroup which finishes after all protos have been printed, or if
// an error is encountered.
func PrintProtos(_ context.Context, ps <-chan ID, errs chan<- error) (*sync.WaitGroup, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		printProtos(ps, errs)
	}()

	return &wg, nil
}

var _ Sink = PrintProtos

func printProtos(ps <-chan ID, errs chan<- error) {
	for p := range ps {
		err := printProto(p)
		if err != nil {
			errs <- err
			return
		}
	}
}

func printProto(p ID) error {
	page := p.(*documents.Page)

	//bytes, err := protojson.MarshalOptions{Indent: "  "}.Marshal(p)
	//if err != nil {
	//	return err
	//}

	fmt.Println(string(page.Text))
	fmt.Println()

	return nil
}
