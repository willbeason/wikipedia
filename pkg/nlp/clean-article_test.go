package nlp_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/willbeason/wikipedia/pkg/article"
	"github.com/willbeason/wikipedia/pkg/nlp"
)

func TestCleanArticle(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		text string
		want string
	}{{
		name: "link text",
		text: "She was described by [[Pavel Alexandrov]], [[Albert Einstein]], " +
			"[[Jean Dieudonné]], [[Hermann Weyl]] and [[Norbert Wiener]] as the " +
			"most important [[List of women in mathematics|woman in the history of mathematics]].",
		want: "She was described by Pavel Alexandrov, Albert Einstein, Jean Dieudonné, " +
			"Hermann Weyl and Norbert Wiener as the most important woman in the history of mathematics.",
	}, {
		name: "image description",
		text: "[[File:Hilbert.jpg|thumb|left|upright|In 1915 David Hilbert invited Noether to join " +
			"the Göttingen mathematics department, challenging the views of some of his colleagues " +
			"that a woman should not be allowed to teach at a university.]]",
		want: "In 1915 David Hilbert invited Noether to join the Göttingen mathematics department, " +
			"challenging the views of some of his colleagues that a woman should not be allowed to " +
			"teach at a university.",
	}, {
		name: "image description 2",
		text: "[[File:Emmy-noether-campus siegen.jpg|thumb|250px|right|The Emmy Noether Campus at the " +
			"University of Siegen is home to its mathematics and physics departments.]]",
		want: "The Emmy Noether Campus at the University of Siegen is home to its mathematics " +
			"and physics departments.",
	}, {
		name: "math",
		text: "An example of an ''invariant'' is the [[discriminant]] " +
			"{{math|''B''<sup>2</sup> − 4 ''AC''}} of a binary [[quadratic form]] " +
			"{{math|'''x·'''A '''x''' + '''y·'''B '''x''' + '''y·'''C '''y'''}}, " +
			"where {{math|'''x'''}} and {{math|'''y'''}} are [[Euclidean vector|vectors]] and " +
			"\"'''{{math|·}}'''\" is the [[dot product]] or \"[[inner product]]\" for the vectors. " +
			"{{math|A}}, {{math|B}}, and {{math|C}} are [[linear operator]]s on the vectors – " +
			"typically [[matrix (mathematics)|matrices]].",
		want: "An example of an ''invariant'' is the discriminant " +
			"_math_ of a binary quadratic form " +
			"_math_, " +
			"where _math_ and _math_ are vectors and " +
			"\"'''_math_'''\" is the dot product or \"inner product\" for the vectors. " +
			"_math_, _math_, and _math_ are linear operators on the vectors " +
			"– typically matrices.",
	}, {
		name: "parentheses",
		text: "In her classic 1921 paper ''Idealtheorie in Ringbereichen'' " +
			"(''Theory of Ideals in Ring Domains''), Noether developed the theory of " +
			"[[ideal (ring theory)|ideals]] in [[commutative ring]]s into a tool with " +
			"wide-ranging applications.",
		want: "In her classic 1921 paper ''Idealtheorie in Ringbereichen'' " +
			"(''Theory of Ideals in Ring Domains''), Noether developed the theory of " +
			"ideals in commutative rings into a tool with " +
			"wide-ranging applications.",
	}, {
		name: "full article",
		text: articles.EmmyNoetherBefore,
		want: articles.EmmyNoetherAfter,
	}}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := nlp.CleanArticle(tc.text)

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Error("-want +got", diff)
			}
		})
	}
}
