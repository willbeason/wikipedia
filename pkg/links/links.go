package links

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/article"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
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
	outPath := cfg.Out
	fmt.Printf("Creating link network for corpus %q from directory %q and writing to %q\n",
		corpusName, articlesDir, outPath)

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	articlesDir = filepath.Join(workspace, corpusName, articlesDir)
	outPath = filepath.Join(workspace, corpusName, outPath)
	titleIndexPath := filepath.Join(workspace, corpusName, cfg.Index)
	redirectsPath := filepath.Join(workspace, corpusName, cfg.Redirects)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	titleIndexFuture := documents.ReadTitleMap(ctx, titleIndexPath, errs)
	titleIndex := <-titleIndexFuture

	redirectIndex := <-documents.MakeRedirects(ctx, redirectsPath, errs)

	pageSource := jobs.NewSource(protos.ReadDir[documents.Page](articlesDir))
	pageSourceWg, pageSourceJob, pages := pageSource()
	go pageSourceJob(ctx, errs)

	linksChannel := makeLinks(parallel, titleIndex, redirectIndex, pages)

	linksSink := jobs.NewSink(protos.WriteFile[*documents.ArticleIdLinks](outPath))
	linksSinkWg, linksSinkJob := linksSink(linksChannel)
	go linksSinkJob(ctx, errs)

	pageSourceWg.Wait()
	linksSinkWg.Wait()
	if ctx.Err() != nil {
		return fmt.Errorf("%w: writing title index: %w", ErrLinks, context.Cause(ctx))
	}

	return nil
}

func makeLinks(
	parallel int,
	titleIndex map[string]uint32,
	redirects map[string]string,
	pages <-chan *documents.Page,
) <-chan *documents.ArticleIdLinks {
	results := make(chan *documents.ArticleIdLinks, jobs.WorkBuffer)

	linksWg := sync.WaitGroup{}
	for range parallel / 2 {
		linksWg.Add(1)
		go func() {
			var page *documents.Page

			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("panic called while processing %q\n", page.Title)
					fmt.Println("Page text as imported:")
					fmt.Println(page.Text)
					panic(r)
				}
			}()

			for page = range pages {
				tokens := article.Tokenize(article.UnparsedText(page.Text))

				links := &documents.ArticleIdLinks{
					Id: page.Id,
				}

				for _, link := range article.ToLinkTargets(tokens) {
					redirectedTarget, err := documents.GetDestination(redirects, titleIndex, link.Target)
					if err != nil {
						panic(err)
					}

					id, found := titleIndex[redirectedTarget]
					if !found {
						continue
					}

					links.Links = append(links.Links, id)
				}

				results <- links
			}
			linksWg.Done()
		}()
	}

	go func() {
		linksWg.Wait()
		close(results)
	}()

	return results
}
