package extract

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
)

const namespaceKey = "namespace"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		Use:  `extract articles_path index_path allowed_namespace out_path`,
		Short: `Extracts the compressed pages-articles-multistream dump of Wikipedia to an output
Badger database, given an already-extracted index file.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)
	cmd.Flags().IntSlice(namespaceKey, []int{int(documents.NamespaceArticle)},
		"the article namespace to ")

	return cmd
}

var ErrExtract = errors.New("unable to run extraction")

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ns, err := strconv.ParseInt(args[2], 10, 16)
	if err != nil {
		return fmt.Errorf("%w: allowed_namespace argument must be an integer", ErrExtract)
	}

	extractCfg := &config.Extract{
		ArticlesPath: args[0],
		IndexPath:    args[1],
		Namespaces:   []int{int(ns)},
		OutPath:      args[3],
	}

	return Extract(cmd, extractCfg)
}

func Extract(cmd *cobra.Command, extract *config.Extract) error {
	articlesPath := extract.GetArticlesPath()
	if _, err := os.Stat(articlesPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: articles not found at %q", ErrExtract, articlesPath)
	}

	indexPath := extract.GetIndexPath()
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: index not found at %q", ErrExtract, indexPath)
	}

	outPath := extract.GetOutPath()
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		if err != nil {
			return fmt.Errorf("%w: unable to determine if output database already exists at %q: %w",
				ErrExtract, outPath, err)
		} else {
			return fmt.Errorf("%w: out directory exists: %q",
				ErrExtract, outPath)
		}
	} else {
		err = os.MkdirAll(outPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating output directory %q: %w", outPath, err)
		}
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

	pages := extractPages(cancel, parallel, extract.Namespaces, compressedItems)

	outDB, err := badger.Open(badger.DefaultOptions(outPath))
	if err != nil {
		return fmt.Errorf("opening %q: %w", outPath, err)
	}

	defer func() {
		closeErr := outDB.Close()
		if closeErr != nil {
			cancel(closeErr)
		}
	}()

	sinkWork := jobs.Reduce(jobs.WorkBuffer, pages, db.WriteProto(outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)
	sinkWg.Wait()

	err = db.RunGC(outDB)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return ctx.Err()
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
