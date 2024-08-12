package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/cmd/ingest"
	"github.com/willbeason/wikipedia/pkg/clean"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/title-index"
	"github.com/willbeason/wikipedia/pkg/workflows"
)

const (
	Version = "0.3.0"
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	ErrNoSubcommand          = errors.New("missing subcommand")
	ErrUnknownSubcommandType = errors.New("unknown subcommand")
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `wikopticon subcommand`,
		Short: `Runs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%w: specify a subcommand", ErrNoSubcommand)
		},
		Version: Version,
	}

	cmd.AddCommand(runCmd())

	cmd.AddCommand(ingest.Cmd())
	cmd.AddCommand(clean.Cmd())
	cmd.AddCommand(title_index.Cmd())

	flags.Parallel(cmd)
	flags.Workspace(cmd)

	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:    cobra.RangeArgs(1, 2),
		Use:     `run corpus_name job_name`,
		Short:   `runs a specific wikopticon job in a corpus`,
		RunE:    runRunE,
		Version: Version,
	}

	return cmd
}

func runRunE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	workspacePath, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(workspacePath)
	if err != nil {
		return err
	}
	r := workflows.Runner{Config: cfg}

	toRun := args[0]
	var corpusName string
	if len(args) > 1 {
		corpusName = args[1]
	}

	if corpusName != "" {
		err = r.RunCorpusWorkflow(cmd, corpusName, toRun)
	}
	if !errors.Is(err, workflows.ErrWorkflowNotExist) {
		return err
	} else if err == nil {
		// Successfully ran a workflow.
		return nil
	}

	return r.RunCorpusJob(cmd, corpusName, toRun)
}
