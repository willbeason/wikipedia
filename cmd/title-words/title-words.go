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
	"strings"
	"sync"
)

var cmd = cobra.Command{
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		inArticles := args[0]
		inDictionary := args[1]
		outDictionary := args[2]

		dictionary, err := nlp.ReadDictionary(inDictionary)
		if err != nil {
			return err
		}

		work := make(chan string)
		errs := make(chan error)
		results := make(chan string)

		go func() {
			err := filepath.WalkDir(inArticles, walker.Files(work))
			if err != nil {
				errs <- err
			}
			close(work)
		}()

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

		var titleDictionary map[string]bool
		titleWg := sync.WaitGroup{}
		titleWg.Add(1)

		go func() {
			titleDictionary = collectTitleDictionary(dictionary, results)
			titleWg.Done()
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
		close(errs)
		errsWg.Wait()

		titleWg.Wait()

		return nlp.WriteDictionary(outDictionary, titleDictionary)
	},
}

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func doWork(path string, results chan<- string) error {
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
		title := doc.Pages[i].Title
		if !nlp.IsArticle(title) {
			continue
		}
		for _, word := range nlp.WordRegex.FindAllString(title, -1) {
			if !nlp.HasLetter(word) {
				continue
			}
			word = nlp.Normalize(word)
			results <- word
		}
	}

	return nil
}

func collectTitleDictionary(dictionary map[string]bool, words <-chan string) map[string]bool {
	counts := make(map[string]int)
	result := make(map[string]bool)

	for word := range words {
		if dictionary[word] {
			continue
		}
		counts[word]++
		if counts[word] == 2 {
			result[word] = true
		}
	}

	return result
}
