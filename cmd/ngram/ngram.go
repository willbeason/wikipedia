package main

import (
	"os"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/jobs"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

const (
	DefaultMinCount = 1000

	DefaultCountFilter   = 20
	DefaultSizeThreshold = 2e7
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inArticles := args[0]
			inDictionaries := args[1]
			outCounts := args[2]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			frequencyTable := &documents.FrequencyTable{}
			if inDictionaries != "" {
				inDictionariesList := strings.Split(inDictionaries, ",")

				frequencyTable, err = documents.ReadFrequencyTables(inDictionariesList...)
				if err != nil {
					return err
				}
			}

			errsWg := sync.WaitGroup{}
			errsWg.Add(1)
			errs, _ := jobs.Errors(&errsWg)

			workWg := sync.WaitGroup{}
			work := jobs.WalkDir(inArticles, errs)

			tokenizer := nlp.NgramTokenizer{
				Underlying: nlp.WordTokenizer{},
				Dictionary: frequencyTable.ToNgramDictionary(),
			}

			results := make(chan map[string]int)
			for i := 0; i < parallel; i++ {
				workWg.Add(1)
				jobs.DoJobs(getNgrams(tokenizer, results), &workWg, work, errs)
			}

			counts := documents.FrequencyMap{
				Counts: make(map[string]int),
			}
			countsWg := sync.WaitGroup{}
			countsWg.Add(1)
			go func() {
				counts.CollectMaps(results, DefaultCountFilter, DefaultSizeThreshold)
				countsWg.Done()
			}()

			workWg.Wait()
			close(results)
			close(errs)

			errsWg.Wait()

			countsWg.Wait()

			counts.Filter(DefaultMinCount)

			frequencies := documents.ToFrequencyTable(counts.Counts)

			return documents.WriteFrequencyTable(outCounts, frequencies)
		},
	}

	flags.Parallel(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getNgrams(tokenizer nlp.Tokenizer, results chan<- map[string]int) jobs.Job {
	return func(page *documents.Page) error {
		frequencies := documents.FrequencyMap{
			Counts: make(map[string]int),
		}

		text := page.Revision.Text
		text = strings.ToLower(text)

		ngrams := tokenizer.Tokenize(text)

		for j := 1; j < len(ngrams); j++ {
			ngram := ngrams[j-1] + " " + ngrams[j]

			if strings.Count(ngram, " ") < 1 {
				continue
			}

			frequencies.Counts[ngram]++
		}

		results <- frequencies.Counts

		return nil
	}
}
