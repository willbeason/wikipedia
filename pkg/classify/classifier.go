package classify

import (
	"math"
	"sort"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
)

type Classifier interface {
	// Classify returns the sorted, normalized log probabilities of
	// each possible classification.
	Classify(doc *documents.Document) []Result
}

type Bayes struct {
	Model
}

// Model represents log probabilities of a random word in an article of
// a given classification being that particular word.
type Model [][]float64

func (c Bayes) Classify(words []int) []Result {
	// Tracks the likelihood of the passed sequence of words being observed in an
	// article of each classification.
	logProbabilities := make([]Result, len(c.Model))

	for _, word := range words {
		for classification, model := range c.Model {
			logProbabilities[classification].P += model[word]
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

type Result struct {
	Classification
	P float64
}

func Normalize(results []Result) []Result {
	maxLogP := results[0].P

	totalP := 0.0
	for _, result := range results {
		totalP += math.Exp(result.P - maxLogP)
	}

	for i := range results {
		results[i].P = math.Exp(results[i].P-maxLogP) / totalP
	}

	return results
}
