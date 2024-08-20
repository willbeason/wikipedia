package ingest

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/willbeason/wikipedia/pkg/protos"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/workflows"
)

const namespaceKey = "namespace"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   "ingest articles_path index_path",
		Short: `extract articles into a wikopticon workspace`,
		RunE:  runCmd,
	}

	cmd.Flags().IntSlice(namespaceKey, []int{int(documents.NamespaceArticle)},
		"the article namespaces to include")

	return cmd
}

var ErrExtract = errors.New("unable to run extraction")

var (
	EnwikiPrefix            = `enwiki`
	MultistreamPattern      = regexp.MustCompile(`enwiki-(\d+)-pages-articles-multistream(\d*)\.xml(?:-p(\d+)p(\d+))?\.bz2`)
	MultistreamIndexPattern = regexp.MustCompile(`enwiki-(\d+)-pages-articles-multistream-index(\d*)\.txt(?:-p(\d+)p(\d+))?\.bz2`)

	WikidataPrefix = `wikidata`
	// WikidataPattern matches wikidata-20240701-all.json.bz2.
	WikidataPattern = regexp.MustCompile(`wikidata-(\d+)-all\.json\.bz2`)
)

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	workspacePath, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(workspacePath) {
		workingDirectory, err2 := os.Getwd()
		if err2 != nil {
			return fmt.Errorf("could not determine working directory: %w", err2)
		}

		workspacePath = filepath.Join(workingDirectory, workspacePath)
	}

	cfg, err := config.Load(workspacePath)
	if err != nil {
		return err
	}
	r := workflows.Runner{Config: cfg}

	return ingestEnwiki(cmd, args, workspacePath, r)
}

func ingestEnwiki(cmd *cobra.Command, args []string, workspacePath string, r workflows.Runner) error {
	// enwiki-20240701-pages-articles-multistream.xml.bz2
	// enwiki-20240801-pages-articles-multistream1.xml-p1p41242.bz2
	articlesPath := args[0]
	articlesPathMatches := MultistreamPattern.FindStringSubmatch(articlesPath)
	if len(articlesPathMatches) < 2 {
		return fmt.Errorf("%w: articles path %q does not match known pattern", ErrExtract, articlesPath)
	}
	enwikiDate := articlesPathMatches[1]
	var enwikiShard string
	if len(articlesPathMatches) > 2 {
		enwikiShard = articlesPathMatches[2]
	}

	// enwiki-20240701-pages-articles-multistream-index.txt.bz2
	// enwiki-20240801-pages-articles-multistream-index1.txt-p1p41242.bz2
	indexPath := args[1]
	indexPathMatches := MultistreamIndexPattern.FindStringSubmatch(indexPath)
	if len(indexPathMatches) < 2 {
		return fmt.Errorf("%w: index path %q does not match known pattern", ErrExtract, indexPath)
	}
	if indexPathMatches[1] != enwikiDate {
		return fmt.Errorf("%w: index date %q does not match articles date %q", ErrExtract, indexPathMatches[1], enwikiDate)
	}
	if len(indexPathMatches) > 2 {
		if indexPathMatches[2] != enwikiShard {
			return fmt.Errorf("%w: index shard %q does not match articles shard %q", ErrExtract, indexPathMatches[2], enwikiDate)
		}
	}

	if enwikiShard == "" {
		fmt.Printf("Extracting enwiki %q to workspace %q\n", enwikiDate, workspacePath)
	} else {
		fmt.Printf("Extracting enwiki %q shard %q to workspace %q\n", enwikiDate, enwikiShard, workspacePath)
	}

	var corpusName string
	if enwikiShard == "" {
		corpusName = enwikiDate
	} else {
		corpusName = fmt.Sprintf("%s.%s", enwikiDate, enwikiShard)
	}

	corpusPath := filepath.Join(workspacePath, corpusName)
	err := os.MkdirAll(corpusPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("%w: could not create corpus directory: %w", ErrExtract, err)
	}

	// Perform the extraction.
	extractPath := filepath.Join(corpusPath, config.ArticlesDir)
	err = flags.CreateOrCheckDirectory(extractPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtract, err)
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())

	compressedItems := source(ctx, cancel, articlesPath, indexPath)

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	namespaces, err := cmd.Flags().GetIntSlice(namespaceKey)
	if err != nil {
		return fmt.Errorf("%w: reading namespace flag :%w", ErrExtract, err)
	}
	pages, redirects := extractPages(cancel, parallel, namespaces, compressedItems)

	outDB, err := badger.Open(
		badger.DefaultOptions(extractPath).
			WithMetricsEnabled(false).
			WithLoggingLevel(badger.WARNING),
	)
	if err != nil {
		return fmt.Errorf("opening %q: %w", extractPath, err)
	}
	defer func() {
		closeErr := outDB.Close()
		if err != nil {
			err = closeErr
		}
	}()

	runner := jobs.NewRunner()
	sinkWork := jobs.Reduce(ctx, jobs.WorkBuffer*100, pages, db.WriteProto[*documents.Page](outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)

	redirectsWg := sync.WaitGroup{}
	redirectsWg.Add(1)
	go func() {
		redirectsProto := &documents.Redirects{
			Redirects: make(map[string]string),
		}
		for redirect := range redirects {
			redirectsProto.Redirects[redirect.GetTitle()] = redirect.GetRedirect()
		}

		redirectsPath := filepath.Join(corpusPath, "redirects.txt")
		protoErr := protos.Write(redirectsPath, redirectsProto)
		if protoErr != nil {
			cancel(err)
		}

		redirectsWg.Done()
	}()

	sinkWg.Wait()
	redirectsWg.Wait()

	err = db.RunGC(outDB)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	// Close the DB before reading.
	err = outDB.Close()
	if err != nil {
		return fmt.Errorf("closing ingested articles database: %w", err)
	}

	err = r.RunWorkflow(cmd, config.PostIngestWorkflow, corpusName)
	if err != nil && !errors.Is(err, workflows.ErrWorkflowNotExist) {
		return err
	}

	return nil
}

