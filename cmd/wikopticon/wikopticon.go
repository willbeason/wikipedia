package main

import (
	"errors"
	"fmt"
	"os"


	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/cmd/clean"
	"github.com/willbeason/wikipedia/cmd/extract"
	title_index "github.com/willbeason/wikipedia/cmd/title-index"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/flags"
)

const Version = "0.3.0"

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

	cmd.AddCommand(extract.Cmd())
	cmd.AddCommand(clean.Cmd())
	cmd.AddCommand(title_index.Cmd())

	flags.Parallel(cmd)

	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:    cobra.ExactArgs(2),
		Use:     `run config_yaml job_name`,
		Short:   `runs a specific wikopticon job in a configuration file`,
		RunE:    runRunE,
		Version: Version,
	}

	return cmd
}

func runRunE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	configPath := args[0]
	jobName := args[1]

	c, err := config.Load(configPath)
	if err != nil {
		return err
	}

	jobConfig, err := c.GetJob(jobName)
	if err != nil {
		return err
	}

	switch cfg := jobConfig.(type) {
	case *config.Extract:
		fmt.Printf("Extracting %q with index %q to directory %q only namespaces %v\n",
			cfg.GetArticlesPath(), cfg.GetIndexPath(), cfg.GetOutPath(), cfg.Namespaces)
		return extract.Extract(cmd, cfg)
	case *config.Clean:
		fmt.Printf("Cleaning %q to directory %q viewing %v\n",
			cfg.GetArticlesPath(), cfg.GetOutPath(), cfg.View)
		return clean.Clean(cmd, cfg)
	case *config.TitleIndex:
		fmt.Printf("Creating title index of %q to %q\n",
			cfg.GetArticlesPath(), cfg.GetOutPath())
		return title_index.TitleIndex(cmd, cfg)
	default:
		return fmt.Errorf("%w: %T",
			ErrUnknownSubcommandType, jobConfig)
	}
}
