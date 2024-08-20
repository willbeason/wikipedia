package gender

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/willbeason/wikipedia/pkg/documents"
	"github.com/willbeason/wikipedia/pkg/protos"

	"github.com/spf13/cobra"
	"github.com/willbeason/wikipedia/pkg/config"
	"github.com/willbeason/wikipedia/pkg/flags"
)

var ErrGenderComparison = errors.New("running gender comparison")

func Comparison(cmd *cobra.Command, cfg *config.GenderComparison, corpusNames ...string) error {
	if len(corpusNames) != 2 {
		return fmt.Errorf("%w: must have exactly two corpora but got %+v", ErrGenderComparison, corpusNames)
	}
	beforeCorpusName := corpusNames[0]
	afterCorpusName := corpusNames[1]

	workspace, err := flags.GetWorkspacePath(cmd)
	if err != nil {
		return err
	}

	beforeGenderPath := filepath.Join(workspace, beforeCorpusName, cfg.GenderIndex)
	beforeGender, err := protos.Read[documents.GenderIndex](beforeGenderPath)
	if err != nil {
		return err
	}

	afterGenderPath := filepath.Join(workspace, afterCorpusName, cfg.GenderIndex)
	_, err = protos.Read[documents.GenderIndex](afterGenderPath)
	if err != nil {
		return err
	}

	beforeLinksPath := filepath.Join(workspace, beforeCorpusName, cfg.Links)
	beforeLinks, err := protos.Read[documents.LinkIndex](beforeLinksPath)
	if err != nil {
		return err
	}

	afterLinksPath := filepath.Join(workspace, afterCorpusName, cfg.Links)
	_, err = protos.Read[documents.LinkIndex](afterLinksPath)
	if err != nil {
		return err
	}

	beforeGenderCountsMap := make(map[string]int)
	beforeGenderLinkCountsMap := make(map[string]int)
	beforeSectionCountsMap := make(map[string]int)
	genderSectionLinkCounts := make(map[string]map[string]int)
	totalGenderedLinked := 0
	genderLinkTargetsMap := make(map[string]int)

	for id, links := range beforeLinks.Articles {
		gender, hasGender := beforeGender.Genders[id]

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
			if hasGender {
				section := strings.TrimSpace(link.Section)
				beforeSectionCountsMap[section]++
				sectionLinkCounts[section]++
				beforeGenderLinkCountsMap[gender]++
			}

			targetGender, targetHasGender := beforeGender.Genders[link.Target]
			if targetHasGender {
				genderLinkTargetsMap[targetGender]++
			}
		}
		genderSectionLinkCounts[gender] = sectionLinkCounts
	}

	var beforeGenderCounts []StringCount
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
		fmt.Printf("%d,%s,%d,%d,%d\n", rank, beforeGenderCount.String, beforeGenderCount.Count, beforeGenderLinkCountsMap[beforeGenderCount.String], genderLinkTargetsMap[beforeGenderCount.String])
	}

	var beforeSectionCounts []StringCount
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

	return nil
}

type StringCount struct {
	String string
	Count  int
}
