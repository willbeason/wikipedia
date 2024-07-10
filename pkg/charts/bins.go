package charts

import "math"

// LogarithmicBins creates a sequence of bins for data expected to vary exponentially.
//
// start is the threshold between the first and second bin, and size is the number of
// thresholds.
// factor is the multiplicative factor between each bin threshold.
func LogarithmicBins(start int, end int, factor float64) []int {
	var bins []int

	curSize := float64(start)
	for curSize <= float64(end) {
		bins = append(bins, int(math.Round(curSize)))
		curSize *= factor
	}

	return bins
}
