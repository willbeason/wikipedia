package title_index

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `title-index corpus_name articles_dir out_file`,
		Short: `Create an index from titles to article IDs`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

var ErrTitleIndex = errors.New("unable to create title index")

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	cfg := &config.TitleIndex{
		In:  args[1],
		Out: args[2],
	}

	return TitleIndex(cmd, cfg, args[0])
}

func TitleIndex(cmd *cobra.Command, cfg *config.TitleIndex, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrTitleIndex, corpusNames)
	}
	corpusName := corpusNames[0]

	articlesDir := cfg.In
	outFile := cfg.Out
	fmt.Printf("Creating title index for corpus %q from directory %q and writing to %q\n",
		corpusName, articlesDir, outFile)

	ctx := cmd.Context()
	ctx, cancel := context.WithCancelCause(ctx)

	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	articlesDir = filepath.Join(workspace, corpusName, articlesDir)
	outFile = filepath.Join(workspace, corpusName, outFile)

	pageSource := jobs.NewSource(protos.ReadDir[documents.Page](articlesDir))
	pageSourceWg, pageSourceJob, ps := pageSource()
	go pageSourceJob(ctx, errs)

	indexMap := jobs.NewMap(jobs.Convert(getTitle))
	indexMapWg, indexMapJob, titles := indexMap(ps)
	go indexMapJob(ctx, errs)

	titlesSink := jobs.NewSink(protos.WriteFile[*documents.ArticleIdTitle](outFile))
	titlesSinkWg, titlesSinkJob := titlesSink(titles)
	go titlesSinkJob(ctx, errs)

	pageSourceWg.Wait()
	indexMapWg.Wait()
	titlesSinkWg.Wait()

	return nil
}

func getTitle(page *documents.Page) (*documents.ArticleIdTitle, error) {
	return &documents.ArticleIdTitle{
		Id:    page.Id,
		Title: page.Title,
	}, nil
}
