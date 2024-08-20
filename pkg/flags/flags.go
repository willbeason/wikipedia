package flags

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Flags and default values for keys used in the various CLIs.
const (
	ParallelKey = "parallel"

	DictionarySizeKey     = "dictionary-size"
	DictionarySizeDefault = 50000

	IDsKey    = "ids"
	TitlesKey = "titles"

	WorkspaceKey           = "workspace"
	WikopticonWorkspaceEnv = "WIKOPTICON_WORKSPACE"
)

var (
	ErrInvalidFlag = errors.New("invalid flag")
	ErrMissingFlag = errors.New("missing flag")
)

func InvalidFlagError(flag string, message string) error {
	return fmt.Errorf("%w %q: %s", ErrInvalidFlag, flag, message)
}

func ParsingFlagError(flag string, err error) error {
	return fmt.Errorf("parsing flag %q: %w", flag, err)
}

func Parallel(cmd *cobra.Command) {
	cmd.PersistentFlags().Int("parallel", runtime.NumCPU(),
		"number of concurrent workers to run on jobs; defaults to number of available logical CPUs")
}

func GetParallel(cmd *cobra.Command) (int, error) {
	parallel, err := cmd.Flags().GetInt(ParallelKey)
	if err != nil {
		return 0, ParsingFlagError(ParallelKey, err)
	}

	return parallel, nil
}

func DictionarySize(cmd *cobra.Command) {
	cmd.Flags().Int(DictionarySizeKey, DictionarySizeDefault,
		"maximum number of top words to keep track of")
}

func IDs(cmd *cobra.Command) {
	cmd.PersistentFlags().UintSlice(IDsKey, nil, "A list of specific article ids to check.")
}

func GetIDs(cmd *cobra.Command) ([]uint, error) {
	ids, err := cmd.Flags().GetUintSlice(IDsKey)
	if err != nil {
		return nil, ParsingFlagError(IDsKey, err)
	}

	return ids, nil
}

func Titles(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSlice(TitlesKey, nil, "A list of specific article titles to check.")
}

func GetTitles(cmd *cobra.Command) ([]string, error) {
	titles, err := cmd.Flags().GetStringSlice(TitlesKey)
	if err != nil {
		return nil, ParsingFlagError(TitlesKey, err)
	}

	return titles, nil
}

func Workspace(cmd *cobra.Command) {
	cmd.PersistentFlags().String(WorkspaceKey, os.Getenv(WikopticonWorkspaceEnv),
		"the workspace to use")
}

func GetWorkspacePath(cmd *cobra.Command) (string, error) {
	workspace, err := cmd.Flags().GetString(WorkspaceKey)
	if err != nil {
		return "", ParsingFlagError(WorkspaceKey, err)
	}

	if workspace == "" {
		return "", ParsingFlagError(WorkspaceKey, fmt.Errorf("%w: \"--%s\" flag not specified", ErrMissingFlag, WorkspaceKey))
	}

	return workspace, nil
}
