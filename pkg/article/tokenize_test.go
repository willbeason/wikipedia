package article_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/article"
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
		name:     "unrendered template",
		wikitext: "some {{my template}} text",
		want:     "some  text",
	}, {
		name:     "IPAc-en US template",
		wikitext: "{{IPAc-en|US|ˈ|n|ʌ|t|ər}}",
		want:     "US: /ˈnʌtər/",
	}, {
		name:     "IPAc-de template",
		wikitext: "{{IPA-de|ˈnøːtɐ|lang}}",
		want:     "German: [ˈnøːtɐ]",
	}, {
		name:     "Ref cite",
		wikitext: "<ref>{{cite web |first=Emily |last=Conover |author-link=Emily Conover |date=12 June 2018 |title=In her short life, mathematician Emmy Noether changed the face of physics |url=https://www.sciencenews.org/article/emmy-noether-theorem-legacy-physics-math |access-date=2 July 2018 |website=[[Science News]] |url-status=live |archive-url=https://web.archive.org/web/20230326222502/https://www.sciencenews.org/article/emmy-noether-theorem-legacy-physics-math |archive-date=26 March 2023}}</ref>",
		want:     "",
	}, {
		name:     "Ref",
		wikitext: "<ref>Stuff</ref>",
		want:     "",
	}, {
		name:     "Unclosed ref",
		wikitext: "<ref>Stuff",
		want:     "<ref>Stuff",
	}, {
		name:     "Unopened ref",
		wikitext: "Stuff</ref>",
		want:     "Stuff</ref>",
	}, {
		name:     "Link no display",
		wikitext: "[[Jewish family]]",
		want:     "Jewish family",
	}, {
		name:     "Link display",
		wikitext: "[[module (mathematics)|modules]]",
		want:     "modules",
	}, {
		name:     "Named reference",
		wikitext: `<ref name="MacTutorStudents"/>`,
		want:     "",
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := article.UnparsedText(tc.wikitext)
			gotParse, err := article.Tokenize(wikitext)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error %v, want %v", err, tc.wantErr)
			}

			text := strings.Builder{}
			for _, token := range gotParse {
				text.WriteString(token.Render())
			}

			if diff := cmp.Diff(tc.want, text.String()); diff != "" {
				t.Errorf("(-want +got): %v", diff)
			}
		})
	}
}

func TestTokenize_Noether(t *testing.T) {
	wikitext := article.UnparsedText(EmmyNoetherBefore)
	gotParse, err := article.Tokenize(wikitext)
	if err != nil {
		t.Fatal(err)
	}

	text := strings.Builder{}
	for _, token := range gotParse {
		text.WriteString(token.Render())
	}

	if diff := cmp.Diff(EmmyNoetherAfter, text.String()); diff != "" {
		t.Errorf("(-want +got): %v", diff)
	}
	// fmt.Println(text.String())
	// t.Fatal()
}
