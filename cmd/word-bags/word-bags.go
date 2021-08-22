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
	"github.com/willbeason/extract-wikipedia/pkg/ordinality"
	"google.golang.org/protobuf/proto"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inDBPath := args[0]
			inDictionary := args[1]
			outDBPath := args[2]

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

			dictionary, err := nlp.ReadDictionary(inDictionary)
			if err != nil {
				return err
			}

			converter := ordinality.WordBagConverter{
				Tokenizer: nlp.NgramTokenizer{
					Underlying: nlp.WordTokenizer{},
					Dictionary: nlp.ToNgramDictionary(dictionary),
				},
				WordOrdinality: ordinality.NewWordOrdinality(dictionary),
			}

			errs, errsWg := jobs.Errors()

			work := jobs.Walk(cmd.Context(), inDB, newPage, parallel, errs)
			wordBags := make(chan jobs.MessageID, 100)

			workWg := jobs.RunProto(parallel, documentWordBags(converter, wordBags), work, errs)

			writeWg := jobs.WriteProtos(outDB, parallel, wordBags, errs)

			workWg.Wait()
			close(wordBags)

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

func documentWordBags(converter ordinality.WordBagConverter, out chan<- jobs.MessageID) jobs.Proto {
	return func(p proto.Message) error {
		page, ok := p.(*documents.Page)
		if !ok {
			return fmt.Errorf("got message type %T, want %T", p, &documents.Page{})
		}

		wordBag := converter.ToPageWordBag(page)
		out <- wordBag

		return nil
	}
}
