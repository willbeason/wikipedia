package classify

import (
	"math"
)

// Distribution is a (possibly non-normalized) collection of non-negative probabilities.
type Distribution []float64

func (d Distribution) Normalize() {
	if d == nil || len(d) == 0 {
		return
	}

	sumP := 0.0
	for _, p := range d {
		sumP += p
	}
	if sumP == 0.0 {
		return
	}

	invSumP := 1.0 / sumP
	for i := range d {
		d[i] *= invSumP
	}
}

func (d Distribution) ToLogDistribution() LogDistribution {
	result := make(LogDistribution, len(d))
	for i, p := range d {
		result[i] = math.Log(p)
	}

	return result
}

// LogDistribution is a (possibly non-normalized) collection of log probabilities.
type LogDistribution []float64

// ToDistribution returns an equivalent normalized Distribution.
func (d LogDistribution) ToDistribution() Distribution {
	if d == nil || len(d) == 0 {
		return Distribution{}
	}

	// Offset by the largest probability to prevent math.Exp from underflowing.
	maxLogP := math.Inf(-1)
	for _, logP := range d {
		maxLogP = math.Max(maxLogP, logP)
	}

	result := make(Distribution, len(d))

	for i, logP := range d {
		result[i] = math.Exp(logP - maxLogP)
	}

	result.Normalize()
	return result
}

// Add adds the elements of other to d. The result may not be normalized.
func (d LogDistribution) Add(other LogDistribution) {
	for i, p := range other {
		d[i] += p
	}
}
