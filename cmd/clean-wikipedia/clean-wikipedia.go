package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"gopkg.in/yaml.v3"
)

// clean-wikipedia removes parts of articles we never want to analyze, such as xml tags, tables, and
// formatting directives.

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		Use: `clean-wikipedia path/to/input path/to/output`,
		Short: `Cleans an extracted set of Wikipedia articles by removing irrelevant pages and formatting
directives.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inArticles := args[0]
			out := args[1]

			parallelJobs, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			errs, errsWg := jobs.Errors()
			work := jobs.WalkDir(inArticles, errs)

			workWg := jobs.DoDocumentJobs(parallelJobs, doWork(inArticles, out), work, errs)

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

func doWork(in, out string) jobs.Document {
	return func(doc *documents.Document) error {
		result := documents.Document{}

		for i := range doc.Pages {
			if doc.Pages[i].Redirect.Title != "" {
				continue
			}

			title := doc.Pages[i].Title
			if !nlp.IsArticle(title) {
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

		path := doc.Path

		err = yaml.Unmarshal(outBytes, &testDoc)
		if err != nil {
			// Print out encountered yaml parsing errors.
			// We want to panic and not complete the job as otherwise we'll write data
			// we will not be able to read in the future.
			panic(fmt.Sprintf("%s: %v", path, err))
		}

		if !strings.HasPrefix(path, in) {
			panic(path + "\n" + in + "\n")
		}

		outPath := filepath.Join(out, strings.TrimPrefix(path, in))
		outExt := filepath.Ext(outPath)
		outPath = strings.TrimSuffix(outPath, outExt) + ".yaml"

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
