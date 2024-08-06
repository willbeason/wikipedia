package nlp_test

import "testing"

func TestParseTextClean(t *testing.T) {
	tt := []struct {
		name   string
		before string
		want   string
	}{{
		name: "Strip sections",
		before: `==Early Life==
stuff

==See also==
discard`,
		want: `stuff`,
	}}
}