func source(ctx context.Context, cancel context.CancelCauseFunc, repo, index string) <-chan compressedDocument {
	// Decouple file operations: reading indices from the index and extracting articles.
	endIndices := extractEndIndices(ctx, cancel, index)
	compressedItems := extractArticles(ctx, cancel, repo, endIndices)

	return compressedItems
}

func extractEndIndices(ctx context.Context, cancel context.CancelCauseFunc, index string) <-chan int64 {
	endIndices := make(chan int64, jobs.WorkBuffer*100)

	go func() {
		defer close(endIndices)

		// Open the index file.
		fIndex, err := os.Open(index)
		if err != nil {
			cancel(fmt.Errorf("%w: opening %q: %w", ErrExtract, index, err))
			return
		}

		defer func() {
			closeErr := fIndex.Close()
			if closeErr != nil {
				fmt.Println(fmt.Errorf("%w: closing %q: %w", ErrExtract, index, closeErr))
			}
		}()

		var rIndex *bufio.Reader
		switch filepath.Ext(index) {
		case ".bz2":
			rIndex = bufio.NewReader(bzip2.NewReader(fIndex))
		case ".txt":
			rIndex = bufio.NewReader(fIndex)
		default:
			cancel(fmt.Errorf("%w: unrecognized index extension: %q", ErrExtract, index))
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, readErr := rIndex.ReadString('\n')
				switch {
				case errors.Is(readErr, io.EOF):
					fmt.Println("got last article")
					return
				case readErr != nil:
					cancel(readErr)
					return
				case len(line) == 0:
					continue
				default:
					parts := strings.SplitN(line, ":", 3)
					if len(parts) != 3 {
						cancel(fmt.Errorf("%w: invalid index line %q", ErrExtract, line))
						return
					}

					endIndex, parseErr := strconv.ParseInt(parts[0], 10, strconv.IntSize)
					if parseErr != nil {
						cancel(fmt.Errorf("parsing end index from %q: %w", line, parseErr))
						return
					}

					endIndices <- endIndex
				}
			}
		}
	}()

	return endIndices
}

