package gender_test

import (
	"testing"

	"github.com/willbeason/wikipedia/pkg/analysis/gender"
	"github.com/willbeason/wikipedia/pkg/entities"
)

func TestInfer(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		claims       []*entities.Claim
		wantInferred string
	}{{
		name:         "no claims",
		claims:       []*entities.Claim{},
		wantInferred: gender.NoClaims,
	}, {
		name: "one claim",
		claims: []*entities.Claim{{
			Value: "male",
		}},
		wantInferred: "male",
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := gender.Infer(tc.claims)

			if got != tc.wantInferred {
				t.Errorf("got inferred gender %v, want %v", got, tc.wantInferred)
			}
		})
	}
}
