package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/nlp/words"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/db"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/pages"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
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

	var source pages.Source
	if len(pageIDs) == 0 {
		source = pages.StreamDB(inDB, parallel)
	} else {
		source = pages.StreamDBKeys(inDB, parallel, pageIDs)
	}

	cmd.SilenceUsage = true
	ctx := cmd.Context()

	found, men, women, job := run()

	// out := make(map[string]uint32)

	err = pages.Run(ctx, source, parallel, job, sink)
	// err = pages.Run(ctx, source, parallel, findInfoboxes(out), sink)
	if err != nil {
		return err
	}

	for gender, count := range found {
		fmt.Println(gender, ":", count)
	}

	// counts := make([]*nlp.WordCount, 0, len(out))
	// for item, count := range out {
	//	counts = append(counts, &nlp.WordCount{
	//		Word:  item,
	//		Count: count,
	//	})
	//}
	//
	// sort.Slice(counts, func(i, j int) bool {
	//	return counts[i].Count > counts[j].Count
	// })
	//
	// for _, c := range counts {
	//	fmt.Println(c.Word, ":", c.Count)
	//}

	allowedWords := make(map[string]bool)
	for _, word := range words.WordList {
		allowedWords[word] = true
	}

	menTable := &nlp.FrequencyTable{}
	for word, count := range men {
		if !allowedWords[word] {
			continue
		}
		menTable.Words = append(menTable.Words, &nlp.WordCount{
			Word:  word,
			Count: uint32(count),
		})
	}

	const nMen = 945267
	const nWomen = 262992
	ratio := float64(nMen) / float64(nWomen)

	womenTable := &nlp.FrequencyTable{}
	for word, count := range women {
		if !allowedWords[word] {
			continue
		}
		womenTable.Words = append(womenTable.Words, &nlp.WordCount{
			Word: word,
			// Adjust for equity.
			Count: uint32(float64(count) * ratio),
		})
	}

	wbs := nlp.CharacteristicWords(5, menTable, womenTable)

	for i := 0; i < 5000; i++ {
		wb := wbs[i]
		fmt.Println(wb.Word, ":", wb.Bits)
	}

	return nil
}

var infoboxRegex = regexp.MustCompile(`infobox( \w+)+`)

func findInfoboxes(out map[string]uint32) func(chan<- protos.ID) jobs.Page {
	outMtx := sync.Mutex{}

	return func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			text := strings.ToLower(page.Text)
			matches := infoboxRegex.FindAllString(text, -1)

			outMtx.Lock()
			for _, match := range matches {
				out[match]++
			}
			outMtx.Unlock()

			return nil
		}
	}
}

func run() (map[Gender]int, map[string]int, map[string]int, func(chan<- protos.ID) jobs.Page) {
	mtx := sync.Mutex{}

	found := make(map[Gender]int)

	men := map[string]int{}
	women := map[string]int{}

	personInfoboxes, err := documents.NewInfoboxChecker(documents.PersonInfoboxes)
	if err != nil {
		panic(err)
	}

	return found, men, women, func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			text := strings.ToLower(page.Text)

			if !personInfoboxes.Matches(page.Text) {
				return nil
			}

			gender := determineGender(page.Text)

			f := make(map[string]int)

			tokenizer := nlp.WordTokenizer{}
			documentWords := tokenizer.Tokenize(text)

			for _, word := range documentWords {
				f[word]++
			}

			mtx.Lock()
			// fmt.Println(page.Title, "|", gender)
			found[gender]++

			switch gender {
			case Male:
				for word := range f {
					men[word] += 1
				}
			case Female:
				for word := range f {
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
	Both             = "both"
	Unknown          = "unknown"
)

var (
	categoryRegex = regexp.MustCompile("\\[\\[Category:.+]]")
	womenRegex    = regexp.MustCompile("\\b(women|female)\\b")
	menRegex      = regexp.MustCompile("\\b(men|male)\\b")

	femalePronouns    = regexp.MustCompile("\\b(she|hers|her|herself)\\b")
	malePronouns      = regexp.MustCompile("\\b(he|his|him|himself)\\b")
	nonbinaryPronouns = regexp.MustCompile("\\b(they|theirs|them|themself)\\b")
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

	text = nlp.CleanArticle(text)

	femaleUsages := len(femalePronouns.FindAllString(text, -1))
	maleUsages := len(malePronouns.FindAllString(text, -1))
	nonbinaryUsages := len(nonbinaryPronouns.FindAllString(text, -1))

	switch {
	case femaleUsages > maleUsages && femaleUsages > nonbinaryUsages:
		foundFemale = true
	case maleUsages > femaleUsages && maleUsages > nonbinaryUsages:
		foundMale = true
	case nonbinaryUsages > maleUsages && nonbinaryUsages > femaleUsages:
		foundNonbinary = true
	}

	switch {
	case foundMale && foundFemale:
		return Both
	case foundMale:
		return Male
	case foundFemale:
		return Female
	case foundNonbinary:
		return Nonbinary
	}

	return Unknown
}
