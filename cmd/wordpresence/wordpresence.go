package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
	"gopkg.in/yaml.v3"
)

var cmd = cobra.Command{
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		inArticles := args[0]
		inWordOrder := args[1]
		out := args[2]

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

		work := make(chan string)
		errs := make(chan error)

		go func() {
			err := filepath.WalkDir(inArticles, walker.Files(work))
			if err != nil {
				errs <- err
			}
			close(work)
		}()

		results := make(chan documents.WordSets)
		workWg := sync.WaitGroup{}
		for i := 0; i < 8; i++ {
			workWg.Add(1)
			go func() {
				for item := range work {
					err := doWork(dictionary, item, results)
					if err != nil {
						errs <- fmt.Errorf("%s: %w", item, err)
					}
				}
				workWg.Done()
			}()
		}

		errsWg := sync.WaitGroup{}
		errsWg.Add(1)
		go func() {
			for err := range errs {
				fmt.Println(err)
			}
			errsWg.Done()
		}()

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

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func doWork(dictionary map[string]int, path string, results chan<- documents.WordSets) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	bytes = []byte(strings.ReplaceAll(string(bytes), "\t", ""))

	var doc documents.Document
	err = yaml.Unmarshal(bytes, &doc)
	if err != nil {
		return err
	}

	docWords := make([]documents.WordSet, len(doc.Pages))
	for i := range doc.Pages {
		title := doc.Pages[i].Title
		if !nlp.IsArticle(title) {
			continue
		}

		seen := getPresences(dictionary, doc.Pages[i].Revision.Text)
		docWords[i] = documents.WordSet{
			ID:    doc.Pages[i].ID,
			Words: seen,
		}
	}

	results <- documents.WordSets{
		InFile:    path,
		Documents: docWords,
	}

	return nil
}

func getPresences(dictionary map[string]int, text string) []uint16 {
	seen := make(map[string]bool)

	var result []uint16
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
