package main

import (
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			dictionarySize, err := cmd.Flags().GetInt(flags.DictionarySizeKey)
			if err != nil {
				return err
			}

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			inArticles := args[0]
			inDictionaryFile1 := args[1]
			inDictionaryFile2 := args[2]
			out := args[3]

			dictionary1, err := nlp.ReadDictionary(inDictionaryFile1)
			if err != nil {
				return err
			}

			var dictionary2 map[string]bool
			if inDictionaryFile2 != "" {
				dictionary2, err = nlp.ReadDictionary(inDictionaryFile2)
				if err != nil {
					return err
				}
			}

			errs, errsWg := jobs.Errors()

			work := jobs.WalkDir(inArticles, errs)

			results := make(chan map[string]int)

			workWg := jobs.DoPageJobs(parallel, doWork(results), work, errs)

			wordCountsWg := sync.WaitGroup{}
			var wordCounts map[string]int

			wordCountsWg.Add(1)
			go func() {
				wordCounts = collect(dictionary1, dictionary2, results)
				wordCountsWg.Done()
			}()

			workWg.Wait()
			close(results)
			close(errs)
			errsWg.Wait()

			wordCountsWg.Wait()

			frequencies := documents.ToFrequencyTable(wordCounts)
			frequencies.Frequencies = frequencies.Frequencies[:dictionarySize]

			return documents.WriteFrequencyTable(out, frequencies)
		},
	}

	flags.Parallel(cmd)
	flags.DictionarySize(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func doWork(results chan<- map[string]int) jobs.Page {
	return func(page *documents.Page) error {
		counts := countWords(page)
		results <- counts

		return nil
	}
}

func countWords(p *documents.Page) map[string]int {
	result := make(map[string]int)

	words := nlp.WordRegex.FindAllString(p.Revision.Text, -1)
	for _, word := range words {
		result[word]++
	}

	return result
}

func collect(dictionary1, dictionary2 map[string]bool, results <-chan map[string]int) map[string]int {
	result := make(map[string]int)

	for counts := range results {
		for word, v := range counts {
			word = nlp.Normalize(word)
			if word == "" {
				continue
			}

			if !dictionary1[word] && !dictionary2[word] {
				// Not in either dictionary.
				continue
			}

			result[word] += v
		}
	}

	return result
}
