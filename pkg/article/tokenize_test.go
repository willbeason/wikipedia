package article_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/article"
)

func TestClean(t *testing.T) {
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
		name:     "nowiki within nowiki 2",
		wikitext: "<nowiki><nowiki></nowiki></nowiki>",
		want:     "<nowiki></nowiki>",
	}, {
		name:     "nowiki not closed",
		wikitext: "<nowiki>a b c",
		want:     "<nowiki>a b c",
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
	}, {
		name:     "Reference unquotes",
		wikitext: `<ref name=Weyl></ref>`,
		want:     "",
	}, {
		name:     "Reference spaced",
		wikitext: `<ref name = Weyl ></ref>`,
		want:     "",
	}, {
		name:     "File link",
		wikitext: `[[File:Wikipedesketch.png|thumb|alt=A cartoon centipede ... detailed description.|The Wikipede edits ''[[Myriapoda]]''.]]`,
		want:     "The Wikipede edits ''Myriapoda''.",
	}, {
		name:     "NBSP",
		wikitext: `Noether c.&nbsp;1930`,
		want:     "Noether c. 1930",
	}, {
		name:     "Blockquote",
		wikitext: `<blockquote>The development of abstract algebra, which is one of the most distinctive innovations of twentieth century mathematics, is largely due to her – in published papers, in lectures, and in personal influence on her contemporaries.</blockquote>`,
		want:     "The development of abstract algebra, which is one of the most distinctive innovations of twentieth century mathematics, is largely due to her – in published papers, in lectures, and in personal influence on her contemporaries.",
	}, {
		name:     "Emphasis",
		wikitext: `<em>Noether Boys</em>`,
		want:     "Noether Boys",
	}, {
		name:     "Emphasis2",
		wikitext: `Other <em>Noether Boys</em> included [[Max Deuring]], [[Hans Fitting]], [[Ernst Witt]], [[Chiungtze C. Tsen]] and`,
		want:     "Other Noether Boys included Max Deuring, Hans Fitting, Ernst Witt, Chiungtze C. Tsen and",
	}, {
		name:     "Math",
		wikitext: `<math>A_{1} \subset A_{2} \subset A_{3} \subset \cdots.</math>`,
		want:     article.MathToken,
	}, {
		name:     "Subscript",
		wikitext: `the "m<sub>μν</sub>-riddle of syllables"`,
		want:     `the "m_μν-riddle of syllables"`,
	}, {
		name:     "HTTP Links",
		wikitext: `[https://web.archive.org/web/20070929100418/http://www.physikerinnen.de/noetherlebenslauf.html Noether's application for admission to the University of Erlangen and three of her curriculum vitae] from the website of historian Cordula Tollmien`,
		want:     `Noether's application for admission to the University of Erlangen and three of her curriculum vitae from the website of historian Cordula Tollmien`,
	}, {
		name:     "References",
		wikitext: `A {{refbegin|30em}}Stuff{{refend}} thing`,
		want:     `A  thing`,
	}, {
		name:     "Blockquote template",
		wikitext: `{{Blockquote|In the judgment of the most competent living mathematicians, Fräulein Noether was the most significant creative mathematical [[genius]] thus far produced since the higher education of women began. In the realm of algebra, in which the most gifted mathematicians have been busy for centuries, she discovered methods which have proved of enormous importance in the development of the present-day younger generation of mathematicians.}}`,
		want:     `In the judgment of the most competent living mathematicians, Fräulein Noether was the most significant creative mathematical genius thus far produced since the higher education of women began. In the realm of algebra, in which the most gifted mathematicians have been busy for centuries, she discovered methods which have proved of enormous importance in the development of the present-day younger generation of mathematicians.`,
	}, {
		name:     "Normal header",
		wikitext: "\n==Biography==\n",
		want:     `Biography`,
	}, {
		name:     "Header starts with equals",
		wikitext: "\n===a=b==\n",
		want:     `=a=b`,
	}, {
		name:     "Header ends with equals",
		wikitext: "\n==a=b===\n",
		want:     `a=b=`,
	}, {
		name:     "Max header depth",
		wikitext: "\n======Very specific======\n",
		want:     `Very specific`,
	}, {
		name:     "Beyond max header depth",
		wikitext: "\n=======Too specific=======\n",
		want:     `=Too specific=`,
	}, {
		name:     "Ignore notes section",
		wikitext: "\n==Notes==\nThings\nMore things\n==Other==\nDisplayed",
		want: `Other
Displayed`,
	}, {
		name:     "Header and immediate subheader",
		wikitext: "\n==Biography==\n===Early Life===\n",
		want: `Biography
Early Life`,
	}, {
		name:     "ref group",
		wikitext: "<ref group=note name=\"sBl2q\" />",
		want:     ``,
	}, {
		name:     "comment",
		wikitext: "<!-- Please do not change this—see talk page and its many archives.-->",
		want:     ``,
	}, {
		name:     "superscript",
		wikitext: "E=mc<sup>2</sup>",
		want:     `E=mc^2`,
	}, {
		name:     "comment with angle bracket",
		wikitext: "<!-- some comment>-->a",
		want:     `a`,
	}, {
		name:     "nested comment",
		wikitext: "<!-- <!-- nested comment -->  -->",
		want:     `  -->`,
	}, {
		name:     "link in header",
		wikitext: "=====[[Geometric abstraction]] and related movements=====",
		want:     `Geometric abstraction and related movements`,
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := article.UnparsedText(tc.wikitext)
			gotParse := article.Tokenize(wikitext)

			sb := strings.Builder{}
			for _, token := range gotParse {
				sb.WriteString(token.Render())
			}

			text := sb.String()
			text = strings.Trim(text, "\n")

			if diff := cmp.Diff(tc.want, text); diff != "" {
				t.Errorf("(-want +got): %v", diff)
			}
		})
	}
}

