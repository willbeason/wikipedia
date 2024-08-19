package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/cmd/ingest"
	"github.com/willbeason/wikipedia/pkg/analysis"
	"github.com/willbeason/wikipedia/pkg/clean"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/ingest-wikidata"
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
	cmd.AddCommand(analysis.RenamedArticlesCmd())
	cmd.AddCommand(ingest_wikidata.Cmd())

	flags.Parallel(cmd)
	flags.Workspace(cmd)

	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:    cobra.RangeArgs(1, 3),
		Use:     `run job_name corpus_name...`,
		Short:   `runs a specific wikopticon job in a corpus`,
		RunE:    runRunE,
		Version: Version,
	}

	cmd.Flags().String("cpuprofile", "", "write cpu profile to file")

	return cmd
}

func runRunE(cmd *cobra.Command, args []string) error {
	cpuprofile, _ := cmd.Flags().GetString("cpuprofile")
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			return err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}

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
	corpusNames := args[1:]

	err = r.RunWorkflow(cmd, toRun, corpusNames...)
	if !errors.Is(err, workflows.ErrWorkflowNotExist) {
		return err
	} else if err == nil {
		// Successfully ran a workflow.
		return nil
	}

	return r.RunJob(cmd, toRun, corpusNames...)
}
