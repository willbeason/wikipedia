package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"encoding/xml"
	"fmt"
	"github.com/willbeason/wikipedia/pkg/environment"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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
		Use: `extract-wikipedia path/to/pages-articles-multistream.xml.bz2 \
  path/to/pages-articles-multistream-index.txt \
  path/to/output.db`,
		Short: `Extracts the compressed pages-articles-multistream dump of Wikipedia to an output
Badger database, given an already-extracted index file. Excludes all redirect pages.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)
	cmd.Flags().Int16(namespaceKey, int16(documents.NamespaceArticle),
		"The Namespace of pages to keep. Defaults to Articles.")

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	ns, err := cmd.Flags().GetInt16(namespaceKey)
	if err != nil {
		return err
	}

	repo := ""
	if len(args) > 0 {
		repo = args[0]
	} else {
		repo = filepath.Join(environment.WikiPath, "dump", "pages-articles-multistream.xml.bz2")
	}

	index := ""
	if len(args) > 1 {
		index = args[1]
	} else {
		index = filepath.Join(environment.WikiPath, "dump", "pages-articles-multistream-index.txt")
	}

	outDBPath := ""
	if len(args) > 2 {
		outDBPath = args[2]
	} else {
		outDBPath = filepath.Join(environment.WikiPath, "extracted.db")
	}

	outDB := db.NewRunner(outDBPath, parallel)
	sink := outDB.Write()

	errs, errsWg := jobs.Errors()
	compressedItems, err := source(repo, index, errs)
	if err != nil {
		return err
	}

	pages := extractPages(parallel, documents.Namespace(ns), compressedItems, errs)

	outWg, err := sink(cmd.Context(), pages, errs)
	if err != nil {
		return err
	}

	outWg.Wait()
	close(errs)
	errsWg.Wait()

	return nil
}

func source(repo, index string, errs chan<- error) (<-chan compressedDocument, error) {
	fRepo, err := os.Open(repo)
	if err != nil {
		return nil, err
	}

	fIndex, err := os.Open(index)
	if err != nil {
		return nil, err
	}

	compressedItems := make(chan compressedDocument, jobs.WorkBuffer)
	rIndex := bufio.NewReader(fIndex)

	go func() {
		defer func() {
			err = fRepo.Close()
			if err != nil {
				errs <- err
			}

			err = fIndex.Close()
			if err != nil {
				errs <- err
			}
		}()

		extractFile(rIndex, fRepo, compressedItems, errs)
		close(compressedItems)
	}()

	return compressedItems, nil
}

func extractFile(rIndex *bufio.Reader, fRepo *os.File, work chan<- compressedDocument, errs chan<- error) {
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
			errs <- err
			return
		}

		if startIndex == endIndex {
			continue
		}

		outBytes = make([]byte, endIndex-startIndex)

		_, err = fRepo.Read(outBytes)
		if err != nil {
			errs <- err
			return
		}

		work <- outBytes

		startIndex = endIndex

		if err == io.EOF {
			break
		}
	}

	if err == io.EOF {
		fmt.Println("got last file")

		outBytes, err = io.ReadAll(fRepo)
		if err != nil {
			errs <- err
			return
		}
		work <- outBytes
	}
}

func extractPages(parallel int, ns documents.Namespace, compressedItems <-chan compressedDocument, errs chan<- error) <-chan protos.ID {
	pages := make(chan protos.ID, jobs.WorkBuffer)

	wg := sync.WaitGroup{}
	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			extractPagesWorker(ns, compressedItems, pages, errs)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(pages)
	}()

	return pages
}

func extractPagesWorker(ns documents.Namespace, compressed <-chan compressedDocument, pages chan<- protos.ID, errs chan<- error) {
	for j := range compressed {
		decompress(ns, j, pages, errs)
	}
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

// decompress extracts pages of the passed Namespace which are not redirects.
func decompress(ns documents.Namespace, compressed []byte, outPages chan<- protos.ID, errs chan<- error) {
	bz := bzip2.NewReader(bytes.NewReader(compressed))

	compressed, err := io.ReadAll(bz)
	if err != nil {
		errs <- err
		return
	}

	text := normalize(string(compressed))

	doc := &documents.XMLDocument{}

	err = xml.Unmarshal([]byte(text), doc)
	if err != nil {
		errs <- err
		return
	}

	for _, page := range doc.Pages {
		if page.NS != ns || page.Redirect.Title != "" {
			// Ignore redirects and articles in other Namespaces.
			continue
		}

		outPages <- page.ToProto()
	}

	if err != nil {
		errs <- err
	}
}
