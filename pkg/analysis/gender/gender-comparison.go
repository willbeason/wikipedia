package gender

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/flags"
	"github.com/willbeason/wikipedia/pkg/jobs"
	progress_bar "github.com/willbeason/wikipedia/pkg/progress-bar"
	"github.com/willbeason/wikipedia/pkg/protos"
)

var ErrGenderComparison = errors.New("running gender comparison")

func Comparison(cmd *cobra.Command, cfg *config.GenderComparison, corpusNames ...string) error {
	if len(corpusNames) != 2 {
		return fmt.Errorf("%w: must have exactly two corpora but got %+v", ErrGenderComparison, corpusNames)
	}
	beforeCorpusName := corpusNames[0]

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancelCause(cmd.Context())

	errs := make(chan error)
	go func() {
		for err := range errs {
			cancel(err)
		}
	}()

	beforeGenderPath := filepath.Join(workspace, beforeCorpusName, cfg.GenderIndex)
	beforeGenderSource := jobs.NewSource(protos.ReadFile[documents.ArticleIdGender](beforeGenderPath))
	beforeGenderWg, beforeGenderJob, beforeGenderProtos := beforeGenderSource()
	go beforeGenderJob(ctx, errs)

	beforeGender := make(map[uint32]string)
	for p := range beforeGenderProtos {
		beforeGender[p.Id] = p.Gender
	}
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	beforeLinksPath := filepath.Join(workspace, beforeCorpusName, cfg.Links)
	beforeLinksSource := jobs.NewSource(protos.ReadFile[documents.ArticleIdLinks](beforeLinksPath))
	beforeLinksWg, beforeLinksJob, beforeLinksProtos := beforeLinksSource()
	go beforeLinksJob(ctx, errs)

	beforeLinks := make(map[uint32]*documents.ArticleIdLinks)
	for p := range beforeLinksProtos {
		beforeLinks[p.Id] = p
	}

	beforeGenderWg.Wait()
	beforeLinksWg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	beforeGenderCountsMap := make(map[string]int)
	beforeGenderLinkCountsMap := make(map[string]int)
	beforeSectionCountsMap := make(map[string]int)
	genderSectionLinkCounts := make(map[string]map[string]int)
	totalGenderedLinked := 0
	genderLinkTargetsMap := make(map[string]int)

	fmt.Println("Sections not implemented - reimplement.")

	articlesProgress := progress_bar.NewProgressBar("Articles", int64(len(beforeLinks)), os.Stdout)
	articlesProgress.Start()
	n := 0
	for id, links := range beforeLinks {
		n++
		articlesProgress.Increment()

		gender, hasGender := beforeGender[id]

		sectionLinkCounts, exists := genderSectionLinkCounts[gender]
		if !exists {
			sectionLinkCounts = make(map[string]int)
		}

		if len(links.Links) > 0 {
			if hasGender {
				beforeGenderCountsMap[gender]++
				totalGenderedLinked++
			}
		}

		for _, link := range links.Links {
			targetGender, targetHasGender := beforeGender[link]
			if targetHasGender {
				genderLinkTargetsMap[targetGender]++
			}
		}
		genderSectionLinkCounts[gender] = sectionLinkCounts
	}
	articlesProgress.Stop()

	beforeGenderCounts := make([]StringCount, 0, len(beforeGenderCountsMap))
	for gender, count := range beforeGenderCountsMap {
		beforeGenderCounts = append(beforeGenderCounts, StringCount{
			String: gender,
			Count:  count,
		})
	}
	sort.Slice(beforeGenderCounts, func(i, j int) bool {
		return beforeGenderCounts[i].Count > beforeGenderCounts[j].Count
	})
	for rank, beforeGenderCount := range beforeGenderCounts {
		fmt.Printf("%d,%s,%d,%d,%d\n",
			rank, beforeGenderCount.String, beforeGenderCount.Count,
			beforeGenderLinkCountsMap[beforeGenderCount.String], genderLinkTargetsMap[beforeGenderCount.String])
	}

	beforeSectionCounts := make([]StringCount, 0, len(beforeSectionCountsMap))
	for gender, count := range beforeSectionCountsMap {
		beforeSectionCounts = append(beforeSectionCounts, StringCount{
			String: gender,
			Count:  count,
		})
	}
	sort.Slice(beforeSectionCounts, func(i, j int) bool {
		return beforeSectionCounts[i].Count > beforeSectionCounts[j].Count
	})

	for _, section := range beforeSectionCounts[:20] {
		fmt.Printf("%s,", section.String)
	}
	fmt.Println()

	for _, gender := range beforeGenderCounts {
		fmt.Printf("%s,", gender.String)
		totalLinks := 0
		sectionLinkCounts := genderSectionLinkCounts[gender.String]
		for _, count := range sectionLinkCounts {
			totalLinks += count
		}

		for _, section := range beforeSectionCounts[:20] {
			sectionCount := sectionLinkCounts[section.String]
			percent := 100.0 * float64(sectionCount) / float64(totalLinks)
			fmt.Printf("%6.02f%%,", percent)
		}
		fmt.Println()
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

type StringCount struct {
	String string
	Count  int
}
