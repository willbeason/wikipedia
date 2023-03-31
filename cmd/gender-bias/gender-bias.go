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

	men, women, job := run()

	err = pages.Run(ctx, source, parallel, job, sink)
	if err != nil {
		return err
	}

	menTable := &nlp.FrequencyTable{}
	for word, count := range *men {
		menTable.Words = append(menTable.Words, &nlp.WordCount{
			Word:  word,
			Count: uint32(count),
		})
	}

	womenTable := &nlp.FrequencyTable{}
	for word, count := range *women {
		womenTable.Words = append(womenTable.Words, &nlp.WordCount{
			Word: word,
			// Adjust for equity.
			Count: uint32(count * 151 / 47),
		})
	}

	wbs := nlp.CharacteristicWords(5, menTable, womenTable)

	for i := 0; i < len(wbs); i++ {
		wb := wbs[i]
		fmt.Println(wb.Word, ":", wb.Bits)
	}

	return nil
}

func run() (*map[string]int, *map[string]int, func(chan<- protos.ID) jobs.Page) {
	nMen := 0
	nWomen := 0
	mtx := sync.Mutex{}

	found := make(map[Gender]int)

	men := map[string]int{}
	women := map[string]int{}

	return &men, &women, func(ids chan<- protos.ID) jobs.Page {

		return func(page *documents.Page) error {
			if !strings.Contains(page.Text, "infobox person") {
				return nil
			}

			gender := determineGender(page.Text)

			text := page.Text
			text = strings.ToLower(text)

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
				nMen++
				//fmt.Println("Men", nMen)
				for word, _ := range f {
					men[word] += 1
				}
			case Female:
				nWomen++
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
	Male    Gender = "male"
	Female         = "female"
	Both           = "both"
	Unknown        = "unknown"
)

var (
	categoryRegex = regexp.MustCompile("\\[\\[Category:.+]]")
	womenRegex    = regexp.MustCompile("\\b(women|female)\\b")
	menRegex      = regexp.MustCompile("\\b(men|male)\\b")

	femalePronouns = regexp.MustCompile("\\b(she|hers|her|herself)\\b")
	malePronouns   = regexp.MustCompile("\\b(he|his|him|himself)\\b")
)

func determineGender(text string) Gender {
	categories := categoryRegex.FindAllString(text, -1)

	foundMale := false
	foundFemale := false

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

	if femaleUsages > maleUsages {
		foundFemale = true
	} else {
		foundMale = true
	}

	switch {
	case foundMale && foundFemale:
		return Both
	case foundMale:
		return Male
	case foundFemale:
		return Female
		//default:
		//	fmt.Println(categories)
		//	return Unknown
	}

	return Unknown
}
