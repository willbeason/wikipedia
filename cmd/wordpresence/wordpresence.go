package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"gopkg.in/yaml.v3"
)

// GuessSize is a reasonable guess for the number of unique words in the average
// article, to reduce reallocations from expanding slices.
const GuessSize = 1000

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inArticles := args[0]
			inWordOrder := args[1]
			out := args[2]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			bytes, err := ioutil.ReadFile(inWordOrder)
			if err != nil {
				return err
			}

			frequencies := documents.FrequencyTable{}
			err = yaml.Unmarshal(bytes, &frequencies)
			if err != nil {
				return err
			}

			dictionary := make(map[string]int)
			for i, f := range frequencies.Frequencies {
				dictionary[f.Word] = i
			}

			errs, errsWg := jobs.Errors()
			work := jobs.WalkDir(inArticles, errs)

			results := make(chan documents.WordSets)
			workWg := jobs.DoDocumentJobs(parallel, doWork(dictionary, results), work, errs)

			writeResultsWg := sync.WaitGroup{}
			writeResultsWg.Add(1)
			go func() {
				writeResults(inArticles, out, results, errs)
				writeResultsWg.Done()
			}()

			workWg.Wait()
			close(results)

			writeResultsWg.Wait()

			close(errs)
			errsWg.Wait()

			return nil
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

func doWork(dictionary map[string]int, results chan<- documents.WordSets) jobs.Document {
	return func(doc *documents.Document) error {
		docWords := make([]documents.WordSet, len(doc.Pages))

		for i := range doc.Pages {
			title := doc.Pages[i].Title
			if !nlp.IsArticle(title) {
				continue
			}

			seen := getPresences(GuessSize, dictionary, doc.Pages[i].Revision.Text)
			docWords[i] = documents.WordSet{
				ID:    doc.Pages[i].ID,
				Words: seen,
			}
		}

		results <- documents.WordSets{
			InFile:    doc.Path,
			Documents: docWords,
		}

		return nil
	}
}

func getPresences(guessSize int, dictionary map[string]int, text string) []uint16 {
	seen := make(map[string]bool)

	result := make([]uint16, 0, guessSize)

	for _, word := range nlp.WordRegex.FindAllString(text, -1) {
		word = nlp.Normalize(word)
		if seen[word] {
			continue
		}

		order, found := dictionary[word]
		if !found {
			continue
		}

		seen[word] = true

		result = append(result, uint16(order))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func writeResults(in, out string, results <-chan documents.WordSets, errs chan<- error) {
	for result := range results {
		outFile := filepath.Join(out, strings.TrimPrefix(result.InFile, in))
		outFile = strings.TrimSuffix(outFile, ".txt") + ".json"
		result.InFile = ""

		bytes, err := json.Marshal(result)
		if err != nil {
			errs <- err
			continue
		}

		err = os.MkdirAll(filepath.Dir(outFile), os.ModePerm)
		if err != nil {
			errs <- err
			continue
		}

		err = ioutil.WriteFile(outFile, bytes, os.ModePerm)
		if err != nil {
			errs <- err
		}
	}
}
