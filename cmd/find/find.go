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
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inArticles := args[0]
			search := args[1]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			errsWg := sync.WaitGroup{}
			errsWg.Add(1)
			errs, _ := jobs.Errors(&errsWg)

			workWg := sync.WaitGroup{}
			work := jobs.WalkDir(inArticles, errs)

			results := make(chan string)

			for i := 0; i < parallel; i++ {
				workWg.Add(1)
				jobs.DoJobs(doJob(search, results), &workWg, work, errs)
			}

			resultsWg := sync.WaitGroup{}
			resultsWg.Add(1)

			printResults(&resultsWg, results)

			workWg.Wait()
			close(results)
			close(errs)
			errsWg.Wait()

			resultsWg.Wait()

			return nil
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

func doJob(find string, results chan<- string) jobs.Job {
	return func(page *documents.Page) error {
		text := page.Revision.Text

		lines := strings.Split(text, "\n")

		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), find) {
				results <- page.Title + "\n" + line + "\n"
			}
		}

		return nil
	}
}

func printResults(resultWg *sync.WaitGroup, results <-chan string) {
	go func() {
		n := 0

		for result := range results {
			fmt.Println(result)

			n++
			if n >= 10 {
				panic("printed ten results")
			}
		}

		resultWg.Done()
	}()
}
