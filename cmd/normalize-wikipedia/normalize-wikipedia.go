package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.MinimumNArgs(2),
		Use:  `normalize-wikipedia path/to/input path/to/output`,
		Short: `Normalizes text in Wikipedia by making all text lowercase and replacing certain sequences
(e.g. numbers, dates) with normalized tokens.
Mainly for use in early stages of corpus analysis.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	pageIDs, err := cmd.Flags().GetUintSlice(flags.IDsKey)
	if err != nil {
		return err
	}

	inDB := args[0]

	var outDBPath string
	var sink protos.Sink

	if len(args) > 1 {
		outDBPath = args[1]
		outDB := db.NewRunner(outDBPath, parallel)
		sink = outDB.Write()
	} else {
		sink = protos.PrintProtos
	}

	var source pages.Source
	if len(pageIDs) == 0 {
		source = pages.StreamDB(inDB, parallel)
	} else {
		source = pages.StreamDBKeys(inDB, parallel, pageIDs)
	}

	cmd.SilenceUsage = true
	ctx := cmd.Context()

	return pages.Run(ctx, source, parallel, normalize, sink)
}

func normalize(out chan<- protos.ID) jobs.Page {
	return func(page *documents.Page) error {
		page.Text = nlp.NormalizeArticle(page.Text)
		out <- page

		return nil
	}
}
