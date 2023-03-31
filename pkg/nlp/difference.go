package nlp

import (
	"math"
	"sort"
)

type WordBits struct {
	Word string

	Bits float64
}

// CharacteristicWords finds the words that most differentiate two frequency tables.
// These are the words which provide the "most information" in a bayesian sense.
//
// PriorWeight is how strongly we initially believe a random word gives no information.
//
// Negative bits means biased towards left distribution. Positive bits means
// biased towards right distribution.
//
// Results are sorted by absolute value of bits, and alphabetical order is used as a tie-breaker.
func CharacteristicWords(priorWeight float64, left, right *FrequencyTable) []WordBits {
	if left == nil {
		left = &FrequencyTable{}
	}
	if right == nil {
		right = &FrequencyTable{}
	}

	// Alphabetize frequency tables so we can easily compare.
	sort.Slice(left.Words, func(i, j int) bool {
		return left.Words[i].Word < left.Words[j].Word
	})
	sort.Slice(right.Words, func(i, j int) bool {
		return right.Words[i].Word < right.Words[j].Word
	})

	words := make([]WordBits, 0, len(left.Words))

	l := 0
	r := 0

	end := false
	for !end {
		switch {
		case l < len(left.Words) && r < len(right.Words):
			lWord := left.Words[l]
			rWord := right.Words[r]

			switch {
			case lWord.Word == rWord.Word:
				words = append(words,
					ToWordBits(lWord.Word, priorWeight, lWord.Count, rWord.Count),
				)

				l++
				r++
				continue
			case lWord.Word < rWord.Word:
				words = append(words,
					ToWordBits(lWord.Word, priorWeight, lWord.Count, 0),
				)
				l++
			case rWord.Word < lWord.Word:
				words = append(words,
					ToWordBits(rWord.Word, priorWeight, 0, rWord.Count),
				)
				r++
			default:
				panic("Should be impossible")
			}

		case l < len(left.Words):
			lWord := left.Words[l]
			words = append(words,
				ToWordBits(lWord.Word, priorWeight, lWord.Count, 0),
			)
			l++
		case r < len(right.Words):
			rWord := right.Words[r]
			words = append(words,
				ToWordBits(rWord.Word, priorWeight, 0, rWord.Count),
			)
			r++
		default:
			end = true
		}
	}

	sort.Slice(words, func(i, j int) bool {
		if math.Abs(words[i].Bits) != math.Abs(words[j].Bits) {
			return math.Abs(words[i].Bits) > math.Abs(words[j].Bits)
		}
		return words[i].Word < words[j].Word
	})

	return words
}

func ToWordBits(word string, prior float64, left, right uint32) WordBits {
	p := (float64(left) + prior) / (float64(left+right) + 2*prior)
	return WordBits{
		Word: word,
		Bits: BitsFromOdds(p),
	}
}

// BitsFromOdds converts raw odds to bits.
func BitsFromOdds(odds float64) float64 {
	// 1:1 = 0.5 = 0 bytes
	// 1:2 = 0.333 = -1 byte
	// 1:4 = 0.2 = -2 bytes
	// 1:8 = 0.1111 = -3 bytes
	// 2:1 = 0.667 = +1 byte

	return math.Log2(1.0/odds - 1.0)
}

// BitsToOdds converts from bits to raw odds.
func BitsToOdds(bits float64) float64 {
	return 1.0 / (math.Pow(2.0, bits) + 1)
}

func MultiplyOdds(left, right float64) float64 {
	return 1 / (1 + (1/left-1)*(1/right-1))
}
