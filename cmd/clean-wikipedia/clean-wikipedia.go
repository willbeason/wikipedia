package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/parallel"
	"gopkg.in/yaml.v3"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inArticles := args[0]
			out := args[1]

			parallelJobs, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			work := jobs.WalkDir(inArticles, errs)

			workWg := sync.WaitGroup{}
			for i := 0; i < parallelJobs; i++ {
				workWg.Add(1)
				go func() {
					parallel.DoWork(work, doWork(inArticles, out), errs)
					workWg.Done()
				}()
			}

			workWg.Wait()
			close(errs)
			errsWg.Wait()

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

func doWork(in, out string) func(string) error {
	return func(path string) error {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var doc documents.Document
		err = xml.Unmarshal(bytes, &doc)

		if err != nil {
			return err
		}

		result := documents.Document{}

		for i := range doc.Pages {
			if doc.Pages[i].Redirect.Title != "" {
				continue
			}

			if !nlp.IsArticle(doc.Pages[i].Title) {
				continue
			}

			doc.Pages[i].Revision.Text = nlp.CleanArticle(doc.Pages[i].Revision.Text)
			result.Pages = append(result.Pages, doc.Pages[i])
		}

		outBytes, err := yaml.Marshal(result)
		if err != nil {
			return err
		}

		testDoc := documents.Document{}

		err = yaml.Unmarshal(outBytes, &testDoc)
		if err != nil {
			// Print out encountered yaml parsing errors.
			panic(fmt.Sprintf("%s: %v", path, err))
		}

		if !strings.HasPrefix(path, in) {
			panic(path + "\n" + in + "\n")
		}

		outPath := filepath.Join(out, strings.TrimPrefix(path, in))

		err = os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(outPath, outBytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("writing file %s: %w", outPath, err)
		}

		return nil
	}
}
