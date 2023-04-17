package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/pages"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `calculateTfidf path/to/input path/to/dictionary path/to/out`,
		Short: `Calculate max calculateTfidf for articles`,
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

	cmd.SilenceUsage = true

	parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
	if err != nil {
		return err
	}

	inDB := args[0]
	source := pages.StreamDB(inDB, parallel)

	dictionaryFile := args[1]
	dictionary, err := nlp.ReadDictionary(dictionaryFile)

	ctx := cmd.Context()

	// Calculate idfs
	fmt.Println("Calculating IDFs")

	counts := make([]float64, len(dictionary.Words))
	docs := 0

	err = pages.Run(ctx, source, parallel, count(dictionary, counts, &docs), protos.PrintProtos)
	if err != nil {
		return err
	}

	weights := make([]WordWeight, len(dictionary.Words))

	for i, c := range counts {
		weights[i] = WordWeight{
			Word:   dictionary.Words[i],
			Weight: math.Log(float64(docs) / c),
		}
	}

	// Calculate tfidfs
	fmt.Println("Calculating TFIDFs")

	igws := make(chan IDGenderWord)

	outFile, err := os.Create(args[2])
	if err != nil {
		return err
	}
	defer outFile.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err2 := outFile.WriteString("ID,Gender,Word,TFIDF\n")
		if err2 != nil {
			panic(err2)
		}

		for gs := range igws {
			s := fmt.Sprintf("%d,%s,%s,%.02f\n", gs.ID, gs.Gender, gs.Word, gs.TFIDF)
			_, err2 = outFile.WriteString(s)
			if err2 != nil {
				panic(err2)
			}
		}

		wg.Done()
	}()

	err = pages.Run(ctx, source, parallel, calculateTfidf(dictionary, weights, igws), protos.PrintProtos)
	if err != nil {
		return err
	}

	close(igws)

	wg.Wait()

	return nil
}

type WordWeight struct {
	Word   string
	Weight float64
}

type IDGenderWord struct {
	ID     uint32
	Gender nlp.Gender
	Word   string
	TFIDF  float64
}

func count(dictionary *nlp.Dictionary, counts []float64, articles *int) func(chan<- protos.ID) jobs.Page {
	tokenizer := nlp.WordTokenizer{}

	mtx := sync.Mutex{}

	lookup := make(map[string]int)

	for i, w := range dictionary.Words {
		lookup[w] = i
	}

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

			text := strings.ToLower(nlp.CleanArticle(page.Text))
			words := tokenizer.Tokenize(text)
			idxs := make([]float64, len(dictionary.Words))
			seen := make(map[string]bool)

			for _, w := range words {
				if seen[w] {
					continue
				}
				seen[w] = true

				if gender == nlp.Female {
					idxs[lookup[w]] += 1.0
				} else if gender == nlp.Male {
					idxs[lookup[w]]++
				}
			}

			mtx.Lock()
			for i, w := range idxs {
				counts[i] += w
			}
			*articles = *articles + 1
			mtx.Unlock()

			return nil
		}
	}
}

func calculateTfidf(dictionary *nlp.Dictionary, weights []WordWeight, igws chan IDGenderWord) func(chan<- protos.ID) jobs.Page {
	tokenizer := nlp.WordTokenizer{}

	lookup := make(map[string]int)

	for i, w := range dictionary.Words {
		lookup[w] = i
	}

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

			text := strings.ToLower(nlp.CleanArticle(page.Text))
			words := tokenizer.Tokenize(text)

			tfidfs := make(map[string]float64)

			for _, w := range words {
				idx := lookup[w]
				tfidfs[w] += weights[idx].Weight
			}

			idgs := make([]IDGenderWord, 0, len(tfidfs))
			for w, tfidf := range tfidfs {
				idgs = append(idgs, IDGenderWord{
					ID:     page.Id,
					Gender: gender,
					Word:   w,
					TFIDF:  tfidf,
				})
			}

			sort.Slice(idgs, func(i, j int) bool {
				if idgs[i].TFIDF != idgs[j].TFIDF {
					return idgs[i].TFIDF > idgs[j].TFIDF
				}

				return idgs[i].Word < idgs[j].Word
			})

			for i := 0; i < 10 && i < len(idgs); i++ {
				igws <- idgs[i]
			}

			return nil
		}
	}
}
