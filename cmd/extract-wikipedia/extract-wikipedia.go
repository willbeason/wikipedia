package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

const namespaceKey = "namespace"

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		Use: `extract-wikipedia path/to/pages-articles-multistream.xml.bz2 \
  path/to/pages-articles-multistream-index.txt \
  path/to/output.db`,
		Short: `Extracts the compressed pages-articles-multistream dump of Wikipedia to an output
Badger database, given an already-extracted index file.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)
	cmd.Flags().Int16(namespaceKey, int16(documents.NamespaceArticle), "")

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	ns, err := cmd.Flags().GetInt16(namespaceKey)
	if err != nil {
		return flags.ParsingFlagError(namespaceKey, err)
	}

	repo := args[0]
	index := args[1]

	outDBPath := args[2]

	runner := jobs.NewRunner()

	ctx, cancel := context.WithCancelCause(cmd.Context())

	compressedItems, err := source(cancel, repo, index)
	if err != nil {
		return err
	}

	pages := extractPages(cancel, parallel, documents.Namespace(ns), compressedItems)

	outDB, err := badger.Open(badger.DefaultOptions(outDBPath))
	if err != nil {
		return fmt.Errorf("opening %q: %w", outDBPath, err)
	}

	defer func() {
		err2 := outDB.Close()
		if err2 != nil {
			cancel(err2)
		}
	}()

	sinkWork := jobs.Reduce(jobs.WorkBuffer, pages, db.WriteProto(outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)
	sinkWg.Wait()

	err = db.RunGC(outDB)
	if err != nil {
		return err
	}

	return ctx.Err()
}

func source(cancel context.CancelCauseFunc, repo, index string) (<-chan compressedDocument, error) {
	// Open the compressed data file.
	fRepo, err := os.Open(repo)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %w", repo, err)
	}

	// Open the uncompressed index file.
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

		rIndex := bufio.NewReader(fIndex)
		err2 := extractFile(rIndex, fRepo, compressedItems)
		if err2 != nil {
			cancel(err2)
		}
		close(compressedItems)
	}()

	return compressedItems, nil
}

func extractFile(rIndex *bufio.Reader, fRepo *os.File, work chan<- compressedDocument) error {
	var startIndex, endIndex int64

	var (
		lineBytes []byte
		err       error
		outBytes  []byte
	)

	for ; err == nil; lineBytes, _, err = rIndex.ReadLine() {
		if len(lineBytes) == 0 {
			continue
		}

		line := string(lineBytes)
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

	if err == io.EOF {
		fmt.Println("got last file")

		outBytes, err = io.ReadAll(fRepo)
		if err != nil {
			return fmt.Errorf("reading from %q: %w", fRepo.Name(), err)
		}

		work <- outBytes
	}

	return nil
}

func extractPages(cancel context.CancelCauseFunc, parallel int, ns documents.Namespace, compressedItems <-chan compressedDocument) <-chan protos.ID {
	pages := make(chan protos.ID, jobs.WorkBuffer)

	wg := sync.WaitGroup{}
	for range parallel {
		wg.Add(1)
		go func() {
			err := extractPagesWorker(ns, compressedItems, pages)
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

func extractPagesWorker(ns documents.Namespace, compressed <-chan compressedDocument, pages chan<- protos.ID) error {
	for j := range compressed {
		err := decompress(ns, j, pages)
		if err != nil {
			return err
		}
	}

	return nil
}

type compressedDocument []byte

func normalize(text string) string {
	if !strings.HasPrefix(text, "<mediawiki") {
		text = "<mediawiki>\n" + text
	}

	if !strings.HasSuffix(text, "</mediawiki>") {
		text += "\n</mediawiki>\n"
	}

	return text
}

func decompress(ns documents.Namespace, compressed []byte, outPages chan<- protos.ID) error {
	bz := bzip2.NewReader(bytes.NewReader(compressed))

	compressed, err := io.ReadAll(bz)
	if err != nil {
		return fmt.Errorf("reading from compressed reader: %w", err)
	}

	text := normalize(string(compressed))

	doc := &documents.XMLDocument{}

	err = xml.Unmarshal([]byte(text), doc)
	if err != nil {
		return fmt.Errorf("unmarshalling text %q: %w", text, err)
	}

	// infoboxChecker, err := documents.NewInfoboxChecker(documents.PersonInfoboxes)

	for _, page := range doc.Pages {
		if page.NS != ns || page.Redirect.Title != "" {
			// Ignore redirects and articles in other Namespaces.
			continue
		}

		pageProto := page.ToProto()

		//// Exclude non-biographies.
		//if !infoboxChecker.Matches(pageProto.Text) {
		//	continue
		//}

		outPages <- pageProto
	}

	return nil
}
