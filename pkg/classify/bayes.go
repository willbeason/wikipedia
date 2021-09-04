package classify

import (
	"math"
	"sort"

	"github.com/willbeason/wikipedia/pkg/ordinality"
)

// Model represents log probabilities of a random word in an article of
// a given classification being that particular word.
//
// The first array is by word index.
// The second is by classification ID.
type Model [][]float64

type Counts [][]uint32

// Ignored is the top n words to ignore in training and analysis.
const Ignored = 55

type Bayes struct {
	Classifications int
	Words           int

	Model
}

func NewCounts(nClassifications, nWords int) Counts {
	counts := make(Counts, nWords+1)

	for i := 0; i <= nWords; i++ {
		newCounts := make([]uint32, nClassifications+1)
		for j := 0; j <= nClassifications; j++ {
			newCounts[j] = BaseCount
		}
		counts[i] = newCounts
	}

	return counts
}

func NewModel(nClassifications, nWords int) Model {
	logProbabilities := make(Model, nWords+1)

	for i := 0; i <= nWords; i++ {
		logProbabilities[i] = make([]float64, nClassifications+1)
	}

	return logProbabilities
}

func TrainBayes(nClassifications, nWords int, wordBags <-chan *WordBagClassification) <-chan *Bayes {
	classificationWords := make([]uint32, nClassifications+1)
	for i := range classificationWords {
		classificationWords[i] = uint32(BaseCount * nWords)
	}

	allWordCounts := NewCounts(nClassifications, nWords)

	result := make(chan *Bayes, 1)
	go func() {
		for wordBag := range wordBags {
			for _, wordCounts := range wordBag.Words {
				allWordCounts[wordCounts.Word][wordBag.Classification] += wordCounts.Count
				classificationWords[wordBag.Classification] += wordCounts.Count
			}
		}

		bayes := &Bayes{
			Classifications: nClassifications,
			Words:           nWords,
			Model:           NewModel(nClassifications, nWords),
		}

		for wordID, wordCounts := range allWordCounts {
			if wordID < Ignored {
				continue
			}

			for classificationID, count := range wordCounts {
				bayes.Model[wordID][classificationID] = math.Log(float64(count) / float64(classificationWords[classificationID]))
			}
		}

		for _, wordLogProbabilities := range bayes.Model {
			maxLogProbability := math.Inf(-1)
			for _, logP := range wordLogProbabilities {
				maxLogProbability = math.Max(maxLogProbability, logP)
			}

			for i := range wordLogProbabilities {
				wordLogProbabilities[i] -= maxLogProbability
			}
		}
		result <- bayes
	}()

	return result
}

func (c *Bayes) Classify(page *ordinality.PageWordBag) []ClassificationP {
	// Tracks the likelihood of the passed sequence of words being observed in an
	// article of each classification.
	logProbabilities := make([]ClassificationP, c.Classifications+1)
	logProbabilities[0].P = math.Inf(-1)

	for i := range logProbabilities {
		logProbabilities[i].Classification = Classification(i)
	}

	for _, wordCount := range page.Words {
		for classification, logP := range c.Model[wordCount.Word] {
			logProbabilities[classification].P += logP * float64(wordCount.Count)
		}
	}

	// The classification with the highest log probability has the highest chance
	// of being the true classification of the article, so sort from most likely
	// to least likely.
	sort.Slice(logProbabilities, func(i, j int) bool {
		return logProbabilities[i].P > logProbabilities[j].P
	})

	// In practice, we rarely need more than the first three.
	return Normalize(logProbabilities)
}
