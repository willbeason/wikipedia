package ingest_wikidata_test

import (
	"errors"
	"testing"

	"github.com/willbeason/wikipedia/pkg/documents"
	ingest_wikidata "github.com/willbeason/wikipedia/pkg/ingest-wikidata"
)

func TestParseObject(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name      string
		json      string
		titles    map[string]uint32
		wantError error
	}{{
		name: "Hank Green",
		json: HankGreen,
		titles: map[string]uint32{
			"Hank Green": 123,
		},
	}, {
		name: "Karen Barad",
		json: KarenBarad,
		titles: map[string]uint32{
			"Karen Barad": 123,
		},
	}}

	claimIDs := []string{
		"P21",
		"P27",
		"P31",
		"P39",
		"P101",
		"P106",
		"P172",
		"P279",
		"P361",
		"P569",
		"P570",
		"P734",
		"P735",
	}

	allowedInstances := map[string]bool{
		"Q5": true,
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			index := &documents.TitleIndex{Titles: tc.titles}

			obj, err := ingest_wikidata.ParseObject(allowedInstances, nil, claimIDs, index, []byte(tc.json))
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("got error %v, want %v", err, tc.wantError)
			}

			t.Errorf("%+v", obj)
		})
	}
}
