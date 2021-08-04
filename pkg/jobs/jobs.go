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
//
// The returned channel must be closed before the WaitGroup is waited upon.
func Errors() (chan<- error, *sync.WaitGroup) {
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)

	errs := make(chan error)
	printedErr := SyncBool{}

	go func() {
		for err := range errs {
			printedErr.bool = true

			fmt.Println(err)
		}

		errsWg.Done()
	}()

	return errs, &errsWg
}

type Job func(page *documents.Page) error

// DoJobs run parallel workers performing job on work and filling errs with any
// encountered errors.
//
// Returns a WaitGroup which waits for all workers to finish.
func DoJobs(parallel int, job Job, work <-chan string, errs chan<- error) *sync.WaitGroup {
	workWg := sync.WaitGroup{}

	for i := 0; i < parallel; i++ {
		workWg.Add(1)

		go func() {
			runWorker(job, work, errs)
			workWg.Done()
		}()
	}

	return &workWg
}

func runWorker(job Job, work <-chan string, errs chan<- error) {
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
