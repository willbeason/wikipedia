package jobs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"gopkg.in/yaml.v3"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
)

// WalkDir walks inArticles, returning a channel of files to process.
func WalkDir(inArticles string, errs chan<- error) <-chan string {
	work := make(chan string)

	go func() {
		err := filepath.WalkDir(inArticles, walker.Files(work))
		if err != nil {
			errs <- err
		}

		close(work)
	}()

	return work
}

type SyncBool struct {
	bool
	mtx sync.Mutex
}

func (b *SyncBool) Get() bool {
	var result bool

	b.mtx.Lock()
	result = b.bool
	b.mtx.Unlock()

	return result
}

// Errors creates an input for errors to print and a bool of whether any
// errors have been printed.
//
// Calls Done on the WaitGroup once the channel has been closed and all errors
// have been printed.
func Errors(wg *sync.WaitGroup) (chan<- error, *SyncBool) {
	errs := make(chan error)
	printedErr := SyncBool{}

	go func() {
		for err := range errs {
			printedErr.bool = true

			fmt.Println(err)
		}

		wg.Done()
	}()

	return errs, &printedErr
}

type Job func(page *documents.Page) error

func DoJobs(job Job, workWg *sync.WaitGroup, work <-chan string, errs chan<- error) {
	go func() {
		for path := range work {
			doc, err := readDocument(path)
			if err != nil {
				errs <- err
				continue
			}

			for i := range doc.Pages {
				page := &doc.Pages[i]
				if !nlp.IsArticle(page.Title) {
					continue
				}

				page.Revision.Text = nlp.NormalizeArticle(page.Revision.Text)

				err = job(&doc.Pages[i])
				if err != nil {
					errs <- err
					continue
				}
			}
		}

		workWg.Done()
	}()
}

func readDocument(path string) (*documents.Document, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bytes = []byte(strings.ReplaceAll(string(bytes), "\t", ""))

	doc := &documents.Document{}

	err = yaml.Unmarshal(bytes, doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

var ErrEncountered = errors.New("encountered error")
