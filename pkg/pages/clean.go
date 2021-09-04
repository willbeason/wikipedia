package pages

import (
	"context"

	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/protos"
)

type Source func(ctx context.Context, errs chan<- error) (<-chan *documents.Page, error)

func Run(ctx context.Context, source Source, parallel int, job func(chan<- protos.ID) jobs.Page, sink protos.Sink) error {
	errs, errsWg := jobs.Errors()

	pages, err := source(ctx, errs)
	if err != nil {
		return err
	}

	out := make(chan protos.ID, jobs.WorkBuffer)
	runner := jobs.NewRunner(parallel)
	worker := jobs.PageWorker(pages, job(out))
	cleanWg := runner.Run(ctx, worker, errs)

	go func() {
		cleanWg.Wait()
		close(out)
	}()

	outWg, err := sink(ctx, out, errs)
	if err != nil {
		return err
	}

	outWg.Wait()
	close(errs)

	errsWg.Wait()

	return nil
}

func CleanPages(cleaned chan<- protos.ID) jobs.Page {
	return func(page *documents.Page) error {
		page.Text = nlp.CleanArticle(page.Text)
		cleaned <- page

		return nil
	}
}
