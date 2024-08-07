package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
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
		Use:   `first-links`,
		Short: `Analyzes the network of references between biographical articles.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	inDB := filepath.Join(environment.WikiPath, "extracted.db")
	outDBPath := filepath.Join(environment.WikiPath, "cleaned.db")
	outDB, err := badger.Open(badger.DefaultOptions(outDBPath))
	if err != nil {
		return fmt.Errorf("opening output DB: %q: %w", outDBPath, err)
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())

	source := pages.StreamDB(inDB, parallel)
	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	cleanedChannel, cleanWork := jobs.Map(jobs.WorkBuffer, docs, func(from *documents.Page) (protos.ID, error) {
		from.Text = nlp.CleanArticle(from.Text)
		return from, nil
	})

	runner := jobs.NewRunner()
	cleanWg := runner.Run(ctx, cancel, cleanWork)

	sinkWork := jobs.Reduce(jobs.WorkBuffer, cleanedChannel, db.WriteProto(outDB))
	sinkWg := runner.Run(ctx, cancel, sinkWork)

	cleanWg.Wait()
	sinkWg.Wait()

	err = db.RunGC(outDB)
	if err != nil {
		return err
	}

	return ctx.Err()
}
