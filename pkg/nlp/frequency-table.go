package nlp

import (
	"sort"
)

func ToFrequencyTable(m *FrequencyMap) *FrequencyTable {
	result := &FrequencyTable{
		Words: make([]*WordCount, len(m.Words)),
	}

	i := 0
	for word, count := range m.Words {
		result.Words[i] = &WordCount{
			Word:  word,
			Count: count,
		}
		i++
	}

	result.Sort()
	return result
}

func (x *FrequencyTable) Sort() {
	sort.Slice(x.Words, func(i, j int) bool {
		return x.Words[i].Count > x.Words[j].Count
	})
}
