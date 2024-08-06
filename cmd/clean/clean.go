package clean

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/pages"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `clean articles_path out_path`,
		Short: `Cleans Wikipedia articles.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

var ErrClean = errors.New("unable to run article cleaning")

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	cleanCfg := &config.Clean{
		ArticlesPath: args[0],
		OutPath:      args[1],
	}

	return Clean(cmd, cleanCfg)
}

func Clean(cmd *cobra.Command, clean *config.Clean) error {
	articlesPath := clean.GetArticlesPath()
	if _, err := os.Stat(articlesPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: articles not found at %q", ErrClean, articlesPath)
	}

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())

	var source pages.Source
	var compare func(page *documents.Page) error

	if clean.View == nil {
		source = pages.StreamDB(articlesPath, parallel)
	} else {
		beforeSource := pages.StreamDBKeys(articlesPath, parallel, clean.View)
		beforePages, err2 := beforeSource(ctx, cancel)
		if err2 != nil {
			return fmt.Errorf("getting articles before cleaning: %w", err2)
		}
		compare = pages.Compare(beforePages)

		source = pages.StreamDBKeys(articlesPath, parallel, clean.View)
	}

	docs, err := source(ctx, cancel)
	if err != nil {
		return fmt.Errorf("reading articles for cleaning: %w", err)
	}

	cleanedChannel, cleanWork := jobs.Map(jobs.WorkBuffer, docs, func(from *documents.Page) (*documents.Page, error) {
		from.Text = nlp.CleanArticle(from.Text)
		return from, nil
	})

	var sinkWork jobs.WorkQueue
	var outDB *badger.DB
	if clean.OutPath == "" {
		sinkWork = jobs.ForEach(jobs.WorkBuffer, cleanedChannel, compare)
	} else {
		outPath := clean.GetOutPath()
		outDB, err = toOutDB(outPath)
		if err != nil {
			return err
		}

		sinkWork = jobs.Reduce(jobs.WorkBuffer, cleanedChannel, db.WriteProto[*documents.Page](outDB))
	}

	runner := jobs.NewRunner()
	cleanWg := runner.Run(ctx, cancel, cleanWork)

	sinkWg := runner.Run(ctx, cancel, sinkWork)

	cleanWg.Wait()
	close(cleanedChannel)
	sinkWg.Wait()

	if outDB != nil {
		err = db.RunGC(outDB)
		if err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

func toOutDB(outPath string) (*badger.DB, error) {
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		if err != nil {
			return nil, fmt.Errorf("%w: unable to determine if output database already exists at %q: %w",
				ErrClean, outPath, err)
		} else {
			return nil, fmt.Errorf("%w: out directory exists: %q",
				ErrClean, outPath)
		}
	} else {
		err = os.MkdirAll(outPath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("creating output directory %q: %w", outPath, err)
		}
	}

	outDB, err := badger.Open(badger.DefaultOptions(outPath))
	if err != nil {
		return nil, fmt.Errorf("opening output DB: %q: %w", outPath, err)
	}
	return outDB, nil
}
