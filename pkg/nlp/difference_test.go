package nlp_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
	"testing"
)

func TestCharacteristicWords(t *testing.T) {
	tcs := []struct {
		name  string
		prior float64
		left  *nlp.FrequencyTable
		right *nlp.FrequencyTable
		want  []nlp.WordBits
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name:  "only left",
			prior: 5.0,
			left: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{{
					Word:  "apple",
					Count: 5,
				}},
			},
			want: []nlp.WordBits{{
				Word: "apple",
				Bits: -1.0,
			}},
		},
		{
			name:  "only right",
			prior: 5.0,
			right: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{{
					Word:  "banana",
					Count: 15,
				}},
			},
			want: []nlp.WordBits{{
				Word: "banana",
				Bits: 2.0,
			}},
		},
		{
			name:  "alternate prior",
			prior: 3.0,
			right: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{{
					Word:  "cantaloupe",
					Count: 9,
				}},
			},
			want: []nlp.WordBits{{
				Word: "cantaloupe",
				Bits: 2.0,
			}},
		},
		{
			name:  "both",
			prior: 5.0,
			left: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{{
					Word:  "delta",
					Count: 15,
				}},
			},
			right: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{{
					Word:  "delta",
					Count: 35,
				}},
			},
			want: []nlp.WordBits{{
				Word: "delta",
				Bits: 1.0,
			}},
		},
		{
			name:  "multiple",
			prior: 5.0,
			left: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{
					{
						Word:  "dog",
						Count: 15,
					},
					{
						Word:  "apple",
						Count: 15,
					},
					{
						Word:  "cat",
						Count: 15,
					},
				},
			},
			right: &nlp.FrequencyTable{
				Words: []*nlp.WordCount{
					{
						Word:  "apple",
						Count: 15,
					},
					{
						Word:  "dog",
						Count: 155,
					},
					{
						Word:  "banana",
						Count: 5,
					},
				},
			},
			want: []nlp.WordBits{
				{
					Word: "dog",
					Bits: 3.0,
				},
				{
					Word: "cat",
					Bits: -2.0,
				},
				{
					Word: "banana",
					Bits: 1.0,
				},
				{
					Word: "apple",
					Bits: 0.0,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := nlp.CharacteristicWords(tc.prior, tc.left, tc.right)

			if diff := cmp.Diff(tc.want, got,
				cmpopts.EquateApprox(0.0, 0.001),
				cmpopts.EquateEmpty(),
			); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
