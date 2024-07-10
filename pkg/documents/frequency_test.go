package documents_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/documents"
)

func TestFrequencyTable_ToNgramDictionary(t *testing.T) {
	tcs := []struct {
		name  string
		table documents.FrequencyTable
		want  map[string]bool
	}{
		{
			name:  "empty",
			table: documents.FrequencyTable{},
			want:  map[string]bool{},
		},
		{
			name: "single word",
			table: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the"},
				},
			},
			want: map[string]bool{
				"the": true,
			},
		},
		{
			name: "two words",
			table: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the"},
					{Word: "university"},
				},
			},
			want: map[string]bool{
				"the":        true,
				"university": true,
			},
		},
		{
			name: "bigram",
			table: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the university"},
				},
			},
			want: map[string]bool{
				"the":            true,
				"the university": true,
			},
		},
		{
			name: "tetragram",
			table: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the university of texas"},
				},
			},
			want: map[string]bool{
				"the":                     true,
				"the university":          true,
				"the university of":       true,
				"the university of texas": true,
			},
		},
		{
			name: "multiple ngrams",
			table: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the university of texas"},
					{Word: "fluid dynamics"},
				},
			},
			want: map[string]bool{
				"the":                     true,
				"the university":          true,
				"the university of":       true,
				"the university of texas": true,
				"fluid":                   true,
				"fluid dynamics":          true,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.table.ToNgramDictionary()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestToFrequencyTable(t *testing.T) {
	tcs := []struct {
		name       string
		wordCounts map[string]int
		want       documents.FrequencyTable
	}{
		{
			name:       "nil",
			wordCounts: nil,
			want: documents.FrequencyTable{
				Frequencies: []documents.Frequency{},
			},
		},
		{
			name:       "empty",
			wordCounts: map[string]int{},
			want: documents.FrequencyTable{
				Frequencies: []documents.Frequency{},
			},
		},
		{
			name: "one word",
			wordCounts: map[string]int{
				"the": 100,
			},
			want: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the", Count: 100},
				},
			},
		},
		{
			name: "two words",
			wordCounts: map[string]int{
				"the":        100,
				"university": 10,
			},
			want: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "the", Count: 100},
					{Word: "university", Count: 10},
				},
			},
		},
		{
			name: "four words",
			wordCounts: map[string]int{
				"the":        100,
				"university": 10,
				"of":         100,
				"texas":      20,
			},
			want: documents.FrequencyTable{
				Frequencies: []documents.Frequency{
					{Word: "of", Count: 100},
					{Word: "the", Count: 100},
					{Word: "texas", Count: 20},
					{Word: "university", Count: 10},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := documents.ToFrequencyTable(tc.wordCounts)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
