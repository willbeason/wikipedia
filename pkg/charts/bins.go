package charts

import "math"

// LogarithmicBins creates a sequence of bins for data expected to vary exponentially.
//
// start is the threshold between the first and second bin, and size is the number of
// thresholds.
// factor is the multiplicative factor between each bin threshold.
func LogarithmicBins(start int, size int, factor float64) []int {
	bins := make([]int, size)

	curSize := float64(start)
	for i := range size {
		bins[i] = int(math.Round(curSize))
		curSize *= factor
	}

	return bins
}
