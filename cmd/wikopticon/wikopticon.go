package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
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

	return cmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:    cobra.ExactArgs(2),
		Use:     `run path/to/config.yaml job`,
		Short:   `Runs a specific job in a configuration file.`,
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
		fmt.Printf("extracting %q with index %q only namespaces %v\n",
			cfg.ArticlesPath, cfg.IndexPath, cfg.Namespaces)
	default:
		return fmt.Errorf("%w: %T",
			ErrUnknownSubcommandType, jobConfig)
	}

	return nil
}
