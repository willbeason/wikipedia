package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/pages"
	"github.com/willbeason/extract-wikipedia/pkg/protos"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `article-size path/to/input path/to/dictionary path/to/out`,
		Short: `Calculate article sizes of men and women`,
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

	dictionaryFile := args[1]

	dictionary, err := nlp.ReadDictionary(dictionaryFile)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	sizes := make(chan GenderSize)

	outFile, err := os.Create(args[2])
	if err != nil {
		return err
	}

	defer func(outFile *os.File) {
		err2 := outFile.Close()
		if err2 != nil {
			panic(err2)
		}
	}(outFile)

	writeFileWaitGroup := sync.WaitGroup{}
	writeFileWaitGroup.Add(1)

	go func() {
		_, err2 := outFile.WriteString("ID,Gender,Size\n")
		if err2 != nil {
			panic(err2)
		}

		for gs := range sizes {
			s := fmt.Sprintf("%d,%s,%d\n", gs.ID, gs.Gender, gs.Size)

			_, err2 = outFile.WriteString(s)
			if err2 != nil {
				panic(err2)
			}
		}

		writeFileWaitGroup.Done()
	}()

	err = pages.Run(ctx, source, parallel, run(dictionary, sizes), protos.PrintProtos)
	if err != nil {
		return err
	}

	close(sizes)

	writeFileWaitGroup.Wait()

	return nil
}

type GenderSize struct {
	ID     uint32
	Gender nlp.Gender
	Size   int
}

func run(dictionary *nlp.Dictionary, sizes chan GenderSize) func(chan<- protos.ID) jobs.Page {
	tokenizer := nlp.WordTokenizer{}

	allowedWords := dictionary.ToSet()

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

			seen := make(map[string]bool)
			count := 0

			for _, w := range tokenizer.Tokenize(text) {
				if allowedWords[w] {
					seen[w] = true
					count++
				}
			}

			sizes <- GenderSize{
				ID:     page.Id,
				Gender: gender,
				Size:   len(seen),
			}

			return nil
		}
	}
}
