package pages

import (
	"context"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
)

func StreamDB(inDBPath string, parallel int) func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan *documents.Page, error) {
	return func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan *documents.Page, error) {
		inDB := db.NewRunner(inDBPath, parallel)
		pages := make(chan *documents.Page, jobs.WorkBuffer)

		wg, err := inDB.Process(ctx, cancel, documents.ReadPages(pages))
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
