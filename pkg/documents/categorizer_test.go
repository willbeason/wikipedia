package documents

import "testing"

func TestDisambiguateTags(t *testing.T) {
	tcs := []struct {
		name     string
		page     string
		category string
		want     string
	}{
		{
			name:     "empty string",
			page:     "",
			category: "",
			want:     "",
		},
		{
			name:     "decade from title year",
			page:     "Category:1969 in Argentine tennis",
			category: "Category:{{DECADE|{{Title year}}}} in Argentine tennis",
			want:     "Category:1960s in Argentine tennis",
		},
		{
			name:     "decade from title year",
			page:     "Category:1889 in Los Angeles",
			category: "Category:{{DECADE|{{Title year}}}} in Los Angeles|{{Title year}}",
			want:     "Category:1880s in Los Angeles|1889",
		},
		{
			name:     "month and year",
			page:     "Category:October 1961 events in Oceania",
			category: "Category:{{title year}} events in Oceania by month|{{MONTH|{{title monthname}}}}",
			want:     "Category:1961 events in Oceania by month|October",
		},
		{
			name:     "country",
			page:     "Category:December 1998 sports events in Thailand",
			category: "Category:{{title monthname}} {{title year}} events in {{title country}}|Sports",
			want:     "Category:December 1998 events in Thailand|Sports",
		},
		{
			name:     "century name",
			page:     "Category:Energy companies disestablished in 2015",
			category: "Category:Energy companies disestablished in the {{century from year|{{title year}}|dash}}| ",
			want:     "Category:Energy companies disestablished in the 21st-century| ",
		},
		{
			name:     "century name 2",
			page:     "Category:1997 in Catalonia",
			category: "Category:Years of the {{Century name from title year}} in Catalonia|{{Title year}}",
			want:     "Category:Years of the 20th century in Catalonia|1997",
		},
		{
			name:     "year range",
			page:     "Category:2016–17 Sun Belt Conference men's basketball season",
			category: "Category:{{Title year range}} NCAA Division I men's basketball season|Sun Belt",
			want:     "Category:2016–17 NCAA Division I men's basketball season|Sun Belt",
		},
		{
			name:     "century name 3",
			page:     "Category:1995 pinball machines",
			category: "Category:{{Century name from title year|dash}} pinball machines|{{Title year}}",
			want:     "Category:20th-century pinball machines|1995",
		},
		{
			name:     "century name 4",
			page:     "Category:1990s radio programme endings",
			category: "Category:{{Century name from title decade|dash}} radio programme endings|{{Title decade}}",
			want:     "Category:20th-century radio programme endings|1990s",
		},
		{
			name:     "century name 5",
			page:     "Category:1907 American novels",
			category: "Category:{{Century name from title year|dash}} American novels",
			want:     "Category:20th-century American novels",
		},
		{
			name:     "century name 6",
			page:     "Category:19th-century disasters in Canada",
			category: "Category:{{Ordinal|{{Title century}}}}-century disasters in North America|Canada",
			want:     "Category:19th-century disasters in North America|Canada",
		},
		{
			name:     "century name 7",
			page:     "Category:19th-century disasters in Canada",
			category: "Category:{{Ordinal|{{Title century}}}} century in Canada|Disasters",
			want:     "Category:19th century in Canada|Disasters",
		},
		{
			name:     "country name",
			page:     "Category:2011 events in Thailand by month",
			category: "Category:Events in {{title country}}|*",
			want:     "Category:Events in Thailand|*",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := DisambiguateTags(tc.page, tc.category)

			if got != tc.want {
				t.Errorf("got DisambiguateTags(%q, %q) =\n%q,\nwant %q", tc.page, tc.category, got, tc.want)
			}
		})
	}
}
