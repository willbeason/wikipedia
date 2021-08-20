package main

import (
	"os"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

func mainCmd() *cobra.Command {
	return &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inArticles := args[0]
			inDictionary := args[1]
			outDictionary := args[2]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			dictionary, err := nlp.ReadDictionary(inDictionary)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			work := jobs.WalkFiles(inArticles, errs)

			results := make(chan string)

			workWg := jobs.DoPageJobs(parallel, doWork(results), work, errs)

			var titleDictionary map[string]bool
			titleWg := sync.WaitGroup{}
			titleWg.Add(1)

			go func() {
				titleDictionary = collectTitleDictionary(nlp.ToNgramDictionary(dictionary), results)
				titleWg.Done()
			}()

			workWg.Wait()
			close(results)
			close(errs)
			errsWg.Wait()

			titleWg.Wait()

			return nlp.WriteDictionary(outDictionary, titleDictionary)
		},
	}
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func doWork(results chan<- string) jobs.Page {
	return func(page *documents.Page) error {
		title := page.Title
		if !nlp.IsArticle(title) {
			return nil
		}

		for _, word := range nlp.WordRegex.FindAllString(title, -1) {
			if !nlp.HasLetter(word) {
				continue
			}

			word = nlp.Normalize(word)
			results <- word
		}

		return nil
	}
}

func collectTitleDictionary(dictionary map[string]bool, words <-chan string) map[string]bool {
	frequencies := documents.FrequencyMap{
		Counts: make(map[string]int),
	}
	frequencies.Collect(words)

	result := make(map[string]bool)

	for word, count := range frequencies.Counts {
		if !dictionary[word] && count >= 2 {
			result[word] = true
		}
	}

	return result
}