func extractArticles(ctx context.Context, cancel context.CancelCauseFunc, repo string, endIndices <-chan int64) <-chan compressedDocument {
	// Create a channel of the compressed Wikipedia pages.
	work := make(chan compressedDocument, jobs.WorkBuffer*100)

	go func() {
		defer close(work)

		// Open the compressed data file.
		fRepo, err := os.Open(repo)
		if err != nil {
			cancel(fmt.Errorf("opening %q: %w", repo, err))
			return
		}

		defer func() {
			err = fRepo.Close()
			if err != nil {
				cancel(err)
			}
		}()

		var startIndex, endIndex int64
		var outBytes []byte

		for endIndex = range endIndices {
			select {
			case <-ctx.Done():
				return
			default:
				if startIndex == endIndex {
					continue
				}

				outBytes = make([]byte, endIndex-startIndex)

				_, err = fRepo.Read(outBytes)
				if err != nil {
					cancel(fmt.Errorf("reading from %q: %w", fRepo.Name(), err))
					return
				}

				work <- outBytes

				startIndex = endIndex
			}
		}

		// Read the last document since it isn't marked in the index.
		outBytes, err = io.ReadAll(fRepo)
		if err != nil {
			cancel(fmt.Errorf("reading from %q: %w", fRepo.Name(), err))
			return
		}

		work <- outBytes
	}()

	return work
}

func extractPages(
	cancel context.CancelCauseFunc,
	parallel int,
	namespaces []int,
	compressedItems <-chan compressedDocument,
) (<-chan *documents.Page, <-chan *documents.Redirect) {
	pages := make(chan *documents.Page, jobs.WorkBuffer*100)
	redirects := make(chan *documents.Redirect, jobs.WorkBuffer*100)

	wg := sync.WaitGroup{}
	for range parallel {
		wg.Add(1)
		go func() {
			err := extractPagesWorker(namespaces, compressedItems, redirects, pages)
			if err != nil {
				cancel(err)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(pages)
		close(redirects)
	}()

	return pages, redirects
}

func extractPagesWorker(namespaces []int, compressed <-chan compressedDocument, redirects chan<- *documents.Redirect, pages chan<- *documents.Page) error {
	allowedNamespaces := make(map[documents.Namespace]bool, len(namespaces))
	for _, ns := range namespaces {
		allowedNamespaces[documents.Namespace(ns)] = true
	}

	for j := range compressed {
		err := decompress(allowedNamespaces, j, redirects, pages)
		if err != nil {
			return err
		}
	}

	return nil
}

type compressedDocument []byte

// normalizeXML ensures all passed XML strings begin and end with mediawiki tags.
func normalizeXML(text string) string {
	if !strings.HasPrefix(text, "<mediawiki") {
		text = "<mediawiki>\n" + text
	}

	if !strings.HasSuffix(text, "</mediawiki>") {
		text += "\n</mediawiki>"
	}

	return text
}

func decompressBz2(compressed []byte) ([]byte, error) {
	bz := bzip2.NewReader(bytes.NewReader(compressed))

	uncompressed, err := io.ReadAll(bz)
	if err != nil {
		return nil, fmt.Errorf("reading from compressed reader: %w", err)
	}

	return uncompressed, nil
}

func decompress(allowedNamespaces map[documents.Namespace]bool, compressed []byte, outRedirects chan<- *documents.Redirect, outPages chan<- *documents.Page) error {
	uncompressed, err := decompressBz2(compressed)
	if err != nil {
		return err
	}

	text := normalizeXML(string(uncompressed))

	doc := &documents.XMLDocument{}

	err = xml.Unmarshal([]byte(text), doc)
	if err != nil {
		return fmt.Errorf("unmarshalling text %q: %w", text, err)
	}

	for _, page := range doc.Pages {
		switch {
		case !allowedNamespaces[page.NS]:
			continue
		case page.Redirect.Title != "":
			outRedirects <- &documents.Redirect{
				Title:    page.Title,
				Redirect: page.Redirect.Title,
			}
		default:
			pageProto := page.ToProto()
			outPages <- pageProto
		}
	}

	return nil
}