func TestLinks(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		wikitext string
		want     []string
	}{{
		name: "normal link",
		wikitext: `==Practice==
[[Joey Logano]] was the fastest in the practice session with a time of 32.603 seconds and a speed of {{Convert|138.024|mph|km/h|abbr=on}}.<ref>{{cite web|last=Utter|first=Jim|title=Gateway NASCAR Cup: Penske drivers top Saturday practice|url=https://us.motorsport.com/nascar-cup/news/gateway-nascar-cup-penske-drivers-top-saturday-practice-/10618437/|website=[[Motorsport.com]]|publisher=[[Motorsport Network]]|access-date=June 1, 2024|location=[[Madison, Illinois]]|date=June 1, 2024}}</ref>`,
		want: []string{
			"Joey Logano",
		},
	}, {
		name:     "empty link",
		wikitext: `'''Sør-Gjæslingan''' is an archipelago in [[]], [[Trøndelag]], [[Norway]].`,
		want: []string{
			"Trøndelag",
			"Norway",
		},
	}, {
		name: "whitespace-only link",
		wikitext: `'''Sør-Gjæslingan''' is an archipelago in [[

]], [[Trøndelag]], [[Norway]].`,
		want: []string{
			"Trøndelag",
			"Norway",
		},
	}, {
		name:     "whitespace-only link",
		wikitext: `{{nihongo |'''Nana Fujii'''|藤井 奈々|Fujii Nana|born March 31, 1998}} is a Japanese [[Professional shogi player#Women's professionals|women's professional shogi player]] ranked 1-[[Dan (rank)#Modern usage in shogi|dan]].<ref>{{cite web|url=https://www.shogi.or.jp/player/lady/63.html|script-title=ja:女流棋士データベース: 藤井奈々|title=Joryū Kishi Dētabēsu: Fujii Nana|language=ja|trans-title=Women's Professional Shogi Player Database: Nana Fujii |publisher=[[Japan Shogi Association]]|access-date=April 6, 2020}}</ref>`,
		want: []string{
			"Professional shogi player",
			"Dan (rank)",
		},
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := article.UnparsedText(tc.wikitext)
			gotParse := article.Tokenize(wikitext)
			gotLinks := article.ToLinkTargets(gotParse)

			diff := cmp.Diff(gotLinks, tc.want)
			if diff != "" {
				t.Errorf("(-want +got): %v", diff)
				// for _, l := range gotLinks {
				// 	fmt.Printf("%q,\n", l)
				// }
			}
		})
	}
}

func TestLinks_RealArticles(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		wikitext string
		want     []string
	}{{
		name:     "Emmy Noether",
		wikitext: EmmyNoether,
		want:     NoetherLinks,
	}, {
		name:     "Pungtungia Hilgendorfi",
		wikitext: PungtungiaHilgendorfi,
		want:     PungtungiaHilgendorfiLinks,
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := article.UnparsedText(tc.wikitext)
			gotParse := article.Tokenize(wikitext)
			gotLinks := article.ToLinkTargets(gotParse)

			diff := cmp.Diff(gotLinks, tc.want)
			if diff != "" {
				t.Errorf("(-want +got): %v", diff)
				// for _, l := range gotLinks {
				// 	fmt.Printf("%q,\n", l)
				// }
			}
		})
	}
}

func TestClean_RealArticles(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		wikitext string
		want     string
		debug    bool
	}{{
		name:     "Emmy Noether",
		wikitext: EmmyNoether,
		want:     EmmyNoetherAfter,
	}, {
		name:     "Albert Einstein",
		wikitext: AlbertEinstein,
		want:     AlbertEinsteinAfter,
		// debug:    true,
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wikitext := article.UnparsedText(tc.wikitext)
			gotParse := article.Tokenize(wikitext)

			text := strings.Builder{}
			for _, token := range gotParse {
				text.WriteString(token.Render())
			}

			diff := cmp.Diff(tc.want, text.String())
			if diff != "" {
				if tc.debug {
					fmt.Println(text.String())
					t.Fatal()
				} else {
					t.Errorf("(-want +got): %v", diff)
				}
			}
		})
	}
}
