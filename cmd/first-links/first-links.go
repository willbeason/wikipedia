package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/environment"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	"github.com/willbeason/wikipedia/pkg/pages"
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `first-links`,
		Short: `Analyzes the network of references between biographical articles.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	parallel, err := flags.GetParallel(cmd)
	if err != nil {
		return err
	}

	inDB := filepath.Join(environment.WikiPath, "extracted.db")

	source := pages.StreamDB(inDB, parallel)

	ctx, cancel := context.WithCancelCause(cmd.Context())

	idMap := make(map[string]uint32)
	titleMap := make(map[uint32]string)

	biographies := make(map[uint32]bool)

	resultMtx := sync.Mutex{}

	docs, err := source(ctx, cancel)
	if err != nil {
		return err
	}

	checker, err := documents.NewInfoboxChecker(documents.PersonInfoboxes())
	if err != nil {
		return err
	}

	idMapWork := jobs.Reduce(jobs.WorkBuffer, docs, func(page *documents.Page) error {
		if !checker.Matches(page.Text) {
			resultMtx.Lock()
			idMap[page.Title] = page.Id
			titleMap[page.Id] = page.Title
			resultMtx.Unlock()
			return nil
		}

		resultMtx.Lock()
		idMap[page.Title] = page.Id
		titleMap[page.Id] = page.Title

		biographies[page.Id] = true

		resultMtx.Unlock()

		return nil
	})

	runner := jobs.NewRunner()
	idMapWg := runner.Run(ctx, cancel, idMapWork)
	// Must fully wait for ID Map to be created and the Badger database closed before opening another connection.
	idMapWg.Wait()

	fmt.Println("Articles:", len(idMap))

	// network is the map from all articles to the articles they link to.
	network := make(map[uint32][]uint32)
	networkMtx := sync.Mutex{}

	source2 := pages.StreamDB(inDB, parallel)
	docs2, err := source2(ctx, cancel)
	if err != nil {
		return err
	}

	networkWork := jobs.Reduce(jobs.WorkBuffer, docs2, func(page *documents.Page) error {
		from, foundFrom := idMap[page.Title]
		if !foundFrom {
			return fmt.Errorf("%w: did not add ID for %q", jobs.ErrStream, page.Title)
		}

		// Find first X links in article.
		numLinks := 20
		matches := linkRegex.FindAllString(page.Text, numLinks)

		// Force-exclude self reference.
		var tos []uint32
		seen := map[uint32]bool{from: true}

		for _, match := range matches {
			// Strip square brackets.
			match = match[2 : len(match)-2]

			// Only consider before vertical bar.
			if idx := strings.Index(match, "|"); idx != -1 {
				match = match[:idx]
			}

			// Ignore references to non-biographies.
			to, foundTo := idMap[match]
			if !foundTo {
				continue
			}

			// Don't add duplicates.
			if seen[to] {
				continue
			}
			seen[to] = true

			tos = append(tos, to)
		}

		networkMtx.Lock()
		network[from] = tos
		networkMtx.Unlock()

		return nil
	})

	networkWg := runner.Run(ctx, cancel, networkWork)
	networkWg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	fmt.Println("Nodes:", len(network))

	topicsMap := make(map[uint32]int)

	for id, links := range network {
		if !biographies[id] {
			// Not a biography, so don't count for calculating topics.
			continue
		}

		seen := make(map[uint32]bool)
		// Disallow self-links.
		seen[id] = true

		// Neighborhood 1.
		for _, link := range links {
			if seen[link] {
				continue
			}
			seen[link] = true
			topicsMap[link]++

			// Neighborhood 2.
			for _, subLink := range network[link] {
				if seen[subLink] {
					continue
				}
				seen[subLink] = true
				topicsMap[subLink]++
			}
		}
	}

	topics := make([]Topic, 0, len(topicsMap))
	for id, topicCount := range topicsMap {
		topics = append(topics, Topic{
			Title: titleMap[id],
			Count: topicCount,
		})
	}

	fmt.Printf("%d Topics\n\n", len(topics))

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Count > topics[j].Count
	})

	for i, topic := range topics[:1000] {
		fmt.Printf("%4d. %-30s:%7d\n", i, topic.Title, topic.Count)
	}

	return nil
}

type Topic struct {
	Title string
	Count int
}

var linkRegex = regexp.MustCompile(`\[\[[^]]+]]`)
