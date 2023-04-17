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
	"sort"
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

	genderCounts, data, job := run()

	err = pages.Run(ctx, source, parallel, job, sink)
	if err != nil {
		return err
	}

	fmt.Println(Female, ":", genderCounts[Female])
	fmt.Println(Male, ":", genderCounts[Male])
	fmt.Println(Multiple, ":", genderCounts[Multiple])
	fmt.Println(Unknown, ":", genderCounts[Unknown])

	fmt.Println("Found", len(data), "Articles")

	wordSet := make(map[string]int)
	for _, d := range data {
		for _, w := range d.WordBag {
			wordSet[w] += 1
		}
	}

	topWords := make([]string, 0, len(wordSet))
	for w := range wordSet {
		if !allowedWord[w] {
			continue
		}
		topWords = append(topWords, w)
	}

	sort.Slice(topWords, func(i, j int) bool {
		wci := wordSet[topWords[i]]
		wcj := wordSet[topWords[j]]
		if wci != wcj {
			return wci > wcj
		}
		return topWords[i] < topWords[j]
	})

	if len(topWords) > 400 {
		topWords = topWords[:400]
	}

	// For normalizing for equal counts between men and women.
	ratio := float64(genderCounts[Male]) / float64(genderCounts[Female]) // scientist
	fmt.Printf("Bias %.02f\n", ratio)

	perceptron := nlp.NewPerceptron(topWords, ratio)
	trainingData := make([]nlp.Datum, 0, len(data))
	for _, d := range data {
		trainingData = append(trainingData, perceptron.ToDatum(d.WordBag, d.Label))
	}

	bestCost := 1e9
	var bestPerceptron *nlp.Perceptron

	for modelN := 0; modelN < 1000; modelN++ {
		if modelN%100 == 0 {
			fmt.Println("Training Model", modelN)
		}
		perceptron = nlp.NewPerceptron(topWords, ratio)

		var curCost float64
		var curAccuracy float64

		for i := 0; i < 1000; i++ {
			//fmt.Println(perceptron.Weights[:20])
			n, cost, accuracy, isNew := perceptron.Train(trainingData)

			if !isNew {
				break
			}
			curCost = cost
			curAccuracy = accuracy

			//fmt.Printf("%d %.02f %.02f%% %t\n", i, cost, accuracy*100, isNew)

			perceptron = n
		}

		if curCost < bestCost {
			bestCost = curCost
			bestPerceptron = perceptron
			fmt.Printf("New Best %d %.02f %.02f%%\n", modelN, curCost, curAccuracy*100)
		}
	}

	results := bestPerceptron.WordOrder()
	for i, r := range results[:20] {
		fmt.Println(i, r.Word, ":", r.Weight)
	}

	return nil
}

type LabeledWordBag struct {
	ID      uint32
	Label   nlp.Gender
	WordBag []string
}

func run() (map[Gender]int, map[uint32]LabeledWordBag, func(chan<- protos.ID) jobs.Page) {
	mtx := sync.Mutex{}

	found := make(map[Gender]int)

	data := make(map[uint32]LabeledWordBag)

	return found, data, func(ids chan<- protos.ID) jobs.Page {

		return func(page *documents.Page) error {
			text := strings.ToLower(page.Text)
			if !strings.Contains(text, "infobox scientist") {
				return nil
			}

			gender := determineGender(page.Text)

			wordBagSet := make(map[string]bool)

			tokenizer := nlp.WordTokenizer{}
			words := tokenizer.Tokenize(text)

			for _, word := range words {
				wordBagSet[word] = true
			}

			wordBagList := make([]string, 0, len(wordBagSet))
			for w := range wordBagSet {
				wordBagList = append(wordBagList, w)
			}

			lwb := LabeledWordBag{
				ID:      page.ID(),
				WordBag: wordBagList,
			}

			if gender == Male {
				lwb.Label = nlp.Male
			} else if gender == Female {
				lwb.Label = nlp.Female
			}

			mtx.Lock()
			data[lwb.ID] = lwb
			//fmt.Println(page.Title, "|", gender)
			found[gender]++
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
