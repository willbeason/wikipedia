package nlp

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

const velocity = 1.0

type Perceptron struct {
	// Dictionary is the ordered list of words used in a model.
	// The index of a word is its identifier.
	Dictionary map[string]int

	// Weights is the
	Weights []float64

	// Bias is how much to weight examples which are labeled "female" to adjust for
	// inequitable representation.
	Bias float64
}

type WordWeight struct {
	Word   string
	Weight float64
}

func (p *Perceptron) WordOrder() []WordWeight {
	wws := make([]WordWeight, len(p.Weights))
	for word, i := range p.Dictionary {
		wws[i] = WordWeight{
			Word:   word,
			Weight: p.Weights[i],
		}
	}

	sort.Slice(wws, func(i, j int) bool {
		return math.Abs(wws[i].Weight) > math.Abs(wws[j].Weight)
	})

	return wws
}

func NewPerceptron(dictionary []string, bias float64) *Perceptron {
	result := &Perceptron{
		Dictionary: make(map[string]int, len(dictionary)),
		Weights:    make([]float64, len(dictionary)),
		Bias:       bias,
	}

	// Full copy to ensure mutating input slice doesn't impact Perceptron.
	for i, word := range dictionary {
		result.Dictionary[word] = i
	}

	// Randomize initial weights.
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	for i := range dictionary {
		result.Weights[i] = r.NormFloat64()
	}

	return result
}

func (p *Perceptron) ToDatum(words []string, label Prediction) Datum {
	result := Datum{
		Label: label,
	}

	seen := make(map[string]bool)

	for _, word := range words {
		if seen[word] {
			continue
		}
		seen[word] = true

		idx, found := p.Dictionary[word]
		if found {
			result.Words = append(result.Words, idx)
		}
	}
	sort.Ints(result.Words)

	return result
}

type Datum struct {
	Words []int
	Label Prediction
}

// Train makes a single training run.
func (p *Perceptron) Train(data []Datum) (*Perceptron, float64, float64, bool) {
	beforeCost := p.WeightsCost()

	// Calculate the desired direction of change for each weight.
	// Initial pressure to regress to zero.
	deltas := p.WeightsDerivatives()
	//fmt.Println(deltas[:10])

	correct := 0.0
	total := 0.0

	beforeIncorrectCost := 0.0
	for _, d := range data {
		score := 0.0

		for _, w := range d.Words {
			score += p.Weights[w]
		}

		switch d.Label {
		// We aren't sensitive to changes where the model is already sure about
		// a designation.
		case Female:
			total += p.Bias
			if score > 0.0 {
				correct += p.Bias
			}
			if score < 1.0 {
				beforeIncorrectCost += (1.0 - score) * p.Bias

				// We want the score to be higher, so increasing the score decreases the cost.
				for _, w := range d.Words {
					deltas[w] += 1.0 * p.Bias
					//if w == 0 {
					//	fmt.Println("Female", deltas[0])
					//}
				}
			}
		case Male:
			total += 1.0
			if score < 0.0 {
				correct += 1.0
			}
			if score > -1.0 {
				beforeIncorrectCost += 1.0 + score

				// We want the score to be lower, so decreasing the score decreases the cost.
				for _, w := range d.Words {
					deltas[w] -= 1.0
					//if w == 0 {
					//	fmt.Println("Male", deltas[0])
					//}
				}
			}
		}
	}
	//fmt.Println(beforeCost, beforeIncorrectCost)
	beforeCost += beforeIncorrectCost

	// Normalize to a length of one.
	//for i, d := range deltas {
	//	fmt.Println(i, d)
	//}
	//normalize(deltas)
	//fmt.Println(deltas[:10])
	for i := range deltas {
		deltas[i] *= 0.0001
	}

	next := &Perceptron{
		Dictionary: p.Dictionary,
		Weights:    make([]float64, len(p.Weights)),
		Bias:       p.Bias,
	}

	copy(next.Weights, p.Weights)
	for i, w := range deltas {
		next.Weights[i] += w
	}

	afterCost := next.WeightsCost()

	afterIncorrectCost := 0.0
	for _, d := range data {
		score := 0.0

		for _, w := range d.Words {
			score += next.Weights[w]
		}

		switch d.Label {
		// We aren't sensitive to changes where the model is already sure about
		// a designation.
		case Female:
			if score < 1.0 {
				afterIncorrectCost += (1.0 - score) * p.Bias
			}
		case Male:
			if score > -1.0 {
				afterIncorrectCost += 1.0 + score
			}
		}
	}
	afterCost += afterIncorrectCost

	//fmt.Println(beforeCost, "=>", afterCost)

	accuracy := correct / total

	if afterCost < beforeCost {
		return next, afterCost, accuracy, true
	}

	return p, beforeCost, accuracy, false
}

func (p *Perceptron) Predict(datum Datum) float64 {
	score := 0.0
	for _, w := range datum.Words {
		score += p.Weights[w]
	}

	return score
}

// WeightsCost ensures the model isn't more complex than it really needs to be.
func (p *Perceptron) WeightsCost() float64 {
	cost := 0.0
	for _, weight := range p.Weights {
		cost += (weight*weight + math.Abs(weight)) * 20.0
		//fmt.Println(i, cost)
	}

	return cost
}

// WeightsDerivatives is the incremental change in weight cost of modifying the
// value of each weight by an arbitrarily small delta.
func (p *Perceptron) WeightsDerivatives() []float64 {
	result := make([]float64, len(p.Weights))

	for i, w := range p.Weights {
		dw := 2.0 * w
		if w > 0.0 {
			dw += 1.0
		} else if w < 0.0 {
			dw -= 1.0
		}

		dw *= 20.0 // factor
		// Force pointing towards zero.
		result[i] = dw * -1.0
	}
	//fmt.Println("Weights", result[:20])
	//fmt.Println("Derivatives", result[:20])

	return result
}

func normalize(vector []float64) {
	d2 := 0.0
	for _, w := range vector {
		d2 += w * w
	}
	//fmt.Println("D2", d2)

	invD := 1.0 / math.Sqrt(d2)

	for i, v := range vector {
		vector[i] = v * invD
	}
}
