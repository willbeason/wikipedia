package pages

import (
	"context"
	"fmt"
	"math"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
)

type Source[T any, PT interface {
	*T
	proto.Message
}] func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan PT, error)

func StreamDB[T any, PT interface {
	*T
	proto.Message
}](
	inDBPath string,
	parallel int,
) Source[T, PT] {
	return func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan PT, error) {
		inDB := db.NewRunner(inDBPath, parallel)
		pages := make(chan PT, jobs.WorkBuffer)

		wg, err := inDB.Process(ctx, cancel, documents.ReadPages(pages))
		if err != nil {
			return nil, fmt.Errorf("streaming database: %w", err)
		}

		go func() {
			wg.Wait()
			close(pages)
		}()

		return pages, nil
	}
}

func StreamDBKeys[T any, PT interface {
	*T
	proto.Message
}](
	inDBPath string,
	parallel int,
	keys []uint,
) Source[T, PT] {
	return func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan PT, error) {
		var (
			wg  *sync.WaitGroup
			err error
		)

		inDB := db.NewRunner(inDBPath, parallel)
		pages := make(chan PT, jobs.WorkBuffer)

		wg, err = inDB.ProcessIDs(ctx, cancel, documents.ReadPages(pages), toUint32Chan(keys))
		if err != nil {
			return nil, fmt.Errorf("streaming keys %v: %w", keys, err)
		}

		go func() {
			wg.Wait()
			close(pages)
		}()

		return pages, nil
	}
}

func toUint32Chan(ids []uint) chan uint32 {
	result := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for _, id := range ids {
			if id > math.MaxUint32 {
				panic(id)
			}

			result <- uint32(id)
		}

		close(result)
	}()

	return result
}
