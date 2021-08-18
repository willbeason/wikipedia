package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"github.com/willbeason/extract-wikipedia/pkg/ordinality"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			inArticles := args[0]
			inDictionary := args[1]
			outArticles := args[2]

			parallel, err := cmd.Flags().GetInt(flags.ParallelKey)
			if err != nil {
				return err
			}

			dictionary, err := nlp.ReadDictionary(inDictionary)
			if err != nil {
				return err
			}

			converter := ordinality.WordBagConverter{
				Tokenizer:      nlp.NgramTokenizer{
					Underlying: nlp.WordTokenizer{},
					Dictionary: nlp.ToNgramDictionary(dictionary),
				},
				WordOrdinality: ordinality.NewWordOrdinality(dictionary),
			}

			errs, errsWg := jobs.Errors()

			work := jobs.WalkDir(inArticles, errs)

			workWg := jobs.DoDocumentJobs(parallel, documentWordBags(inArticles, outArticles, converter), work, errs)

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

func documentWordBags(in, out string, converter ordinality.WordBagConverter) jobs.Document {
	return func(doc *documents.Document) error {
		wordBags := converter.ToDocumentWordBag(doc)

		path := doc.Path
		outPath := filepath.Join(out, strings.TrimPrefix(path, in))
		outExt := filepath.Ext(outPath)
		outPath = strings.TrimSuffix(outPath, outExt) + ".pb"

		return ordinality.WriteWordBags(outPath, wordBags)
	}
}
