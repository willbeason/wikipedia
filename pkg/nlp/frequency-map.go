package nlp

import (
	"fmt"
)

func CollectWordCounts(wordCountsChannel <-chan map[string]uint32, known map[string]bool, countFilter uint32, sizeThreshold int, minCount uint32) <-chan *FrequencyMap {
	result := make(chan *FrequencyMap, 100)

	go func() {
		counts := collectWordCounts(wordCountsChannel, known, countFilter, sizeThreshold)

		FilterCounts(counts.Words, minCount)

		result <- counts
		close(result)
	}()

	return result
}

func collectWordCounts(wordCountsChannel <-chan map[string]uint32, known map[string]bool, countFilter uint32, sizeThreshold int) *FrequencyMap {
	knownCounts := &FrequencyMap{
		Words: make(map[string]uint32, len(known)),
	}
	unknownCounts := &FrequencyMap{
		Words: make(map[string]uint32),
	}

	for wordCounts := range wordCountsChannel {
		for word, count := range wordCounts {
			if known[word] {
				knownCounts.Words[word] += count
			} else {
				unknownCounts.Words[word] += count
			}
		}

		if len(unknownCounts.Words) > sizeThreshold {
			FilterCounts(unknownCounts.Words, countFilter)
		}
	}

	result := knownCounts
	for newWord, count := range unknownCounts.Words {
		result.Words[newWord] = count
	}

	return result
}

func FilterCounts(m map[string]uint32, countFilter uint32) {
	for k, v := range m {
		if v < countFilter {
			delete(m, k)
		}
	}
	fmt.Println(len(m))
}
