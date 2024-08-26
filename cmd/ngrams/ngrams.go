package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/nlp"
	"github.com/willbeason/wikipedia/pkg/ordinality"
	"github.com/willbeason/wikipedia/pkg/pages"
)

// Defaults for ngram detection which may be made configurable in the future.
const (
	// DefaultCountFilter is the threshold below which to discard n-grams when filtering.
	DefaultCountFilter = 40
	// DefaultSizeThreshold is when to trigger an automatic filtering of n-grams.
	DefaultSizeThreshold = 1e7

	SubSampleFlag = "subsample"
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `first-links`,
		Short: `Analyzes the network of references between biographical articles.`,
		RunE:  runCmd,
	}

	cmd.Flags().Float64(SubSampleFlag, 1.0, "proportion of articles to analyze")

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

type TokenCount struct {
	Token string
	Count uint32
}

func runCmd(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	inDB := filepath.Join(environment.WikiPath, "normalized.db")

	ctx, cancel := context.WithCancelCause(cmd.Context())

	dictionary := &nlp.Dictionary{}
	prevDictionarySize := -1
	curDictionarySize := 0

	minNgramLength := 1

	ngramCountsMap := make(map[string]uint32)

	subsample, err := cmd.Flags().GetFloat64(SubSampleFlag)
	if err != nil {
		return fmt.Errorf("%w: unable to parse subsample flag", err)
	}

	for prevDictionarySize != curDictionarySize {
		fmt.Println("Finding n-grams length", minNgramLength)

		ngramDictionary := nlp.ToNgramDictionary(dictionary)

		tokenizer := nlp.NgramTokenizer{
			Underlying: nlp.WordTokenizer{},
			Dictionary: ngramDictionary,
		}

		source := pages.StreamDB[documents.Page](inDB, parallel)
		docs, err2 := source(ctx, cancel)
		if err2 != nil {
			return err2
		}

		// Count words on each page.
		countsChanel, countWork := jobs.MapOld(jobs.WorkBuffer, docs, func(page *documents.Page) (map[string]uint32, error) {
			pageWordCounts := make(map[string]uint32)
			// 10% subsample.

			if subsample < 1.0 {
				if rand.Float64() > subsample {
					return pageWordCounts, nil
				}
			}

			// Ignore ngrams resulting from line breaks.
			lines := strings.Split(page.Text, "\n")
			for _, line := range lines {
				tokens := tokenizer.Tokenize(line)

				if minNgramLength == 1 {
					addWords(pageWordCounts, tokens)
				} else {
					addNgrams(pageWordCounts, tokens, tokenizer.Dictionary, minNgramLength)
				}
			}

			return pageWordCounts, nil
		})

		knownNgrams := ordinality.WordCollector{
			CountFilter:   DefaultCountFilter,
			SizeThreshold: DefaultSizeThreshold,
			Counts:        make(map[string]uint32),
		}
		knownNgramsMtx := sync.Mutex{}

		frequencyMapWork := jobs.Reduce(ctx, jobs.WorkBuffer, countsChanel, func(m map[string]uint32) error {
			knownNgramsMtx.Lock()
			knownNgrams.Add(m)
			knownNgramsMtx.Unlock()
			return nil
		})

		runner := jobs.NewRunner()
		countsWg := runner.Run(ctx, cancel, countWork)
		frequencyMapWg := runner.Run(ctx, cancel, frequencyMapWork)

		countsWg.Wait()
		frequencyMapWg.Wait()

		// Ignore ngrams of insufficient frequency.
		knownNgrams.FilterCounts()
		// Save ngrams and counts.
		for ngram, count := range knownNgrams.Counts {
			ngramCountsMap[ngram] = count
			dictionary.Words = append(dictionary.Words, ngram)
		}

		minNgramLength *= 2
		prevDictionarySize = curDictionarySize
		curDictionarySize = len(dictionary.Words)
		fmt.Printf("Previous Dictionary %d; New Dictionary %d\n",
			prevDictionarySize, curDictionarySize)
	}

	ngramCounts := make([]TokenCount, 0, len(ngramCountsMap))
	for ngram, count := range ngramCountsMap {
		ngramCounts = append(ngramCounts, TokenCount{
			Token: ngram,
			Count: count,
		})
	}
	sort.Slice(ngramCounts, func(i, j int) bool {
		iSpace := strings.Count(ngramCounts[i].Token, " ")
		jSpace := strings.Count(ngramCounts[j].Token, " ")
		if iSpace != jSpace {
			return iSpace > jSpace
		}

		return ngramCounts[i].Count > ngramCounts[j].Count
	})

	for _, nc := range ngramCounts {
		fmt.Printf("%s, %d\n", nc.Token, nc.Count)
	}

	return nil
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
		// Ignore ngrams of insufficient length since we've already counted them.
		if strings.Count(ngram, " ") < minSpaces {
			continue
		}

		frequencies[sb.String()]++
	}
}
