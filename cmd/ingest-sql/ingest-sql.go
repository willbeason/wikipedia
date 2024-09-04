package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
)

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}


func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `ingest-sql file`,
		Short: `Runs extraction of SQL links`,
		RunE:  runE,
	}

	flags.Workspace(cmd)

	return cmd
}

func runE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	workspacePath, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	corpus := "20240701"
	titlePath := "title-index.pb"
	titlePath = filepath.Join(workspacePath, corpus, titlePath)

	redirectsPath := "redirects.pb"
	redirectsPath = filepath.Join(workspacePath, corpus, redirectsPath)

	ctx := cmd.Context()
	ctx, cancel := context.WithCancelCause(ctx)

	errs := make(chan error, 1)
	errsWg := sync.WaitGroup{}
	errsWg.Add(1)
	go func() {
		for err := range errs {
			cancel(err)
		}
		errsWg.Done()
	}()

	titleIndexF := documents.ReadTitleMap(ctx, titlePath, errs)
	titleIndexInverseF := documents.ReadReverseTitleMap(ctx, titlePath, errs)
	redirectsF := documents.MakeRedirects(ctx, redirectsPath, errs)
	titleIndex := <-titleIndexF
	titleIndexReverse := <-titleIndexInverseF
	redirects := <-redirectsF
	fmt.Println("Read files")

	redirectsIds := make(map[uint32]uint32)
	for source, target := range redirects {
		sourceId, hasSource := titleIndex[source]
		if !hasSource {
			continue
		}
		targetId, hasTarget := titleIndex[target]
		if !hasTarget {
			continue
		}
		redirectsIds[sourceId] = targetId
	}
	fmt.Println("Formatted redirects")

	sqlPath := "/home/will/Downloads/enwiki-20240901-pagelinks.sql"
	sqlFile, err := os.Open(sqlPath)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(sqlFile)
	// Get to insert line.
	fmt.Println("Finding insert line")
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if s == "/*!40000 ALTER TABLE `pagelinks` DISABLE KEYS */;\n" {
			break
		}
	}

	// Get to first entry.
	fmt.Println("Going to first entry")
	discard := "INSERT INTO `pagelinks` VALUES ("
	discardBytes := make([]byte, len(discard))
	_, err = reader.Read(discardBytes)
	if err != nil {
		return err
	}

	entryPattern := regexp.MustCompile(`(\d+),(\d+),(\d+)`)

	totalLinks := 0

	fmt.Println("Reading entries")
	for {
		entry, err := reader.ReadString('(')
		if err != nil {
			return err
		}

		matches := entryPattern.FindStringSubmatch(entry)
		if matches[2] != "0" {
			fmt.Println("Found last article link", matches)
			break
		}

		target, err := strconv.ParseUint(matches[3], 10, 32)
		if err != nil {
			return err
		}
		_, targetExists := titleIndexReverse[uint32(target)]
		if !targetExists {
			continue
		}

		_, err = strconv.ParseUint(matches[1], 10, 32)
		if err != nil {
			return err
		}

		totalLinks++
		if totalLinks % 1000000 == 0 {
			fmt.Println(totalLinks, matches[1], matches[3])
		}
	}
	fmt.Println(totalLinks)


	close(errs)
	errsWg.Wait()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}