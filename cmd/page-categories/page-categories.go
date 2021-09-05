package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
	"os"
	"sort"
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
		Args: cobra.ExactArgs(3),
		Use:  `view path/to/input id`,
		Short: `View a specific article by its identifier`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	ctx := cmd.Context()

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := args[0]
	inTitleIndex := args[1]
	//outPageCategories := args[2]

	titleIndex := &documents.TitleIndex{}
	err = protos.Read(inTitleIndex, titleIndex)
	if err != nil {
		return err
	}

	source := pages.StreamDB(inDB, parallel)

	errs, errsWg := jobs.Errors()

	ps, err := source(ctx, errs)

	results := makePageCategories(titleIndex, ps)

	_ = <- results
	//pageCategories := <- results

	//err = protos.Write(outPageCategories, pageCategories)
	//if err != nil {
	//	errs <- err
	//}

	close(errs)
	errsWg.Wait()

	return nil
}

func makePageCategories(titleIndex *documents.TitleIndex, pages <-chan *documents.Page) <-chan *documents.PageCategories {
	results := make(chan *documents.PageCategories)

	go func() {
		categorizer := &documents.Categorizer{
			TitleIndex: titleIndex,
		}

		result := &documents.PageCategories{
			Pages: make(map[uint32]*documents.Categories),
		}

		for page := range pages {
			result.Pages[page.Id] = categorizer.Categorize(page)
		}

		fmt.Println(documents.Missed)

		freqs := make([]freq, len(documents.MissedMap))
		idx := 0

		for k, v := range documents.MissedMap {
			freqs[idx] = freq{c: v, s: k}
			idx++
		}
		sort.Slice(freqs, func(i, j int) bool {
			return freqs[i].c > freqs[j].c
		})

		for _, f := range freqs {
			fmt.Println(f.s, ":", f.c)
		}


		results <- result
	}()

	return results
}

type freq struct {
	c int
	s string
}
