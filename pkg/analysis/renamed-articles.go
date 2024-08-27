package analysis

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
)

func RenamedArticlesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(4),
		Use:   `renamed-articles corpus_before title_before corpus_after title_after`,
		Short: `count the renamed articles from the title_index of two corpora`,
		RunE:  runAnalysisCmd,
	}

	return cmd
}

func runAnalysisCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	return RenamedArticles(cmd, args[0], args[1], args[2], args[3])
}

type RenamedArticlesResult struct {
	Before  int
	After   int
	Renamed int
	Deleted int
}

// RenamedArticles counts the number of articles whose titles changed from before to after.
func RenamedArticles(cmd *cobra.Command, corpusBefore, titlesBefore, corpusAfter, titlesAfter string) error {
	workspacePath, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	beforePath := filepath.Join(workspacePath, corpusBefore, titlesBefore)
	afterPath := filepath.Join(workspacePath, corpusAfter, titlesAfter)

	ctx := cmd.Context()
	ctx, cancel := context.WithCancelCause(ctx)

	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	beforeFuture := documents.ReadTitleMap(ctx, beforePath, errs)
	afterFuture := documents.ReadTitleMap(ctx, afterPath, errs)

	before := <- beforeFuture
	after := <- afterFuture

	beforeSize := len(before)
	afterSize := len(after)

	inverseAfter := make(map[uint32]string)
	for title, id := range after {
		inverseAfter[id] = title
	}

	renamed := 0
	deleted := 0
	for beforeTitle, id := range before {
		if afterTitle, ok := inverseAfter[id]; ok && beforeTitle != afterTitle {
			renamed++
		} else if !ok {
			deleted++
		}
	}

	printArticlesChange(beforeSize, afterSize, deleted, renamed)

	return nil
}

func printArticlesChange(before, after, deleted, renamed int) {
	addedPercent := float64(after-before) / float64(after)

	deletedPercent := float64(deleted) / float64(before)
	kept := before - deleted
	renamedPercent := float64(renamed) / float64(kept)

	fmt.Printf("Articles:\n")
	fmt.Printf("  before:  %7d\n", before)
	fmt.Printf("  after:   %7d (%+6.02f%%)\n", after, 100*addedPercent)
	fmt.Printf("  deleted: %7d (%6.02f%%)\n", deleted, 100*deletedPercent)
	fmt.Printf("  renamed: %7d (%6.02f%%)\n", renamed, 100*renamedPercent)
}
