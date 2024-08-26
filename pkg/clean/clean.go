package clean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/article"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/db"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `clean corpus_name articles_path out_path`,
		Short: `Clean Wikipedia articles`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

var ErrClean = errors.New("unable to run article cleaning")

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	cfg := &config.Clean{
		In:  args[1],
		Out: args[2],
	}

	return Clean(cmd, cfg, args[0])
}

func Clean(cmd *cobra.Command, cfg *config.Clean, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrClean, corpusNames)
	}
	corpusName := corpusNames[0]

	articlesDir := cfg.In
	outDir := cfg.Out
	fmt.Printf("Cleaning corpus %q directory %q and writing to %q\n",
		corpusName, articlesDir, outDir)

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	articlesDir = filepath.Join(workspace, corpusName, articlesDir)
	outDir = filepath.Join(workspace, corpusName, outDir)

	if _, err = os.Stat(articlesDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: articles not found at %q", ErrClean, articlesDir)
	}

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())

	source := pages.StreamDB[documents.Page](articlesDir, parallel)

	docs, err := source(ctx, cancel)
	if err != nil {
		return fmt.Errorf("reading articles for cleaning: %w", err)
	}

	cleanedChannel, cleanWork := jobs.MapOld(jobs.WorkBuffer, docs, func(from *documents.Page) (*documents.Page, error) {
		tokens := article.Tokenize(article.UnparsedText(from.Text))
		from.Text = article.Render(tokens)
		return from, nil
	})

	var sinkWork jobs.WorkQueue
	var outDB *badger.DB

	outDB, err = toOutDB(outDir)
	if err != nil {
		return err
	}

	sinkWork = jobs.Reduce(ctx, jobs.WorkBuffer, cleanedChannel, db.WriteProto[*documents.Page](outDB))

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
