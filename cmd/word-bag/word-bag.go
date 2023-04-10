package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/pages"
	"github.com/willbeason/extract-wikipedia/pkg/protos"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `word-bag path/to/input path/to/dictionary path/to/out`,
		Short: `Convert articles to easily-processable word bags.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := args[0]

	source := pages.StreamDB(inDB, parallel)

	dictionaryPath := args[1]

	dictionary, err := readDictionary(dictionaryPath)
	if err != nil {
		return err
	}

	outDBPath := args[2]

	ctx := cmd.Context()

	fmt.Println("here before wordsets")
	wordSets := make(chan documents.WordSet)

	fmt.Println("here before open out")
	f, err := os.Create(outDBPath)

	if err != nil {
		return err
	}

	defer f.Close()

	fmt.Println("here before wg")
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for ws := range wordSets {
			_, err2 := f.WriteString(ws.String() + "\n")
			if err2 != nil {
				panic(err2)
			}
		}

		wg.Done()
	}()

	fmt.Println("here before run")
	err = pages.Run(ctx, source, parallel, run(dictionary, wordSets), protos.PrintProtos)
	if err != nil {
		return err
	}
	close(wordSets)

	fmt.Println("here before wait write")

	wg.Wait()

	return nil
}

func readDictionary(path string) (map[string]uint32, error) {
	out, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]uint32)

	for i, word := range strings.Split(string(out), "\n") {
		result[strings.TrimSpace(word)] = uint32(i)
	}

	return result, nil
}

func run(dictionary map[string]uint32, out chan documents.WordSet) func(chan<- protos.ID) jobs.Page {
	tokenizer := nlp.WordTokenizer{}

	infoboxFilter, err := documents.NewInfoboxChecker(documents.PersonInfoboxes)
	if err != nil {
		panic(err)
	}

	return func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			if !infoboxFilter.Matches(page.Text) {
				return nil
			}

			text := nlp.CleanArticle(page.Text)
			text = strings.ToLower(text)

			seen := make(map[string]bool)

			documentWords := tokenizer.Tokenize(text)

			wordSet := documents.WordSet{
				ID:    page.Id,
				Words: nil,
			}

			for _, word := range documentWords {
				// Don't repeat words.
				if seen[word] {
					continue
				}

				seen[word] = true

				// Only look for words in dictionary.
				n, exists := dictionary[word]
				if !exists {
					continue
				}

				wordSet.Words = append(wordSet.Words, n)
			}

			sort.Slice(wordSet.Words, func(i, j int) bool {
				return wordSet.Words[i] < wordSet.Words[j]
			})

			out <- wordSet

			return nil
		}
	}
}
