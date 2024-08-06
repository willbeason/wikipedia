package articles_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	articles "github.com/willbeason/wikipedia/pkg/article"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		wikitext string
		want     string
		wantErr  error
	}{{
		name:     "empty",
		wikitext: "",
		want:     "",
	}, {
		name:     "nowiki autoclose",
		wikitext: "<nowiki />",
		want:     "",
	}, {
		name:     "nowiki autoclose in text",
		wikitext: "some <nowiki />text",
		want:     "some text",
	}, {
		name:     "nowiki section",
		wikitext: "some <nowiki>more</nowiki> text",
		want:     "some more text",
	}, {
		name:     "nowiki within nowiki",
		wikitext: "some <nowiki><nowiki></nowiki> text",
		want:     "some <nowiki> text",
	}, {
		name:     "nowiki autoclose within nowiki",
		wikitext: "some <nowiki><nowiki /></nowiki> text",
		want:     "some <nowiki /> text",
	}, {
		name:     "template",
		wikitext: "some {{my template}} text",
		want:     "some  text",
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := articles.UnparsedText(tc.wikitext)
			gotParse, err := articles.Tokenize(wikitext)
			text := strings.Builder{}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error %v, want %v", err, tc.wantErr)
			}

			for _, token := range gotParse {
				text.WriteString(token.Render())
			}
			if diff := cmp.Diff(tc.want, text.String()); diff != "" {
				t.Errorf("(-want +got): %v", diff)
			}
		})
	}
}
