package jobs

import (
	"context"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
)

type Page func(page *documents.Page) error

func RunPage(ctx context.Context, parallel int, job Page, pages <-chan *documents.Page, errs chan<- error) *sync.WaitGroup {
	ctx, cancel := context.WithCancel(ctx)

	wg := sync.WaitGroup{}
	wg.Add(parallel)

	go func() {
		defer cancel()
		wg.Wait()
	}()

	for i := 0; i < parallel; i++ {
		go func() {
			err := runPageWorker(ctx, job, pages)
			if err != nil {
				errs <- err
				cancel()
			}

			wg.Done()
		}()
	}

	return &wg
}

func runPageWorker(ctx context.Context, job func(*documents.Page) error, pages <-chan *documents.Page) error {
	for page := range pages {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := job(page)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
