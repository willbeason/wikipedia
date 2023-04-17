package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/db"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/pages"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
	"io"
	"os"
	"strings"
	"sync"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MinimumNArgs(1),
		Use:   `gender-bias path/to/input`,
		Short: `Analyze gender disparity in articles.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

const wordList = "data/wordlist.txt"

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

	pageIDs, err := cmd.Flags().GetUintSlice(flags.IDsKey)
	if err != nil {
		return err
	}

	inDB := args[0]

	var outDBPath string
	var sink protos.Sink

	if len(args) > 1 {
		outDBPath = args[1]
		outDB := db.NewRunner(outDBPath, parallel)
		sink = outDB.Write()
	} else {
		sink = protos.PrintProtos
	}

	wordListFile, err := os.Open(wordList)
	if err != nil {
		return err
	}

	wordsBytes, err := io.ReadAll(wordListFile)
	if err != nil {
		return err
	}

	words := strings.Split(string(wordsBytes), "\n")

	allowedWord := make(map[string]bool)
	for _, word := range words {
		allowedWord[strings.TrimSpace(word)] = true
	}
	fmt.Println(len(allowedWord), "allowed words")

	var source pages.Source
	if len(pageIDs) == 0 {
		source = pages.StreamDB(inDB, parallel)
	} else {
		source = pages.StreamDBKeys(inDB, parallel, pageIDs)
	}

	cmd.SilenceUsage = true
	ctx := cmd.Context()

	genderCounts, men, women, job := run()

	err = pages.Run(ctx, source, parallel, job, sink)
	if err != nil {
		return err
	}

	fmt.Println(nlp.Female, ":", genderCounts[nlp.Female])
	fmt.Println(nlp.Male, ":", genderCounts[nlp.Male])
	fmt.Println(nlp.Multiple, ":", genderCounts[nlp.Multiple])
	fmt.Println(nlp.Unknown, ":", genderCounts[nlp.Unknown])

	menTable := &nlp.FrequencyTable{}
	for word, count := range men {
		if !allowedWord[word] {
			if word == "her" {
				panic("WHAT")
			}
			continue
		}
		menTable.Words = append(menTable.Words, &nlp.WordCount{
			Word:  word,
			Count: uint32(count),
		})
	}

	// For normalizing for equal counts between men and women.
	ratio := float64(genderCounts[nlp.Male]) / float64(genderCounts[nlp.Female]) // scientist

	womenTable := &nlp.FrequencyTable{}
	for word, count := range women {
		if !allowedWord[word] {
			continue
		}
		womenTable.Words = append(womenTable.Words, &nlp.WordCount{
			Word: word,
			// Adjust for equity.
			Count: uint32(float64(count) * ratio),
		})
	}

	size := uint32(genderCounts[nlp.Male])
	wbs := nlp.MostAccurateWords(menTable, womenTable, size, size)

	for i := 0; i < 1000; i++ {
		wb := wbs[i]
		if wb.Prediction == nlp.Female {
			fmt.Printf("%s => %s : %.02f%%\n", wb.Word, nlp.Female, wb.Accuracy*100)
		} else {
			fmt.Printf("%s => %s : %.02f%%\n", wb.Word, nlp.Male, wb.Accuracy*100)
		}
	}

	return nil
}

func run() (map[nlp.Gender]int, map[string]int, map[string]int, func(chan<- protos.ID) jobs.Page) {
	mtx := sync.Mutex{}

	found := make(map[nlp.Gender]int)

	men := map[string]int{}
	women := map[string]int{}

	return found, men, women, func(_ chan<- protos.ID) jobs.Page {

		return func(page *documents.Page) error {
			text := strings.ToLower(page.Text)
			if !strings.Contains(text, "infobox scientist") {
				return nil
			}

			gender := nlp.DetermineGender(page.Text)

			f := make(map[string]int)

			tokenizer := nlp.WordTokenizer{}
			words := tokenizer.Tokenize(text)

			for _, word := range words {
				f[word]++
			}

			mtx.Lock()
			//fmt.Println(page.Title, "|", gender)
			found[gender]++

			switch gender {
			case nlp.Male:
				//fmt.Println("Men", nMen)
				for word, _ := range f {
					men[word] += 1
				}
			case nlp.Female:
				//fmt.Println("Women", nWomen)
				for word, _ := range f {
					women[word] += 1
				}
			}
			mtx.Unlock()

			return nil
		}
	}
}
