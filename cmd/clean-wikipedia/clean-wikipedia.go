package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/clean"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.
func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.RangeArgs(1, 2),
		Use:  `clean-wikipedia path/to/input path/to/output`,
		Short: `Cleans an extracted set of Wikipedia articles by removing irrelevant pages and formatting
directives.`,
		RunE: runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	inDBPath := args[0]

	var outDBPath string
	if len(args) > 1 {
		outDBPath = args[1]
	}

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	pageIDs, err := cmd.Flags().GetUintSlice(flags.IDsKey)
	if err != nil {
		return err
	}

	cmd.SilenceUsage = true
	ctx := cmd.Context()

	return clean.Run(ctx, inDBPath, parallel, pageIDs, outDBPath)
}
