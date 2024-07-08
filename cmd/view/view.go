package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"path/filepath"
)

func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `view path/to/input`,
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

	pageIDs, err := cmd.Flags().GetUintSlice(flags.IDsKey)
	if err != nil {
		return err
	}

	titles, err := cmd.Flags().GetStringSlice(flags.TitlesKey)
	if err != nil {
		return err
	}
	if len(titles) > 0 {
		indexFilepath := filepath.Join(environment.WikiPath, environment.TitleIndex)
		indexBytes, err2 := os.ReadFile(indexFilepath)
		if err2 != nil {
			return err2
		}

		index := documents.TitleIndex{}
		err2 = protojson.Unmarshal(indexBytes, &index)
		if err2 != nil {
			return err2
		}

		for _, title := range titles {
			id, found := index.Titles[title]
			if !found {
				return fmt.Errorf("unable to find article of title %q", title)
			}
			pageIDs = append(pageIDs, uint(id))
		}
	}

	if len(pageIDs) == 0 {
		return fmt.Errorf("must specify at least one ID or title")
	}

	inDB := ""
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

	printWork := jobs.ForEach(jobs.WorkBuffer, ps, func(from *documents.Page) error {
		fmt.Println(from.Id)
		fmt.Println(from.Title)
		fmt.Println(from.Text)
		return nil
	})

	printWg := runner.Run(ctx, cancel, printWork)
	printWg.Wait()

	return nil
}
