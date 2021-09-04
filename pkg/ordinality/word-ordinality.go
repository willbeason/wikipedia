package ordinality

import (
	"github.com/willbeason/wikipedia/pkg/nlp"
)

// WordOrdinality is a map from a word (or n-gram) to its ordinal position in a
// dictionary. The positions are 1-based so 0 (the null value) can be used to
// represent words not in the dictionary.
type WordOrdinality map[string]uint32

func NewWordOrdinality(dictionary *nlp.Dictionary) WordOrdinality {
	if dictionary == nil {
		return nil
	}
	result := make(WordOrdinality, len(dictionary.Words))

	for i, word := range dictionary.Words {
		result[word] = uint32(i) + 1
	}

	return result
}
