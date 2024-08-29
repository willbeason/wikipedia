package gender

import (
	"github.com/willbeason/wikipedia/pkg/entities"
)

const (
	DeprecatedRank = "deprecated"
	NormalRank     = "normal"
	PreferredRank  = "preferred"

	IntersexGender       = "Q1097630"
	IntersexPersonGender = "Q104717073"
	TransgenderGender    = "Q189125"
	EunuchGender         = "Q179294"
	XGender              = "Q96000630"
	NonBinaryGender      = "Q48270"
	TransWomanGender     = "Q1052281"
	WomanGender          = "Q6581072"
	TransManGender       = "Q2449503"
	ManGender            = "Q6581097"

	NoClaims          = "NO CLAIMS"
	ConflictingClaims = "CONFLICTING CLAIMS"
)

var ReadableGender = map[string]string{
	ManGender: "man",
	WomanGender: "woman",
	TransWomanGender: "trans woman",
	ConflictingClaims: "conflict",
	NonBinaryGender: "non-binary",
	TransManGender: "trans man",
	NoClaims: "no claims",
	"Q121307100": "intersex woman",
	"Q18116794": "genderfluid",
	"Q121307094": "intersex man",
	"Q12964198": "genderqueer",
	EunuchGender: "eunuch",
	"Q189125": "transgender",
	"Q301702": "two-spirit",
	"Q505371": "agender",
}

func filterDeprecated(claims []*entities.Claim) []*entities.Claim {
	result := make([]*entities.Claim, 0, len(claims))

	for _, c := range claims {
		if c.Rank == DeprecatedRank {
			continue
		}
		result = append(result, c)
	}

	return result
}

func Infer(claims []*entities.Claim) string {
	// Always ignore gender claims marked deprecated.
	claims = filterDeprecated(claims)

	if len(claims) == 0 {
		return NoClaims
	}

	if len(claims) == 1 {
		return claims[0].Value
	}

	// Choose preferred if only one preferred.
	numPreferred := 0
	preferred := ""
	claimedGenders := make(map[string]int)
	for _, claim := range claims {
		if claim.Rank == PreferredRank {
			preferred = claim.Value
			numPreferred++
		}

		switch claim.Value {
		case XGender:
			claimedGenders[NonBinaryGender]++
		case TransManGender:
			// Count trans men as men.
			claimedGenders[ManGender]++
		case TransWomanGender:
			// Count trans women as women.
			claimedGenders[WomanGender]++
		case IntersexPersonGender:
			claimedGenders[IntersexGender]++
		default:
			claimedGenders[claim.Value]++
		}
	}
	if numPreferred == 1 {
		return preferred
	}

	// If just multiple claims for the same gender, use that gender.
	if len(claimedGenders) == 1 {
		for k := range claimedGenders {
			return k
		}
	}

	return ConflictingClaims
}
