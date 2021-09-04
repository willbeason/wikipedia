package jobs

import (
	"context"

	"github.com/willbeason/wikipedia/pkg/documents"
)

type Page func(page *documents.Page) error

func PageWorker(pages <-chan *documents.Page, job func(*documents.Page) error) Run {
	return func(ctx context.Context) error {
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
}
