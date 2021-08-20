package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/indexes"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"google.golang.org/protobuf/proto"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inDocuments := args[0]
			outIndex := args[1]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			work := jobs.WalkFiles(inDocuments, errs)

			entries := make(chan *indexes.Entry, 100)

			workWg := jobs.DoDocumentJobs(parallel, indexExtracted(entries), work, errs)
			indexChan := collectEntries(inDocuments, entries)

			workWg.Wait()
			close(entries)

			close(errs)
			errsWg.Wait()

			index := <- indexChan

			return writeIndex(outIndex, index)
		},
	}

	flags.Parallel(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func indexExtracted(entries chan<- *indexes.Entry) jobs.Document {
	return func(doc *documents.Document) error {
		nPages := len(doc.Pages)
		if nPages == 0 {
			return nil
		}

		entry := &indexes.Entry{
			File: doc.Path,
			Max:  uint32(doc.Pages[nPages-1].ID),
		}

		entries <- entry

		return nil
	}
}

func collectEntries(root string, entries <-chan *indexes.Entry) <-chan *indexes.Index {
	result := make(chan *indexes.Index)

	go func() {
		index := &indexes.Index{
			Root: root,
		}

		for entry := range entries {
			file := entry.File
			if strings.HasPrefix(entry.File, root) {
				var err error
				file, err = filepath.Rel(root, entry.File)
				if err != nil {
					panic(err)
				}
			}

			entry.File = file
			index.Entries = append(index.Entries, entry)
		}

		sort.Slice(index.Entries, func(i, j int) bool {
			return index.Entries[i].Max < index.Entries[j].Max
		})

		result <- index

		close(result)
	}()

	return result
}

func writeIndex(out string, index *indexes.Index) error {
	bytes, err := proto.Marshal(index)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(out, bytes, os.ModePerm)
}
