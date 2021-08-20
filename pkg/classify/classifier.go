package classify

import (
	"math"

	"github.com/willbeason/extract-wikipedia/pkg/ordinality"
)

const (
	BaseCount = 5
)

type Classifier interface {
	// Classify returns the sorted, normalized log probabilities of
	// each possible classification.
	Classify(page *ordinality.PageWordBag) []ClassificationP
}

type ClassificationP struct {
	Classification
	P float64
}

func Normalize(results []ClassificationP) []ClassificationP {
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

type WordBagClassification struct {
	Classification
	*ordinality.PageWordBag
}
