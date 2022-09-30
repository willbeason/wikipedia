package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type FrequencyTable struct {
	Frequencies []Frequency
}

func WriteFrequencyTable(out string, t FrequencyTable) error {
	bytes, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(out), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(out, bytes, os.ModePerm)
}

func ReadFrequencyTables(paths ...string) (*FrequencyTable, error) {
	result := &FrequencyTable{}

	for _, path := range paths {
		fmt.Println(path)

		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		frequencyTable := &FrequencyTable{}

		err = yaml.Unmarshal(bytes, frequencyTable)
		if err != nil {
			return nil, err
		}

		result.Frequencies = append(result.Frequencies, frequencyTable.Frequencies...)
	}

	return result, nil
}

func (t *FrequencyTable) ToNgramDictionary() map[string]bool {
	result := make(map[string]bool, len(t.Frequencies))

	for _, f := range t.Frequencies {
		words := strings.Split(f.Word, " ")
		for i := 1; i <= len(words); i++ {
			ngram := strings.Join(words[:i], " ")
			result[ngram] = true
		}
	}

	return result
}

func ToFrequencyTable(wordCounts map[string]int) FrequencyTable {
	frequencies := make([]Frequency, len(wordCounts))
	i := 0

	for word, count := range wordCounts {
		frequencies[i] = Frequency{
			Word:  word,
			Count: count,
		}

		i++
	}

	sort.Slice(frequencies, func(i, j int) bool {
		if frequencies[i].Count != frequencies[j].Count {
			return frequencies[i].Count > frequencies[j].Count
		}
		return frequencies[i].Word < frequencies[j].Word
	})

	// Just the top words.
	return FrequencyTable{Frequencies: frequencies}
}

type Frequency struct {
	Word  string
	Count int
}

type FrequencyMap struct {
	Counts map[string]int
}

func (f *FrequencyMap) CollectMaps(
	wordCountChannel <-chan map[string]int,
	countFilter,
	sizeThreshold int,
) *sync.WaitGroup {
	countsWg := sync.WaitGroup{}
	countsWg.Add(1)

	go func() {
		for wordCounts := range wordCountChannel {
			for word, count := range wordCounts {
				f.Counts[word] += count
			}

			if len(f.Counts) > sizeThreshold {
				f.Filter(countFilter)
				fmt.Println(len(f.Counts))
			}
		}

		f.Filter(countFilter)
		fmt.Println(len(f.Counts))

		countsWg.Done()
	}()

	return &countsWg
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
