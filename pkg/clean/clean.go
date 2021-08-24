package clean

import (
	"context"
	"fmt"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/db"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

func Run(ctx context.Context, inDBPath string, parallel int, pageIDs []uint, outDBPath string) error {
	inDB := db.NewRunner(inDBPath, parallel)
	pages := make(chan *documents.Page)
	errs, errsWg := jobs.Errors()

	err := runReadPages(ctx, pageIDs, inDB, pages, errs)
	if err != nil {
		return err
	}

	cleaned := make(chan db.MessageID, jobs.WorkBuffer)
	runner := jobs.NewRunner(parallel)
	cleanWg := runner.Run(ctx, jobs.PageWorker(cleanPages(cleaned), pages), errs)

	go func() {
		cleanWg.Wait()
		close(cleaned)
	}()

	outWg, err := runWritePages(outDBPath, cleaned, parallel, errs)
	if err != nil {
		return err
	}

	outWg.Wait()
	close(errs)

	errsWg.Wait()

	return nil
}

func runReadPages(ctx context.Context, pageIDs []uint, inDB *db.Runner, pages chan<- *documents.Page, errs chan<- error) error {
	var wg *sync.WaitGroup

	var err error

	if len(pageIDs) == 0 {
		wg, err = inDB.Process(ctx, documents.ReadPages(pages), errs)
	} else {
		wg, err = inDB.ProcessIDs(ctx, documents.ReadPages(pages), toUint32Chan(pageIDs), errs)
	}

	if err != nil {
		return err
	}

	go func() {
		wg.Wait()
		close(pages)
	}()

	return nil
}

func runWritePages(outDBPath string, cleaned <-chan db.MessageID, parallel int, errs chan<- error) (*sync.WaitGroup, error) {
	if outDBPath == "" {
		return printPages(cleaned), nil
	}

	outDB := db.NewRunner(outDBPath, parallel)

	return outDB.Write(cleaned, errs)
}

func cleanPages(cleaned chan<- db.MessageID) jobs.Page {
	return func(page *documents.Page) error {
		page.Text = nlp.CleanArticle(page.Text)
		cleaned <- page

		return nil
	}
}

func printPages(cleaned <-chan db.MessageID) *sync.WaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for p := range cleaned {
			page, ok := p.(*documents.Page)
			if !ok {
				panic(fmt.Sprintf("got message type %T, want %T", p, &documents.Page{}))
			}

			fmt.Println(page.Text)
		}

		wg.Done()
	}()

	return &wg
}

func toUint32Chan(ids []uint) chan uint32 {
	result := make(chan uint32, jobs.WorkBuffer)

	go func() {
		for _, id := range ids {
			result <- uint32(id)
		}

		close(result)
	}()

	return result
}
