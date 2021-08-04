package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/jobs"

	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/parallel"
	"gopkg.in/yaml.v3"
)

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

		bytes = []byte("<documents>\n" + string(bytes) + "\n</documents>")

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

			doc.Pages[i].Revision.Text = cleanText(doc.Pages[i].Revision.Text)
			result.Pages = append(result.Pages, doc.Pages[i])
		}

		outBytes, err := yaml.Marshal(result)
		if err != nil {
			return err
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
			return fmt.Errorf("writing file: %w", err)
		}

		return nil
	}
}

var (
	widgets = regexp.MustCompile(`{{[^{}]+}}`)
	links   = regexp.MustCompile(`\[\[([^]|]+)(|[^]]+)?]]`) //nolint: gocritic // We want two capture groups.

	wikiClass = regexp.MustCompile(`(?s){\| class=.+?
\|}`)
	timeline = regexp.MustCompile(`(?s)<timeline>
.+?
</timeline>`)

	alpha = regexp.MustCompile(`[A-Za-z]`)
	ref   = regexp.MustCompile(`(?s)<ref( name=.+?)?(>.*?</ref>| />)`)
	link  = regexp.MustCompile(`\[http[^]]+]`)
)

func cleanText(text string) string {
	prevLen := len(text)
	text = widgets.ReplaceAllString(text, "")
	nextLen := len(text)

	for prevLen != nextLen {
		prevLen = nextLen

		text = widgets.ReplaceAllString(text, "")
		nextLen = len(text)
	}

	text = wikiClass.ReplaceAllString(text, "")
	text = ref.ReplaceAllString(text, "")
	text = timeline.ReplaceAllString(text, "")

	lines := strings.Split(text, "\n")

	result := make([]string, 0, len(lines))

	lastLineEmpty := false

	for _, line := range lines {
		if line == "" {
			if !lastLineEmpty {
				result = append(result, line)
			}

			lastLineEmpty = true

			continue
		}

		lastLineEmpty = false

		line = links.ReplaceAllString(line, "$1")

		line = strings.ReplaceAll(line, "&nbsp;", " ")

		line = link.ReplaceAllString(line, "")

		if !hasWords(line) {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func hasWords(line string) bool {
	if !alpha.MatchString(line) {
		return false
	}

	if strings.HasPrefix(line, "Category:") {
		return false
	}

	return true
}
