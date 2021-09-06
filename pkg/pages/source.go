package pages

import (
	"context"
	"math"
	"sync"

	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
)

func StreamDB(inDBPath string, parallel int) Source {
	return func(ctx context.Context, errs chan<- error) (<-chan *documents.Page, error) {
		var (
			wg  *sync.WaitGroup
			err error
		)

		inDB := db.NewRunner(inDBPath, parallel)
		pages := make(chan *documents.Page, jobs.WorkBuffer)

		wg, err = inDB.Process(ctx, documents.ReadPages(pages), errs)
		if err != nil {
			return nil, err
		}

		go func() {
			wg.Wait()
			close(pages)
		}()

		return pages, nil
	}
}

func StreamDBKeys(inDBPath string, parallel int, keys []uint) Source {
	return func(ctx context.Context, errs chan<- error) (<-chan *documents.Page, error) {
		var (
			wg  *sync.WaitGroup
			err error
		)

		inDB := db.NewRunner(inDBPath, parallel)
		pages := make(chan *documents.Page, jobs.WorkBuffer)

		wg, err = inDB.ProcessIDs(ctx, documents.ReadPages(pages), toUint32Chan(keys), errs)
		if err != nil {
			return nil, err
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
