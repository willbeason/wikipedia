package links

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/article"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/pages"
	"github.com/willbeason/wikipedia/pkg/protos"
)

var ErrLinks = errors.New("unable to create links")

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(4),
		Use:   "links corpus_name articles_dir title_index out_file",
		Short: `Create a listing links for every article`,
		RunE:  runCmd,
	}

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	cfg := &config.Links{
		In:    args[1],
		Index: args[2],
		Out:   args[3],
	}

	return Links(cmd, cfg, args[0])
}

func Links(cmd *cobra.Command, cfg *config.Links, corpusNames ...string) error {
	if len(corpusNames) != 1 {
		return fmt.Errorf("%w: must have exactly one corpus but got %+v", ErrLinks, corpusNames)
	}
	corpusName := corpusNames[0]
	articlesDir := cfg.In
	outFile := cfg.Out
	fmt.Printf("Creating link network for corpus %q from directory %q and writing to %q\n",
		corpusName, articlesDir, outFile)

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	articlesDir = filepath.Join(workspace, corpusName, articlesDir)
	outFile = filepath.Join(workspace, corpusName, outFile)
	titleIndexPath := filepath.Join(workspace, corpusName, cfg.Index)

	source := pages.StreamDB[documents.Page](articlesDir, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())
	ps, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	titleIndex, err := protos.Read[documents.TitleIndex](titleIndexPath)
	if err != nil {
		return err
	}

	linksChannel := makeLinks((*titleIndex).Titles, ps)
	links := <-linksChannel

	err = protos.Write(outFile, links)
	if err != nil {
		return fmt.Errorf("%w: writing title index: %w", ErrLinks, err)
	}

	return nil
}

func makeLinks(titleIndex map[string]uint32, pages <-chan *documents.Page) <-chan *documents.LinkIndex {
	results := make(chan *documents.LinkIndex)

	go func() {
		result := &documents.LinkIndex{Articles: make(map[uint32]*documents.Links)}

		for page := range pages {
			fmt.Println(page.Title)
			tokens := article.Tokenize(article.UnparsedText(page.Text))

			links := &documents.Links{}

			for _, token := range tokens {
				switch t := token.(type) {
				case article.Link:
					id, found := titleIndex[t.Target.Render()]
					if !found {
						continue
					}

					links.Links = append(links.Links, id)
				}
			}

			result.Articles[page.Id] = links
		}

		results <- result
		close(results)
	}()

	return results
}
