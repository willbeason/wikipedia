package nlp

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNgramTokenizer_Tokenize(t *testing.T) {
	testCases := []struct {
		name       string
		s          string
		dictionary map[string]bool
		want       []string
	}{
		{
			name:       "empty string",
			s:          "",
			dictionary: map[string]bool{},
			want:       nil,
		},
		{
			name:       "words",
			s:          "hello world",
			dictionary: map[string]bool{},
			want:       []string{"hello", "world"},
		},
		{
			name: "bigram",
			s:    "hello world",
			dictionary: map[string]bool{
				"hello world": true,
			},
			want: []string{"hello world"},
		},
		{
			name: "ambiguous bigram",
			s:    "hello world peace",
			dictionary: map[string]bool{
				"hello world": true,
				"world peace": true,
			},
			want: []string{"hello world", "peace"},
		},
		{
			name: "ambiguous bigram 2",
			s:    "big big big",
			dictionary: map[string]bool{
				"big big": true,
			},
			want: []string{"big big", "big"},
		},
		{
			name: "trigram",
			s:    "university of texas",
			dictionary: map[string]bool{
				"university of":       true,
				"university of texas": true,
			},
			want: []string{"university of texas"},
		},
		{
			name: "sentence",
			s:    "she got a degree in advanced pharmacology from the university of texas",
			dictionary: map[string]bool{
				"got a":                 true,
				"degree in":             true,
				"advanced pharmacology": true,
				"from the":              true,
				"university of":         true,
				"of texas":              true,
				"university of texas":   true,
			},
			want: []string{"she", "got a", "degree in", "advanced pharmacology", "from the", "university of texas"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenizer := NgramTokenizer{
				Underlying: WordTokenizer{},
				Dictionary: tc.dictionary,
			}

			got := tokenizer.Tokenize(tc.s)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
