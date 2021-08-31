package main

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
	"google.golang.org/protobuf/proto"
)

const (
	idsKey = "ids"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.MinimumNArgs(2),
		Use:  `normalize-wikipedia path/to/input path/to/output`,
		Short: `Normalizes text in Wikipedia by making all text lowercase and replacing certain sequences
(e.g. numbers, dates) with normalized tokens.
Mainly for use in early stages of corpus analysis.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inDBPath := args[0]
			outDBPath := args[1]

			inDB, err := badger.Open(badger.DefaultOptions(inDBPath))
			defer func() {
				_ = inDB.Close()
			}()
			if err != nil {
				return err
			}

			outDB, err := badger.Open(badger.DefaultOptions(outDBPath))
			defer func() {
				_ = outDB.Close()
			}()
			if err != nil {
				return err
			}

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			pageIds, err := cmd.Flags().GetUintSlice(idsKey)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			var work <-chan proto.Message

			if len(pageIds) == 0 {
				work = jobs.Walk(cmd.Context(), inDB, newPage, parallel, errs)
			} else {
				work = jobs.IDs(inDB, newPage, pageIds, errs)
			}

			normalized := make(chan protos.ID, 100)
			normalizeWg := jobs.RunProto(parallel, normalize(normalized), work, errs)

			writeWg := jobs.WriteProtos(outDB, parallel, normalized, errs)

			normalizeWg.Wait()
			close(normalized)

			writeWg.Wait()
			close(errs)

			errsWg.Wait()

			err = inDB.Close()
			if err != nil {
				return err
			}

			err = outDB.Close()
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags.Parallel(cmd)

	cmd.Flags().UintSlice(idsKey, nil, "A list of specific article ids to check.")

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func newPage() proto.Message {
	return &documents.Page{}
}

func normalize(out chan<- protos.ID) jobs.Proto {
	return func(p proto.Message) error {
		page, ok := p.(*documents.Page)
		if !ok {
			return fmt.Errorf("got message type %T, want %T", p, &documents.Page{})
		}

		page.Text = nlp.NormalizeArticle(page.Text)
		out <- page

		return nil
	}
}
