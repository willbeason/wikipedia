package nlp

import (
	"math"
	"sort"
)

type Prediction int

const (
	Female = 0
	Male   = 1
)

type WordAccuracy struct {
	Word       string
	Prediction Prediction
	Accuracy   float64
}

func MostAccurateWords(left, right *FrequencyTable, totalMen, totalWomen uint32) []WordAccuracy {
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

	words := make([]WordAccuracy, 0, len(left.Words))

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
					ToWordAccuracy(lWord.Word, lWord.Count, rWord.Count, totalMen, totalWomen),
				)

				l++
				r++
				continue
			case lWord.Word < rWord.Word:
				words = append(words,
					ToWordAccuracy(lWord.Word, lWord.Count, 0, totalMen, totalWomen),
				)
				l++
			case rWord.Word < lWord.Word:
				words = append(words,
					ToWordAccuracy(rWord.Word, 0, rWord.Count, totalMen, totalWomen),
				)
				r++
			default:
				panic("Should be impossible")
			}

		case l < len(left.Words):
			lWord := left.Words[l]
			words = append(words,
				ToWordAccuracy(lWord.Word, lWord.Count, 0, totalMen, totalWomen),
			)
			l++
		case r < len(right.Words):
			rWord := right.Words[r]
			words = append(words,
				ToWordAccuracy(rWord.Word, 0, rWord.Count, totalMen, totalWomen),
			)
			r++
		default:
			end = true
		}
	}

	sort.Slice(words, func(i, j int) bool {
		if math.Abs(words[i].Accuracy) != math.Abs(words[j].Accuracy) {
			return math.Abs(words[i].Accuracy) > math.Abs(words[j].Accuracy)
		}
		return words[i].Word < words[j].Word
	})

	return words
}

func ToWordAccuracy(word string, men, women, totalMen, totalWomen uint32) WordAccuracy {
	f1 := float64(women) / float64(totalWomen)
	f0 := 1.0 - f1

	m1 := float64(men) / float64(totalMen)
	m0 := 1.0 - m1

	af := (f1 + m0) / 2.0
	am := (f0 + m1) / 2.0

	if af > am {
		// Word more predictive of an article about a woman.
		return WordAccuracy{
			Word:       word,
			Prediction: Female,
			Accuracy:   af,
		}
	}
	// Word more predictive of an article about a man.
	return WordAccuracy{
		Word:       word,
		Prediction: Male,
		Accuracy:   am,
	}
}
