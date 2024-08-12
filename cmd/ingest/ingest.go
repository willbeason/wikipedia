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

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
	"github.com/willbeason/wikipedia/pkg/workflows"
)

const namespaceKey = "namespace"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   "ingest workspace_path articles_path index_path",
		Short: `extract articles into a wikopticon workspace`,
		RunE:  runCmd,
	}

	cmd.Flags().IntSlice(namespaceKey, []int{int(documents.NamespaceArticle)},
		"the article namespaces to include")

	return cmd
}

var ErrExtract = errors.New("unable to run extraction")

var (
	MultistreamPattern      = regexp.MustCompile(`enwiki-(\d+)-pages-articles-multistream(\d*)\.xml-p(\d+)p(\d+)\.bz2`)
	MultistreamIndexPattern = regexp.MustCompile(`enwiki-(\d+)-pages-articles-multistream-index(\d*)\.txt-p(\d+)p(\d+)\.bz2`)
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

	// enwiki-20240801-pages-articles-multistream1.xml-p1p41242.bz2
	articlesPath := args[0]
	articlesPathMatches := MultistreamPattern.FindStringSubmatch(articlesPath)
	if len(articlesPathMatches) < 5 {
		return fmt.Errorf("%w: articles path %q does not match known pattern", ErrExtract, articlesPath)
	}
	enwikiDate := articlesPathMatches[1]
	enwikiShard := articlesPathMatches[2]

	// enwiki-20240801-pages-articles-multistream-index1.txt-p1p41242.bz2
	indexPath := args[1]
	indexPathMatches := MultistreamIndexPattern.FindStringSubmatch(indexPath)
	if len(indexPathMatches) < 5 {
		return fmt.Errorf("%w: index path %q does not match known pattern", ErrExtract, indexPath)
	}
	if indexPathMatches[1] != enwikiDate {
		return fmt.Errorf("%w: index date %q does not match articles date %q", ErrExtract, indexPath, enwikiDate)
	}
	if indexPathMatches[2] != enwikiShard {
		return fmt.Errorf("%w: index shard %q does not match articles date %q", ErrExtract, indexPath, enwikiDate)
	}

	if enwikiShard == "" {
		fmt.Printf("Extracting enwiki %q\n to workspace %q", enwikiDate, workspacePath)
	} else {
		fmt.Printf("Extracting enwiki %q shard %q to workspace %q\n", enwikiDate, enwikiShard, workspacePath)
	}

	var corpusName string
	if enwikiShard == "" {
		corpusName = fmt.Sprintf("%s", enwikiDate)
	} else {
		corpusName = fmt.Sprintf("%s.%s", enwikiDate, enwikiShard)
	}

	corpusPath := filepath.Join(workspacePath, corpusName)
	err = os.MkdirAll(corpusPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("%w: could not create corpus directory: %w", ErrExtract, err)
	}

	// Perform the extraction.
	extractPath := filepath.Join(corpusPath, config.ArticlesDir)
	err = flags.CreateOrCheckDirectory(extractPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtract, err)
	}

	runner := jobs.NewRunner()

	ctx, cancel := context.WithCancelCause(cmd.Context())

	compressedItems, err := source(cancel, articlesPath, indexPath)
	if err != nil {
		return err
	}

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	namespaces, err := cmd.Flags().GetIntSlice(namespaceKey)
	if err != nil {
		return fmt.Errorf("%w: reading namespace flag :%w", ErrExtract, err)
	}
	pages := extractPages(cancel, parallel, namespaces, compressedItems)

	outDB, err := badger.Open(badger.DefaultOptions(extractPath))
	if err != nil {
		return fmt.Errorf("opening %q: %w", extractPath, err)
	}
	defer func() {
		closeErr := outDB.Close()
		if err != nil {
			err = closeErr
		}
	}()

	sinkWork := jobs.Reduce(jobs.WorkBuffer, pages, db.WriteProto[protos.ID](outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)
	sinkWg.Wait()

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
		return err
	}

	err = r.RunCorpusWorkflow(cmd, corpusName, config.PostIngestWorkflow)
	if err != nil && !errors.Is(err, workflows.ErrWorkflowNotExist) {
		return err
	}

	return nil
}

func source(cancel context.CancelCauseFunc, repo, index string) (<-chan compressedDocument, error) {
	// Open the compressed data file.
	fRepo, err := os.Open(repo)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %w", repo, err)
	}

	// Open the index file.
	fIndex, err := os.Open(index)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %w", index, err)
	}

	// Create a channel of the compressed Wikipedia pages.
	compressedItems := make(chan compressedDocument, jobs.WorkBuffer)

	go func() {
		defer func() {
			err = fRepo.Close()
			if err != nil {
				cancel(err)
			}

			err = fIndex.Close()
			if err != nil {
				cancel(err)
			}
		}()

		var rIndex *bufio.Reader
		compressed := filepath.Ext(index) == ".bz2"
		if compressed {
			rIndex = bufio.NewReader(bzip2.NewReader(fIndex))
		} else {
			rIndex = bufio.NewReader(fIndex)
		}

		err2 := extractFile(rIndex, fRepo, compressedItems)
		if err2 != nil {
			cancel(err2)
		}
		close(compressedItems)
	}()

	return compressedItems, nil
}

func extractFile(articleIndex *bufio.Reader, fRepo *os.File, work chan<- compressedDocument) error {
	var startIndex, endIndex int64

	var (
		line     string
		err      error
		outBytes []byte
	)

	for ; err == nil; line, err = articleIndex.ReadString('\n') {
		if len(line) == 0 {
			continue
		}

		parts := strings.SplitN(line, ":", 3)

		endIndex, err = strconv.ParseInt(parts[0], 10, strconv.IntSize)
		if err != nil {
			return fmt.Errorf("parsing end index from %q: %w", line, err)
		}

		if startIndex == endIndex {
			continue
		}

		outBytes = make([]byte, endIndex-startIndex)

		_, err = fRepo.Read(outBytes)
		if err != nil {
			return fmt.Errorf("reading from %q: %w", fRepo.Name(), err)
		}

		work <- outBytes

		startIndex = endIndex
	}

	if errors.Is(err, io.EOF) {
		fmt.Println("got last file")

		outBytes, err = io.ReadAll(fRepo)
		if err != nil {
			return fmt.Errorf("reading from %q: %w", fRepo.Name(), err)
		}

		work <- outBytes
	}

	return nil
}

func extractPages(
	cancel context.CancelCauseFunc,
	parallel int,
	namespaces []int,
	compressedItems <-chan compressedDocument,
) <-chan protos.ID {
	pages := make(chan protos.ID, jobs.WorkBuffer)

	wg := sync.WaitGroup{}
	for range parallel {
		wg.Add(1)
		go func() {
			err := extractPagesWorker(namespaces, compressedItems, pages)
			if err != nil {
				cancel(err)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(pages)
	}()

	return pages
}

func extractPagesWorker(namespaces []int, compressed <-chan compressedDocument, pages chan<- protos.ID) error {
	allowedNamespaces := make(map[documents.Namespace]bool, len(namespaces))
	for _, ns := range namespaces {
		allowedNamespaces[documents.Namespace(ns)] = true
	}

	for j := range compressed {
		err := decompress(allowedNamespaces, j, pages)
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

func decompress(allowedNamespaces map[documents.Namespace]bool, compressed []byte, outPages chan<- protos.ID) error {
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
		if !allowedNamespaces[page.NS] || page.Redirect.Title != "" {
			// Ignore redirects and articles in other Namespaces.
			continue
		}

		pageProto := page.ToProto()

		outPages <- pageProto
	}

	return nil
}
