package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/walker"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var cmd = cobra.Command{
	Args: cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		in := args[0]
		inDictionaryFile1 := args[1]
		inDictionaryFile2 := args[2]
		out := args[3]

		dictionary1, err := nlp.ReadDictionary(inDictionaryFile1)
		if err != nil {
			return err
		}
		dictionary2, err := nlp.ReadDictionary(inDictionaryFile2)
		if err != nil {
			return err
		}

		work := make(chan string)
		errs := make(chan error)

		go func() {
			err := filepath.WalkDir(in, walker.Files(work))
			if err != nil {
				errs <- err
			}
			close(work)
		}()

		results := make(chan map[string]int)
		workWg := sync.WaitGroup{}

		for i := 0; i < 8; i++ {
			workWg.Add(1)
			go func() {
				for item := range work {
					err := doWork(item, results)
					if err != nil {
						errs <- fmt.Errorf("%s: %w", item, err)
					}
				}
				workWg.Done()
			}()
		}

		wordCountsWg := sync.WaitGroup{}
		var wordCounts map[string]int

		wordCountsWg.Add(1)
		go func() {
			wordCounts = collect(dictionary1, dictionary2, results)
			wordCountsWg.Done()
		}()

		errsWg := sync.WaitGroup{}
		errsWg.Add(1)
		go func() {
			for err := range errs {
				fmt.Println(err)
			}
			errsWg.Done()
		}()

		workWg.Wait()
		close(results)

		wordCountsWg.Wait()

		frequencies := make([]documents.Frequency, len(wordCounts))
		i := 0
		for word, count := range wordCounts {
			frequencies[i] = documents.Frequency{
				Word:  word,
				Count: count,
			}

			i++
		}

		fmt.Println(len(wordCounts))
		sort.Slice(frequencies, func(i, j int) bool {
			return frequencies[i].Count > frequencies[j].Count
		})
		// Just the top 10,000 words.
		frequencies = frequencies[:20000]

		return writeFrequencyTable(out, documents.FrequencyTable{Frequencies: frequencies})
	},
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func writeFrequencyTable(out string, t documents.FrequencyTable) error {
	bytes, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(out), os.ModePerm)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(out, bytes, os.ModePerm)
}

func doWork(path string, results chan<- map[string]int) error {
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

	for i := range doc.Pages {
		counts := countWords(&doc.Pages[i])
		results <- counts
	}

	return nil
}

func countWords(p *documents.Page) map[string]int {
	result := make(map[string]int)

	words := nlp.WordRegex.FindAllString(p.Revision.Text, -1)
	for _, word := range words {
		result[word]++
	}

	return result
}

func dropBelow(m map[string]int, threshold int) map[string]int {
	for w, c := range m {
		if c < threshold {
			delete(m, w)
		}
	}
	return m
}

func collect(dictionary1, dictionary2 map[string]bool, results <-chan map[string]int) map[string]int {
	result := make(map[string]int)

	for counts := range results {
		for word, v := range counts {
			word = nlp.Normalize(word)
			if len(word) == 0 {
				continue
			}
			if !dictionary1[word] && !dictionary2[word] {
				// Not in either dictionary.
				continue
			}
			result[word] += v
			//if len(result) > 1e6 {
			//	result = dropBelow(result, 100)
			//	fmt.Println(len(result))
			//}
		}
	}

	return result
}
