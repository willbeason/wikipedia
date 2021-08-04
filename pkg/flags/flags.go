package flags

import "github.com/spf13/cobra"

const (
	ParallelKey     = "parallel"
	ParallelDefault = 8

	DictionarySizeKey     = "dictionary-size"
	DictionarySizeDefault = 50000
)

func Parallel(cmd *cobra.Command) {
	var p int

	cmd.Flags().IntVar(&p, "parallel", ParallelDefault,
		"number of concurrent workers to run on jobs")
}

func DictionarySize(cmd *cobra.Command) {
	var p int

	cmd.Flags().IntVar(&p, DictionarySizeKey, DictionarySizeDefault,
		"maximum number of top words to keep track of")
}
