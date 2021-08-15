package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"gopkg.in/yaml.v3"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
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
		for i := range doc.Pages {
			text := doc.Pages[i].Revision.Text
			doc.Pages[i].Revision.Text = nlp.NormalizeArticle(text)
		}

		outBytes, err := yaml.Marshal(doc)
		if err != nil {
			return err
		}

		path := doc.Path

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
