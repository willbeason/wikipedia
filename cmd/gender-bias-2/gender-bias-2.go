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
	"regexp"
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

	fmt.Println(Female, ":", genderCounts[Female])
	fmt.Println(Male, ":", genderCounts[Male])
	fmt.Println(Multiple, ":", genderCounts[Multiple])
	fmt.Println(Unknown, ":", genderCounts[Unknown])

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
	ratio := float64(genderCounts[Male]) / float64(genderCounts[Female]) // scientist

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

	size := uint32(genderCounts[Male])
	wbs := nlp.MostAccurateWords(menTable, womenTable, size, size)

	for i := 0; i < 1000; i++ {
		wb := wbs[i]
		if wb.Prediction == nlp.Female {
			fmt.Printf("%s => %s : %.02f%%\n", wb.Word, Female, wb.Accuracy*100)
		} else {
			fmt.Printf("%s => %s : %.02f%%\n", wb.Word, Male, wb.Accuracy*100)
		}
	}

	return nil
}

func run() (map[Gender]int, map[string]int, map[string]int, func(chan<- protos.ID) jobs.Page) {
	mtx := sync.Mutex{}

	found := make(map[Gender]int)

	men := map[string]int{}
	women := map[string]int{}

	return found, men, women, func(ids chan<- protos.ID) jobs.Page {

		return func(page *documents.Page) error {
			text := strings.ToLower(page.Text)
			if !strings.Contains(text, "infobox scientist") {
				return nil
			}

			gender := determineGender(page.Text)

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
			case Male:
				//fmt.Println("Men", nMen)
				for word, _ := range f {
					men[word] += 1
				}
			case Female:
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

type Gender string

const (
	Male      Gender = "male"
	Female           = "female"
	Nonbinary        = "nonbinary"
	Multiple         = "multiple"
	Unknown          = "unknown"
)

var (
	categoryRegex = regexp.MustCompile("\\[\\[Category:.+]]")
	womenRegex    = regexp.MustCompile("\\b(women|female)\\b")
	menRegex      = regexp.MustCompile("\\b(men|male)\\b")

	femalePronouns    = regexp.MustCompile("\\b(she|hers|her|herself)\\b")
	malePronouns      = regexp.MustCompile("\\b(he|his|him|himself)\\b")
	nonbinaryPronouns = regexp.MustCompile("\\b(they|their|theirs|them|themself)\\b")
)

func determineGender(text string) Gender {
	categories := categoryRegex.FindAllString(text, -1)

	foundMale := false
	foundFemale := false
	foundNonbinary := false

	for _, category := range categories {
		category = strings.ToLower(category)
		if womenRegex.MatchString(category) {
			foundFemale = true
		} else if menRegex.MatchString(category) {
			foundMale = true
		}
	}

	femaleUsages := len(femalePronouns.FindAllString(text, -1))
	maleUsages := len(malePronouns.FindAllString(text, -1))
	nonbinaryUsages := len(nonbinaryPronouns.FindAllString(text, -1))

	switch {
	case femaleUsages > maleUsages && femaleUsages > nonbinaryUsages:
		foundFemale = true
	case maleUsages > femaleUsages && maleUsages > nonbinaryUsages:
		foundMale = true
	case nonbinaryUsages > femaleUsages && nonbinaryUsages > maleUsages:
		foundNonbinary = true
	}

	switch {
	case foundMale && foundFemale || foundMale && foundNonbinary || foundFemale && foundNonbinary:
		return Multiple
	case foundMale:
		return Male
	case foundFemale:
		return Female
	case foundNonbinary:
		return Nonbinary
	}

	// No signals.
	return Unknown
}
