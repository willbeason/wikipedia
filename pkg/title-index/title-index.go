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
	"github.com/willbeason/wikipedia/pkg/pages"
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

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	articlesDir = filepath.Join(workspace, corpusName, articlesDir)
	outFile = filepath.Join(workspace, corpusName, outFile)

	source := pages.StreamDB[documents.Page](articlesDir, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())
	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	results := makeIndex(ps)

	index := <-results

	err = protos.Write(outFile, index)
	if err != nil {
		return fmt.Errorf("%w: writing title index: %w", ErrTitleIndex, err)
	}

	return nil
}

func makeIndex(pages <-chan *documents.Page) <-chan *documents.TitleIndex {
	results := make(chan *documents.TitleIndex)

	go func() {
		result := &documents.TitleIndex{
			Titles: make(map[string]uint32),
		}

		for page := range pages {
			result.Titles[page.Title] = page.Id
		}

		results <- result
		close(results)
	}()

	return results
}
