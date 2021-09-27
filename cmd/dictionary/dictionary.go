package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/ordinality"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

// Defaults for ngram detection which may be made configurable in the future.
const (
	DefaultMinCount = 10000

	DefaultCountFilter   = 40
	DefaultSizeThreshold = 1e7
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: runCmd,
	}

	flags.Parallel(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	inDB := args[0]
	outDictionary := args[1]

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	source := pages.StreamDB(inDB, parallel)

	dictionary := &nlp.Dictionary{}
	prevDictionarySize := -1
	curDictionarySize := 0

	minNgramLength := 1

	errs, errsWg := jobs.Errors()

	ctx := cmd.Context()

	for prevDictionarySize != curDictionarySize {
		fmt.Println("Finding n-grams length", minNgramLength)

		ngramDictionary := nlp.ToNgramDictionary(dictionary)

		tokenizer := nlp.NgramTokenizer{
			Underlying: nlp.WordTokenizer{},
			Dictionary: ngramDictionary,
		}

		results := make(chan *ordinality.PageWordMap, jobs.WorkBuffer)

		var sink protos.Sink = func(_ context.Context, ids <-chan protos.ID, errors chan<- error) (*sync.WaitGroup, error) {
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				for p := range ids {
					results <- p.(*ordinality.PageWordMap)
				}
				wg.Done()
			}()

			return &wg, nil
		}

		frequencyMapChan := ordinality.CollectWordCounts(
			results, ngramDictionary, DefaultCountFilter, DefaultSizeThreshold, DefaultMinCount)

		err = pages.Run(ctx, source, parallel, getNgrams(tokenizer, ngramDictionary, minNgramLength), sink)
		if err != nil {
			return err
		}
		close(results)

		frequencyMap := <-frequencyMapChan

		fmt.Println("Dictionary Words:", len(frequencyMap.Words))

		frequencies := nlp.ToFrequencyTable(frequencyMap)

		prevSize := len(dictionary.Words)
		dictionary.Words = append(dictionary.Words, make([]string, len(frequencies.Words))...)

		for i, word := range frequencies.Words {
			dictionary.Words[prevSize+i] = word.Word
		}

		minNgramLength *= 2
		prevDictionarySize = curDictionarySize
		curDictionarySize = len(dictionary.Words)
		fmt.Printf("Previous Dictionary %d; New Dictionary %d\n",
			prevDictionarySize, curDictionarySize)
	}

	close(errs)
	errsWg.Wait()

	return protos.Write(outDictionary, dictionary)
}

func getNgrams(tokenizer nlp.Tokenizer, known map[string]bool, minLen int) func(chan<- protos.ID) jobs.Page {
	return func(results chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			frequencies := make(map[string]uint32)

			text := page.Text

			// Ignore ngrams resulting from line breaks.
			lines := strings.Split(text, "\n")

			for _, line := range lines {
				tokens := tokenizer.Tokenize(line)

				if minLen == 1 {
					addWords(frequencies, tokens)
				} else {
					addNgrams(frequencies, tokens, known, minLen)
				}
			}

			results <- &ordinality.PageWordMap{
				Id:    page.Id,
				Words: frequencies,
			}

			return nil
		}
	}
}

func addWords(frequencies map[string]uint32, tokens []string) {
	for _, token := range tokens {
		frequencies[token]++
	}
}

func addNgrams(frequencies map[string]uint32, tokens []string, known map[string]bool, minLen int) {
	if len(tokens) < 2 {
		return
	}

	minSpaces := minLen - 1

	jPrevKnown := known[tokens[0]]
	for j := 1; j < len(tokens); j++ {
		// Only add an n-gram if both of its halves are already known.
		jCurKnown := known[tokens[j]]
		if !jPrevKnown || !jCurKnown {
			jPrevKnown = jCurKnown
			continue
		}

		jPrevKnown = true

		sb := strings.Builder{}
		sb.WriteString(tokens[j-1])
		sb.WriteString(" ")
		sb.WriteString(tokens[j])

		ngram := sb.String()
		if strings.Count(ngram, " ") < minSpaces {
			continue
		}

		frequencies[sb.String()]++
	}
}
