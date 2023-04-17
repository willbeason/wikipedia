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
	"os"
	"regexp"
	"strings"
	"sync"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   `infobox-types path/to/input path/to/out`,
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

	outFile, err := os.Create(args[1])
	if err != nil {
		return err
	}
	defer outFile.Close()

	infoboxes := make(chan GenderInfobox)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err2 := outFile.WriteString("ID,Gender,Infobox\n")
		if err2 != nil {
			panic(err2)
		}

		for i := range infoboxes {
			s := fmt.Sprintf("%d,%s,%s\n", i.ID, i.Gender, i.Infobox)
			_, err2 = outFile.WriteString(s)
			if err2 != nil {
				panic(err2)
			}
		}

		wg.Done()
	}()

	ctx := cmd.Context()

	err = pages.Run(ctx, source, parallel, run(infoboxes), protos.PrintProtos)
	if err != nil {
		return err
	}

	close(infoboxes)

	wg.Wait()

	return nil
}

type GenderInfobox struct {
	ID      uint32
	Gender  nlp.Gender
	Infobox string
}

func run(infoboxes chan GenderInfobox) func(chan<- protos.ID) jobs.Page {

	infoboxFilter, err := documents.NewInfoboxChecker(documents.PersonInfoboxes)
	if err != nil {
		panic(err)
	}

	infoboxFinder := regexp.MustCompile(`\binfobox( \w+)+\n`)

	return func(ids chan<- protos.ID) jobs.Page {
		return func(page *documents.Page) error {
			if !infoboxFilter.Matches(page.Text) {
				return nil
			}

			text := strings.ToLower(page.Text)

			found := infoboxFinder.FindAllString(text, -1)

			gender := nlp.DetermineGender(text)

			for _, infobox := range found {
				infoboxes <- GenderInfobox{
					ID:      page.Id,
					Gender:  gender,
					Infobox: infobox[8 : len(infobox)-1],
				}
			}

			return nil
		}
	}
}
