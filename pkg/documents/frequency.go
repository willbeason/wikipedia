package documents

type FrequencyTable struct {
	Frequencies []Frequency
}

type Frequency struct {
	Word  string
	Count int
}

type FrequencyMap struct {
	Counts map[string]int
}

// Collect reads the words in a channel into a frequency table.
func (f *FrequencyMap) Collect(words <-chan string) {
	for word := range words {
		f.Counts[word]++
	}
}

// Filter drops all words which have been seen fewer than minCount times.
func (f *FrequencyMap) Filter(minCount int) {
	for word, count := range f.Counts {
		if count < minCount {
			delete(f.Counts, word)
		}
	}
}
