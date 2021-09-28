package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/willbeason/wikipedia/pkg/flags"
)

var version = "dirty"

func main() {
	ctx := context.Background()

	err := mainCmd().ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Version: version,
		Use:     `wikoptica subcommand`,
		Short:   `Run wikoptica analysis on a Wikipedia corpus`,
		RunE:    runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	fmt.Println("hello")

	return nil
}
