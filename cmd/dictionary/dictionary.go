package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

// Defaults for ngram detection which may be made configurable in the future.
const (
	DefaultMinCount = 10000

	DefaultCountFilter   = 80
	DefaultSizeThreshold = 4e6
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inArticles := args[0]
			outDictionary := args[1]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			dictionary := &nlp.Dictionary{}
			prevDictionarySize := -1
			curDictionarySize := 0

			minNgramLength := 1

			errs, errsWg := jobs.Errors()

			for prevDictionarySize != curDictionarySize {
				fmt.Println("Finding n-grams length", minNgramLength)

				ngramDictionary := nlp.ToNgramDictionary(dictionary)

				work := jobs.WalkDir(inArticles, errs)
				tokenizer := nlp.NgramTokenizer{
					Underlying: nlp.WordTokenizer{},
					Dictionary: ngramDictionary,
				}

				results := make(chan map[string]uint32, 100)
				workWg := jobs.DoPageJobs(parallel, getNgrams(tokenizer, ngramDictionary, minNgramLength-1, results), work, errs)

				frequencyMapChan := nlp.CollectWordCounts(
					results, ngramDictionary, DefaultCountFilter, DefaultSizeThreshold, DefaultMinCount)

				workWg.Wait()
				close(results)

				frequencyMap := <-frequencyMapChan

				fmt.Println("Dictionary Words:", len(frequencyMap.Words))

				frequencies := nlp.ToFrequencyTable(frequencyMap)

				prevSize := len(dictionary.Words)
				dictionary.Words = append(dictionary.Words, make([]string, len(frequencies.Words))...)

				for i, word := range frequencies.Words {
					dictionary.Words[prevSize + i] = word.Word
				}

				minNgramLength *= 2
				prevDictionarySize = curDictionarySize
				curDictionarySize = len(dictionary.Words)
				fmt.Printf("Previous Dictionary %d; New Dictionary %d\n",
					prevDictionarySize, curDictionarySize)
			}

			close(errs)
			errsWg.Wait()

			return nlp.WriteDictionary(outDictionary, dictionary)
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

func addWords(frequencies map[string]uint32, tokens []string) {
	for _, token := range tokens {
		frequencies[token]++
	}
}

func addNgrams(frequencies map[string]uint32, tokens []string, known map[string]bool, minLen int) {
	if len(tokens) < 2 {
		return
	}

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
		if strings.Count(ngram, " ") < minLen {
			continue
		}

		frequencies[sb.String()]++
	}
}

func getNgrams(tokenizer nlp.Tokenizer, known map[string]bool, minLen int, results chan<- map[string]uint32) jobs.Page {
	return func(page *documents.Page) error {
		frequencies := make(map[string]uint32)

		text := page.Revision.Text

		// Ignore ngrams resulting from line breaks.
		lines := strings.Split(text, "\n")

		for _, line := range lines {
			tokens := tokenizer.Tokenize(line)

			if minLen == 0 {
				addWords(frequencies, tokens)
			} else {
				addNgrams(frequencies, tokens, known, minLen)
			}
		}

		results <- frequencies

		return nil
	}
}
