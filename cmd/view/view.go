package view

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"google.golang.org/protobuf/encoding/protojson"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `view articles_path articles...`,
		Short: `View specific articles by identifier (--ids) or title (--titles)`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)
	flags.Titles(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx, cancel := context.WithCancelCause(cmd.Context())

	// We don't want to print pages in parallel.
	parallel := 1

	pageIDs, err := flags.GetIDs(cmd)
	if err != nil {
		return err
	}

	titles, err := flags.GetTitles(cmd)
	if err != nil {
		return err
	}
	if len(titles) > 0 {
		indexFilepath := filepath.Join(environment.WikiPath, environment.TitleIndex)
		indexBytes, err2 := os.ReadFile(indexFilepath)
		if err2 != nil {
			return fmt.Errorf("reading index %q: %w", indexFilepath, err2)
		}

		index := documents.TitleIndex{}
		err2 = protojson.Unmarshal(indexBytes, &index)
		if err2 != nil {
			return fmt.Errorf("unmarshalling index %q: %w", indexFilepath, err2)
		}

		for _, title := range titles {
			id, found := index.Titles[title]
			if !found {
				return flags.InvalidFlagError(flags.TitlesKey, fmt.Sprintf("unable to find article of title %q", title))
			}
			pageIDs = append(pageIDs, uint(id))
		}
	}

	if len(pageIDs) == 0 {
		return fmt.Errorf("%w: must specify %q or %q flag", flags.ErrInvalidFlag, flags.IDsKey, flags.TitlesKey)
	}

	var inDB string
	if len(args) > 0 {
		inDB = args[0]
	} else {
		inDB = filepath.Join(environment.WikiPath, "extracted.db")
	}

	if _, err = os.Stat(inDB); err != nil {
		return fmt.Errorf("unable to open %q: %w", inDB, err)
	}

	var source func(ctx context.Context, cancel context.CancelCauseFunc) (<-chan *documents.Page, error)
	if len(pageIDs) == 0 {
		source = pages.StreamDB(inDB, parallel)
	} else {
		fmt.Println("Page IDs", pageIDs)
		source = pages.StreamDBKeys(inDB, parallel, pageIDs)
	}

	runner := jobs.NewRunner()

	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	printWork := jobs.ForEach(jobs.WorkBuffer, ps, pages.Print)

	printWg := runner.Run(ctx, cancel, printWork)
	printWg.Wait()

	return nil
}
