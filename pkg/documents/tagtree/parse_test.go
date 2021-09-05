package tagtree

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestParse(t *testing.T) {
	tcs := []struct {
		name         string
		category     string
		wantNode     Node
		title        string
		wantCategory string
	}{
		{
			name:     "empty",
			category: "",
			wantNode: &NodeString{Value: ""},
		},
		{
			name:     "no tags",
			category: "Category:Philosophy",
			wantNode: &NodeString{Value: "Category:Philosophy"},
		},
		{
			name:         "title year node",
			title:        "After 1969",
			category:     "{{title year}}",
			wantNode:     &NodeTitleYear{},
			wantCategory: "1969",
		},
		{
			name:         "title year node 2",
			title:        "1969 Disco",
			category:     "{{Title year}}",
			wantNode:     &NodeTitleYear{},
			wantCategory: "1969",
		},
		{
			name:     "open close mismatch",
			category: "Category:{{Philosophy",
			wantNode: &NodeString{Value: "Category:{{Philosophy"},
		},
		{
			name:     "first close before first open",
			category: "Category:}}Philosophy{{",
			wantNode: &NodeString{Value: "Category:}}Philosophy{{"},
		},
		{
			name:     "last close before last open",
			category: "Category:{{}}Philosophy}}{{",
			wantNode: &NodeString{Value: "Category:{{}}Philosophy}}{{"},
		},
		{
			name:     "title year child",
			title:    "2010 in Football",
			category: "Category:{{title year}} in Sports",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleYear{},
				&NodeString{Value: " in Sports"},
			}},
			wantCategory: "Category:2010 in Sports",
		},
		{
			name:     "two title years",
			title:    "2011 in Soccer",
			category: "Category:{{title year}} in {{title year}} Sports",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleYear{},
				&NodeString{Value: " in "},
				&NodeTitleYear{},
				&NodeString{Value: " Sports"},
			}},
			wantCategory: "Category:2011 in 2011 Sports",
		},
		{
			name:     "decade from title year",
			title:    "Category:1889 in Los Angeles",
			category: "Category:{{DECADE|{{Title year}}}} in Los Angeles|{{Title year}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeDecade{Value: &NodeTitleYear{}},
				&NodeString{Value: " in Los Angeles|"},
				&NodeTitleYear{},
			}},
			wantCategory: "Category:1880s in Los Angeles|1889",
		},
		{
			name:     "month and year",
			title:    "Category:October 1961 events in Oceania",
			category: "Category:{{title year}} events in Oceania by month|{{MONTH|{{title monthname}}}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleYear{},
				&NodeString{Value: " events in Oceania by month|"},
				&NodeMonth{Value: &NodeTitleMonth{}},
			}},
			wantCategory: "Category:1961 events in Oceania by month|October",
		},
		{
			name:     "country",
			title:    "Category:December 1998 sports events in Thailand",
			category: "Category:{{title monthname}} {{title year}} events in {{title country}}|Sports",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleMonth{},
				&NodeString{Value: " "},
				&NodeTitleYear{},
				&NodeString{Value: " events in "},
				&NodeTitleCountry{},
				&NodeString{Value: "|Sports"},
			}},
			wantCategory: "Category:December 1998 events in Thailand|Sports",
		},
		{
			name:     "country 2",
			title:    "Category:2011 events in Thailand by month",
			category: "Category:Events in {{title country}}|*",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:Events in "},
				&NodeTitleCountry{},
				&NodeString{Value: "|*"},
			}},
			wantCategory: "Category:Events in Thailand|*",
		},
		{
			name:     "year range",
			title:    "Category:2016–17 Sun Belt Conference men's basketball season",
			category: "Category:{{Title year range}} NCAA Division I men's basketball season|Sun Belt",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleYearRange{},
				&NodeString{Value: " NCAA Division I men's basketball season|Sun Belt"},
			}},
			wantCategory: "Category:2016–17 NCAA Division I men's basketball season|Sun Belt",
		},
		{
			name:     "century name",
			title:    "Category:Energy companies disestablished in 2015",
			category: "Category:Energy companies disestablished in the {{century from year|{{title year}}|dash}}| ",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:Energy companies disestablished in the "},
				&NodeCentury{Value: &NodeTitleYear{}, Dash: true},
				&NodeString{Value: "| "},
			}},
			wantCategory: "Category:Energy companies disestablished in the 21st-century| ",
		},
		{
			name:     "century name 2",
			title:    "Category:1997 in Catalonia",
			category: "Category:Years of the {{Century name from title year}} in Catalonia|{{Title year}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:Years of the "},
				&NodeCentury{Value: &NodeTitleYear{}, Dash: false},
				&NodeString{Value: " in Catalonia|"},
				&NodeTitleYear{},
			}},
			wantCategory: "Category:Years of the 20th century in Catalonia|1997",
		},
		{
			name:     "century name 3",
			title:    "Category:1995 pinball machines",
			category: "Category:{{Century name from title year|dash}} pinball machines|{{Title year}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeCentury{Value: &NodeTitleYear{}, Dash: true},
				&NodeString{Value: " pinball machines|"},
				&NodeTitleYear{},
			}},
			wantCategory: "Category:20th-century pinball machines|1995",
		},
		{
			name:     "century name 4",
			title:    "Category:1990s radio programme endings",
			category: "Category:{{Century name from title decade|dash}} radio programme endings|{{Title decade}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeCentury{Value: &NodeTitleDecade{}, Dash: true},
				&NodeString{Value: " radio programme endings|"},
				&NodeTitleDecade{},
			}},
			wantCategory: "Category:20th-century radio programme endings|1990",
		},
		{
			name:     "century name 5",
			title:    "Category:19th-century disasters in Canada",
			category: "Category:{{Ordinal|{{Title century}}}}-century disasters in North America|Canada",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeOrdinal{Value: &NodeTitleCentury{}},
				&NodeString{Value: "-century disasters in North America|Canada"},
			}},
			wantCategory: "Category:19th-century disasters in North America|Canada",
		},
		{
			name:     "century name 6",
			title:    "Category:1910s in the United States by city",
			category: "Category:{{Century name from decade or year|{{Title decade}}s}} in the United States by city",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeCentury{Value: &NodeParent{Children: []Node{
					&NodeTitleDecade{},
					&NodeString{Value: "s"},
				}}},
				&NodeString{Value: " in the United States by city"},
			}},
			wantCategory: "Category:20th century in the United States by city",
		},
		{
			name:     "century to year",
			title:    "Category:11th-century famines",
			category: "Category:{{Century name from decade or year|{{Title century}}00|dash}} disasters|Famines",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeCentury{Value: &NodeParent{Children: []Node{
					&NodeTitleCentury{},
					&NodeString{Value: "00"},
				}}, Dash: true},
				&NodeString{Value: " disasters|Famines"},
			}},
			wantCategory: "Category:11th-century disasters|Famines",
		},
		{
			name:     "expression",
			title:    "Category:1978–79 Southern Hemisphere tropical cyclone season",
			category: "Category:Tropical cyclones in {{#expr:1+{{Title year}}}}|Southern Hemisphere",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:Tropical cyclones in "},
				&NodeExpression{Value: &NodeParent{Children: []Node{
					&NodeString{Value: "1+"},
					&NodeTitleYear{},
				}}},
				&NodeString{Value: "|Southern Hemisphere"},
			}},
			wantCategory: "Category:Tropical cyclones in 1979|Southern Hemisphere",
		},
		{
			name:     "month number",
			title:    "Category:August 2021 events in Germany",
			category: "Category:{{Title year}} events in {{Title country}} by month|{{MONTHNUMBER|{{Title monthname}}}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleYear{},
				&NodeString{Value: " events in "},
				&NodeTitleCountry{},
				&NodeString{Value: " by month|"},
				&NodeMonthNumber{Value: &NodeTitleMonth{}},
			}},
			wantCategory: "Category:2021 events in Germany by month|8",
		},
		{
			name:     "country2continent",
			title:    "Category:April 1960 events in Canada",
			category: "Category:{{title monthname}} {{title year}} events in {{country2continent|{{title country}}}}|{{title country}}",
			wantNode: &NodeParent{Children: []Node{
				&NodeString{Value: "Category:"},
				&NodeTitleMonth{},
				&NodeString{Value: " "},
				&NodeTitleYear{},
				&NodeString{Value: " events in "},
				&NodeCountry2Continent{Value: &NodeTitleCountry{}},
				&NodeString{Value: "|"},
				&NodeTitleCountry{},
			}},
			wantCategory: "Category:April 1960 events in North America|Canada",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gotNode := Parse(tc.category)

			if diff := cmp.Diff(tc.wantNode, gotNode); diff != "" {
				t.Fatal(diff)
			}

			gotCategory := gotNode.String(tc.title)
			wantCategory := tc.wantCategory
			if wantCategory == "" {
				wantCategory = tc.category
			}

			if diff := cmp.Diff(wantCategory, gotCategory); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
