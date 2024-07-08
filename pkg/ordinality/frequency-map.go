package ordinality

import (
	"fmt"
)

type WordCollector struct {
	// CountFilter
	CountFilter uint32

	// SizeThreshold is the
	SizeThreshold int

	Counts map[string]uint32
}

func (wc *WordCollector) Add(counts map[string]uint32) {
	if len(wc.Counts) > wc.SizeThreshold {
		wc.FilterCounts()
	}
	for ngram, count := range counts {
		wc.Counts[ngram] += count
	}
}

func (wc *WordCollector) FilterCounts() {
	for k, v := range wc.Counts {
		if v < wc.CountFilter {
			delete(wc.Counts, k)
		}
	}

	newLen := len(wc.Counts)
	fmt.Println("Reduced to", newLen, "ngrams")

	if newLen >= wc.SizeThreshold {
		panic("unable to reduce to below size threshold")
	}
}
