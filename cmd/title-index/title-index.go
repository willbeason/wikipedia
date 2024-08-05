package title_index

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `title-index path/to/input`,
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
		ArticlesPath: args[0],
		OutPath:      args[1],
	}

	return TitleIndex(cmd, cfg)
}

func TitleIndex(cmd *cobra.Command, cfg *config.TitleIndex) error {
	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	source := pages.StreamDB(cfg.GetArticlesPath(), parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())
	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	results := makeIndex(ps)

	index := <-results

	err = protos.Write(cfg.GetOutPath(), index)
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
