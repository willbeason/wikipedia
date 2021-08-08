package jobs

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"gopkg.in/yaml.v3"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
)

// WalkDir walks inArticles, returning a channel of files to process.
func WalkDir(inArticles string, errs chan<- error) <-chan string {
	work := make(chan string)

	if filepath.Ext(inArticles) != "" {
		// We were passed a single file, so there is only one work item.
		go func() {
			work <- inArticles
			close(work)
		}()
	} else {
		// We were passed a directory, so walk it recursively.
		go func() {
			err := filepath.WalkDir(inArticles, walker.Files(work))
			if err != nil {
				errs <- err
			}

			close(work)
		}()
	}

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

type Page func(page *documents.Page) error

type Document func(doc *documents.Document) error

// DoDocumentJobs runs parallel workers performing job on work and filling errs with any encountered
// errors. Jobs perform actions on Documents.
//
// Returns a WaitGroup which waits for all workers to finish.
func DoDocumentJobs(parallel int, job Document, work <-chan string, errs chan<- error) *sync.WaitGroup {
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

// DoPageJobs run parallel workers performing job on work and filling errs with any
// encountered errors. Jobs perform actions on Pages.
//
// Returns a WaitGroup which waits for all workers to finish.
func DoPageJobs(parallel int, job Page, work <-chan string, errs chan<- error) *sync.WaitGroup {
	return DoDocumentJobs(parallel, func(doc *documents.Document) error {
		for i := range doc.Pages {
			page := &doc.Pages[i]
			if !nlp.IsArticle(page.Title) {
				continue
			}

			// page.Revision.Text = nlp.NormalizeArticle(page.Revision.Text)

			err := job(&doc.Pages[i])
			if err != nil {
				errs <- err
			}
		}

		return nil
	}, work, errs)
}

func runWorker(job Document, work <-chan string, errs chan<- error) {
	for path := range work {
		doc, err := readDocument(path)
		if err != nil {
			errs <- err
			continue
		}

		err = job(doc)
		if err != nil {
			errs <- err
		}
	}
}

// ErrUnsupportedExtension indicates that the file is not a supported format for Documents.
var ErrUnsupportedExtension = errors.New("unsupported Document extension")

func readDocument(path string) (*documents.Document, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc := &documents.Document{}

	switch ext := filepath.Ext(path); ext {
	case ".xml":
		err = xml.Unmarshal(bytes, doc)
	case ".yaml":
		err = yaml.Unmarshal(bytes, doc)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedExtension, ext)
	}

	if err != nil {
		return nil, err
	}

	doc.Path = path

	return doc, nil
}
