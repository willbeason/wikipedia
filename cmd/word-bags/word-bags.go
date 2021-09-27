package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/ordinality"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	inDB := args[0]
	inDictionary := args[1]
	outDBPath := args[2]

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	dictionary, err := nlp.ReadDictionary(inDictionary)
	if err != nil {
		return err
	}

	source := pages.StreamDB(inDB, parallel)

	outDB := db.NewRunner(outDBPath, parallel)
	sink := outDB.Write()

	converter := ordinality.WordBagConverter{
		Tokenizer: nlp.NgramTokenizer{
			Underlying: nlp.WordTokenizer{},
			Dictionary: nlp.ToNgramDictionary(dictionary),
		},
		WordOrdinality: ordinality.NewWordOrdinality(dictionary),
	}

	ctx := cmd.Context()
	return pages.Run(ctx, source, parallel, documentWordBags(converter), sink)
}

func documentWordBags(converter ordinality.WordBagConverter) func(chan<- protos.ID) jobs.Page {
	return func(out chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			wordBag := converter.ToPageWordBag(page)
			out <- wordBag

			return nil
		}
	}
}
