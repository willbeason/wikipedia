package main

import (
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
		Args:  cobra.ExactArgs(4),
		Use:   `word-bag path/to/input path/to/dictionary path/to/wordbags path/to/genders`,
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

	outWordBags := args[2]
	outGenders := args[3]

	ctx := cmd.Context()

	wordSets := make(chan documents.WordSet)
	documentGenders := make(chan nlp.DocumentGender)

	wordBagFile, err := os.Create(outWordBags)
	if err != nil {
		return err
	}

	gendersFile, err := os.Create(outGenders)
	if err != nil {
		return err
	}

	defer wordBagFile.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for ws := range wordSets {
			_, err2 := wordBagFile.WriteString(ws.String() + "\n")
			if err2 != nil {
				panic(err2)
			}
		}

		wg.Done()
	}()

	go func() {
		for dg := range documentGenders {
			_, err2 := gendersFile.WriteString(dg.String() + "\n")
			if err2 != nil {
				panic(err2)
			}
		}

		wg.Done()
	}()

	err = pages.Run(ctx, source, parallel, run(dictionary, wordSets, documentGenders), protos.PrintProtos)
	if err != nil {
		return err
	}
	close(wordSets)
	close(documentGenders)

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

func run(dictionary map[string]uint32, out chan documents.WordSet, out2 chan nlp.DocumentGender) func(chan<- protos.ID) jobs.Page {
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

			gender := nlp.DetermineGender(page.Text)

			out2 <- nlp.DocumentGender{
				ID:     page.Id,
				Gender: gender,
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
