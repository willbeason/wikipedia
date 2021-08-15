package classify

import (
	"sort"
)

// PClassification is a Classification and its probability.
type PClassification struct {
	P float64
	Classification
}

type PClassifications []PClassification

// Sort sorts r from the highest probability to the lowest probability.
func (r PClassifications) Sort() {
	sort.Slice(r, func(i, j int) bool {
		return r[i].P > r[j].P
	})
}

// ToPClassifications converts a distribution to the respective Classifications and probabilities
// it represents.
func ToPClassifications(d Distribution) PClassifications {
	result := make(PClassifications, len(d))
	for i, p := range d {
		result[i] = PClassification{P: p, Classification: Classification(i)}
	}

	return result
}
