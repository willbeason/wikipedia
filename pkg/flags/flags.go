package flags

import (
	"runtime"

	"github.com/spf13/cobra"
)

// Flags and default values for keys used in the various CLIs.
const (
	ParallelKey = "parallel"

	DictionarySizeKey     = "dictionary-size"
	DictionarySizeDefault = 50000

	IDsKey = "ids"
)

func Parallel(cmd *cobra.Command) {
	cmd.PersistentFlags().Int("parallel", runtime.NumCPU(),
		"number of concurrent workers to run on jobs; defaults to number of available logical CPUs")
}

func DictionarySize(cmd *cobra.Command) {
	cmd.Flags().Int(DictionarySizeKey, DictionarySizeDefault,
		"maximum number of top words to keep track of")
}

func IDs(cmd *cobra.Command) {
	cmd.PersistentFlags().UintSlice(IDsKey, nil, "A list of specific article ids to check.")
}
