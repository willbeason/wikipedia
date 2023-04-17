package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/flags"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"os"
	"sort"
	"strings"
)

func mainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(3),
		Use:   `decision-tree path/to/bags path/to/labels path/to/dictionary`,
		Short: `Convert articles to easily-processable word bags.`,
		RunE:  runCmd,
	}

	flags.Parallel(cmd)
	flags.IDs(cmd)

	return cmd
}

func main() {
	err := mainCmd().Execute()
	if err != nil {
		os.Exit(1)
	}

}

func runCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	wordBags, err := documents.ReadWordSets(args[0])
	if err != nil {
		return err
	}

	labels, err := nlp.ReadDocumentGenders(args[1])
	if err != nil {
		return err
	}

	femaleLabels := 0
	maleLabels := 0

	for _, l := range labels {
		switch l.Gender {
		case nlp.Female:
			femaleLabels++
		case nlp.Male:
			maleLabels++
		}
	}

	disparity := float64(maleLabels) / float64(femaleLabels)
	fmt.Printf("Disparity: %.02f\n", disparity)

	labelsMap := make(map[uint32]nlp.Gender)
	for _, l := range labels {
		labelsMap[l.ID] = l.Gender
	}

	dictionary, err := nlp.ReadDictionary(args[2])
	if err != nil {
		return err
	}

	labeledSets := make([]LabeledSet, len(wordBags))

	for i, ws := range wordBags {
		labeledSets[i].Label = labelsMap[ws.ID]
		labeledSets[i].Words = ws.ToBits(len(dictionary.Words))
	}

	findBranches(0, disparity, dictionary, labeledSets)

	return nil
}

type LabeledSet struct {
	Label nlp.Gender
	Words []bool
}

func findBranches(depth int, disparity float64, dictionary *nlp.Dictionary, toPartition []LabeledSet) {
	padding := strings.Repeat("  ", depth)

	bestWords := findDecider(disparity, len(dictionary.Words), toPartition)

	nextWord := bestWords[0]

	//nw := dictionary.Words[nextWord.Word]
	fmt.Printf("%s%d Documents\n", padding, len(toPartition))
	//fmt.Printf("%s%s (%d) => %s | %.02f%%\n", padding, nw, nextWord.Word, nextWord.Gender, nextWord.Accuracy*100)

	for _, w := range bestWords {
		fmt.Printf("%s,%d,%s,%.02f%%\n",
			dictionary.Words[w.Word],
			w.Word,
			w.Gender,
			w.Accuracy*100,
		)
	}

	var trueSet []LabeledSet
	var falseSet []LabeledSet

	for _, ls := range toPartition {
		if ls.Words[nextWord.Word] {
			trueSet = append(trueSet, ls)
		} else {
			falseSet = append(falseSet, ls)
		}
	}

	if nextWord.Accuracy > 0.999 {
		fmt.Printf("%s--- %.02f%%\n", padding, nextWord.Accuracy*100)
		return
	}

	fmt.Println(padding + "True:")
	if depth < 0 {
		findBranches(depth+1, disparity, dictionary, trueSet)
	}

	fmt.Println(padding + "False:")

	if depth < 0 {
		findBranches(depth+1, disparity, dictionary, falseSet)
	}
}

var forbiddenWords = map[int]bool{
	//55796: true, // she
	//29061: true, // her
	//29401: true, // himself
	//29158: true, // herself
	//28621: true, // he
	//29470: true, // his
	//29395: true, // him
	//844:   true, // actress
	//69168: true, // woman
}

type WordAccuracy struct {
	Word     uint32
	Gender   nlp.Gender
	Accuracy float64
}

// findDecider determines the next node appropriate for a decision tree to partition
// a set of articles about men and women into sets.
func findDecider(disparity float64, dictionarySize int, labeledSets []LabeledSet) []WordAccuracy {
	total := 0.0

	for _, ls := range labeledSets {
		switch ls.Label {
		case nlp.Female:
			total += disparity
		case nlp.Male:
			total += 1.0
		}
	}

	wordAccuracies := make([]WordAccuracy, dictionarySize)

	for n := 0; n < dictionarySize; n++ {
		if forbiddenWords[n] {
			continue
		}

		female1 := 0.0
		female0 := 0.0
		male1 := 0.0
		male0 := 0.0

		for _, ws := range labeledSets {
			found := ws.Words[uint32(n)]

			switch ws.Label {
			case nlp.Female:
				if found {
					female1 += disparity
				} else {
					female0 += disparity
				}
			case nlp.Male:
				if found {
					male1 += 1.0
				} else {
					male0 += 1.0
				}
			}

		}

		scoreFemale := female1 + male0
		scoreMale := male1 + female0

		gender := nlp.Unknown
		if scoreFemale > scoreMale {
			gender = nlp.Female
		} else if scoreMale > scoreFemale {
			gender = nlp.Male
		}

		score := 0.0
		if scoreFemale > scoreMale {
			score = scoreFemale
		} else if scoreMale > scoreFemale {
			score = scoreMale
		}

		wordAccuracies[n] = WordAccuracy{
			Word:     uint32(n),
			Gender:   gender,
			Accuracy: score / total,
		}
	}

	sort.Slice(wordAccuracies, func(i, j int) bool {
		if wordAccuracies[i] != wordAccuracies[j] {
			return wordAccuracies[i].Accuracy > wordAccuracies[j].Accuracy
		}

		return wordAccuracies[i].Word < wordAccuracies[j].Word
	})

	return wordAccuracies[:50]
}
