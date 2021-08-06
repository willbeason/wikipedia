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
)

func Parallel(cmd *cobra.Command) {
	var p int

	cmd.Flags().IntVar(&p, "parallel", runtime.NumCPU(),
		"number of concurrent workers to run on jobs; defaults to number of available logical CPUs")
}

func DictionarySize(cmd *cobra.Command) {
	var p int

	cmd.Flags().IntVar(&p, DictionarySizeKey, DictionarySizeDefault,
		"maximum number of top words to keep track of")
}
