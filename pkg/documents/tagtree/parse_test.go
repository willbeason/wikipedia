package tagtree_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/willbeason/wikipedia/pkg/documents/tagtree"
)

func TestParse(t *testing.T) {
	tcs := []struct {
		name         string
		category     string
		wantNode     tagtree.Node
		title        string
		wantCategory string
	}{
		{
			name:     "empty",
			category: "",
			wantNode: &tagtree.NodeString{Value: ""},
		},
		{
			name:     "no tags",
			category: "Category:Philosophy",
			wantNode: &tagtree.NodeString{Value: "Category:Philosophy"},
		},
		{
			name:         "title year node",
			title:        "After 1969",
			category:     "{{title year}}",
			wantNode:     &tagtree.NodeTitleYear{},
			wantCategory: "1969",
		},
		{
			name:         "title year node 2",
			title:        "1969 Disco",
			category:     "{{Title year}}",
			wantNode:     &tagtree.NodeTitleYear{},
			wantCategory: "1969",
		},
		{
			name:     "open close mismatch",
			category: "Category:{{Philosophy",
			wantNode: &tagtree.NodeString{Value: "Category:{{Philosophy"},
		},
		{
			name:     "first close before first open",
			category: "Category:}}Philosophy{{",
			wantNode: &tagtree.NodeString{Value: "Category:}}Philosophy{{"},
		},
		{
			name:     "last close before last open",
			category: "Category:{{}}Philosophy}}{{",
			wantNode: &tagtree.NodeString{Value: "Category:{{}}Philosophy}}{{"},
		},
		{
			name:     "title year child",
			title:    "2010 in Football",
			category: "Category:{{title year}} in Sports",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " in Sports"},
			}},
			wantCategory: "Category:2010 in Sports",
		},
		{
			name:     "two title years",
			title:    "2011 in Soccer",
			category: "Category:{{title year}} in {{title year}} Sports",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " in "},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " Sports"},
			}},
			wantCategory: "Category:2011 in 2011 Sports",
		},
		{
			name:     "decade from title year",
			title:    "Category:1889 in Los Angeles",
			category: "Category:{{DECADE|{{Title year}}}} in Los Angeles|{{Title year}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeDecade{Value: &tagtree.NodeTitleYear{}},
				&tagtree.NodeString{Value: " in Los Angeles|"},
				&tagtree.NodeTitleYear{},
			}},
			wantCategory: "Category:1880s in Los Angeles|1889",
		},
		{
			name:     "month and year",
			title:    "Category:October 1961 events in Oceania",
			category: "Category:{{title year}} events in Oceania by month|{{MONTH|{{title monthname}}}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " events in Oceania by month|"},
				&tagtree.NodeMonth{Value: &tagtree.NodeTitleMonth{}},
			}},
			wantCategory: "Category:1961 events in Oceania by month|October",
		},
		{
			name:     "country",
			title:    "Category:December 1998 sports events in Thailand",
			category: "Category:{{title monthname}} {{title year}} events in {{title country}}|Sports",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleMonth{},
				&tagtree.NodeString{Value: " "},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " events in "},
				&tagtree.NodeTitleCountry{},
				&tagtree.NodeString{Value: "|Sports"},
			}},
			wantCategory: "Category:December 1998 events in Thailand|Sports",
		},
		{
			name:     "country 2",
			title:    "Category:2011 events in Thailand by month",
			category: "Category:Events in {{title country}}|*",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:Events in "},
				&tagtree.NodeTitleCountry{},
				&tagtree.NodeString{Value: "|*"},
			}},
			wantCategory: "Category:Events in Thailand|*",
		},
		{
			name:     "year range",
			title:    "Category:2016–17 Sun Belt Conference men's basketball season",
			category: "Category:{{Title year range}} NCAA Division I men's basketball season|Sun Belt",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYearRange{},
				&tagtree.NodeString{Value: " NCAA Division I men's basketball season|Sun Belt"},
			}},
			wantCategory: "Category:2016–17 NCAA Division I men's basketball season|Sun Belt",
		},
		{
			name:     "century name",
			title:    "Category:Energy companies disestablished in 2015",
			category: "Category:Energy companies disestablished in the {{century from year|{{title year}}|dash}}| ",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:Energy companies disestablished in the "},
				&tagtree.NodeCentury{Value: &tagtree.NodeTitleYear{}, Dash: true},
				&tagtree.NodeString{Value: "| "},
			}},
			wantCategory: "Category:Energy companies disestablished in the 21st-century| ",
		},
		{
			name:     "century name 2",
			title:    "Category:1997 in Catalonia",
			category: "Category:Years of the {{Century name from title year}} in Catalonia|{{Title year}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:Years of the "},
				&tagtree.NodeCentury{Value: &tagtree.NodeTitleYear{}, Dash: false},
				&tagtree.NodeString{Value: " in Catalonia|"},
				&tagtree.NodeTitleYear{},
			}},
			wantCategory: "Category:Years of the 20th century in Catalonia|1997",
		},
		{
			name:     "century name 3",
			title:    "Category:1995 pinball machines",
			category: "Category:{{Century name from title year|dash}} pinball machines|{{Title year}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeCentury{Value: &tagtree.NodeTitleYear{}, Dash: true},
				&tagtree.NodeString{Value: " pinball machines|"},
				&tagtree.NodeTitleYear{},
			}},
			wantCategory: "Category:20th-century pinball machines|1995",
		},
		{
			name:     "century name 4",
			title:    "Category:1990s radio program endings",
			category: "Category:{{Century name from title decade|dash}} radio program endings|{{Title decade}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeCentury{Value: &tagtree.NodeTitleDecade{}, Dash: true},
				&tagtree.NodeString{Value: " radio program endings|"},
				&tagtree.NodeTitleDecade{},
			}},
			wantCategory: "Category:20th-century radio program endings|1990",
		},
		{
			name:     "century name 5",
			title:    "Category:19th-century disasters in Canada",
			category: "Category:{{Ordinal|{{Title century}}}}-century disasters in North America|Canada",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeOrdinal{Value: &tagtree.NodeTitleCentury{}},
				&tagtree.NodeString{Value: "-century disasters in North America|Canada"},
			}},
			wantCategory: "Category:19th-century disasters in North America|Canada",
		},
		{
			name:     "century name 6",
			title:    "Category:1910s in the United States by city",
			category: "Category:{{Century name from decade or year|{{Title decade}}s}} in the United States by city",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeCentury{Value: &tagtree.NodeParent{Children: []tagtree.Node{
					&tagtree.NodeTitleDecade{},
					&tagtree.NodeString{Value: "s"},
				}}},
				&tagtree.NodeString{Value: " in the United States by city"},
			}},
			wantCategory: "Category:20th century in the United States by city",
		},
		{
			name:     "century to year",
			title:    "Category:11th-century famines",
			category: "Category:{{Century name from decade or year|{{Title century}}00|dash}} disasters|Famines",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeCentury{Value: &tagtree.NodeParent{Children: []tagtree.Node{
					&tagtree.NodeTitleCentury{},
					&tagtree.NodeString{Value: "00"},
				}}, Dash: true},
				&tagtree.NodeString{Value: " disasters|Famines"},
			}},
			wantCategory: "Category:11th-century disasters|Famines",
		},
		{
			name:     "expression",
			title:    "Category:1978–79 Southern Hemisphere tropical cyclone season",
			category: "Category:Tropical cyclones in {{#expr:1+{{Title year}}}}|Southern Hemisphere",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:Tropical cyclones in "},
				&tagtree.NodeExpression{Value: &tagtree.NodeParent{Children: []tagtree.Node{
					&tagtree.NodeString{Value: "1+"},
					&tagtree.NodeTitleYear{},
				}}},
				&tagtree.NodeString{Value: "|Southern Hemisphere"},
			}},
			wantCategory: "Category:Tropical cyclones in 1979|Southern Hemisphere",
		},
		{
			name:     "month number",
			title:    "Category:August 2021 events in Germany",
			category: "Category:{{Title year}} events in {{Title country}} by month|{{MONTHNUMBER|{{Title monthname}}}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " events in "},
				&tagtree.NodeTitleCountry{},
				&tagtree.NodeString{Value: " by month|"},
				&tagtree.NodeMonthNumber{Value: &tagtree.NodeTitleMonth{}},
			}},
			wantCategory: "Category:2021 events in Germany by month|8",
		},
		{
			name:     "country2continent",
			title:    "Category:April 1960 events in Canada",
			category: "Category:{{title monthname}} {{title year}} events in {{country2continent|{{title country}}}}|{{title country}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleMonth{},
				&tagtree.NodeString{Value: " "},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " events in "},
				&tagtree.NodeCountry2Continent{Value: &tagtree.NodeTitleCountry{}},
				&tagtree.NodeString{Value: "|"},
				&tagtree.NodeTitleCountry{},
			}},
			wantCategory: "Category:April 1960 events in North America|Canada",
		},
		{
			name:     "country2nationality",
			title:    "Category:July 2018 sports events in Switzerland",
			category: "Category:{{title year}} in {{country2nationality|{{title country}}}} sport|{{MONTH|{{title monthname}}}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " in "},
				&tagtree.NodeCountry2Nationality{Value: &tagtree.NodeTitleCountry{}},
				&tagtree.NodeString{Value: " sport|"},
				&tagtree.NodeMonth{Value: &tagtree.NodeTitleMonth{}},
			}},
			wantCategory: "Category:2018 in Swiss sport|July",
		},
		{
			name:     "continent2continental",
			title:    "Category:August 1968 sports events in Europe",
			category: "Category:{{title year}} in {{Continent2continental|{{title country}}}} sport||{{MONTH|{{title monthname}}}}",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeTitleYear{},
				&tagtree.NodeString{Value: " in "},
				&tagtree.NodeContinent2Continental{Value: &tagtree.NodeTitleCountry{}},
				&tagtree.NodeString{Value: " sport||"},
				&tagtree.NodeMonth{Value: &tagtree.NodeTitleMonth{}},
			}},
			wantCategory: "Category:1968 in european sport||August",
		},
		{
			name:     "first word",
			title:    "Category:Örebro Garrison",
			category: "Category:{{first word|{{PAGENAME}}}} Municipality",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeFirstWord{Value: &tagtree.NodePageName{}},
				&tagtree.NodeString{Value: " Municipality"},
			}},
			wantCategory: "Category:Örebro Municipality",
		},
		{
			name:     "last word",
			title:    "Category:Örebro Garrison",
			category: "Category:{{last word|{{PAGENAME}}}} Municipality",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeLastWord{Value: &tagtree.NodePageName{}},
				&tagtree.NodeString{Value: " Municipality"},
			}},
			wantCategory: "Category:Garrison Municipality",
		},
		{
			name:     "first decade in century",
			title:    "Category:2000s in space",
			category: "Category:{{Century name from title decade}} in space",
			wantNode: &tagtree.NodeParent{Children: []tagtree.Node{
				&tagtree.NodeString{Value: "Category:"},
				&tagtree.NodeCentury{Value: &tagtree.NodeTitleDecade{}},
				&tagtree.NodeString{Value: " in space"},
			}},
			wantCategory: "Category:21st century in space",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gotNode := tagtree.Parse(tc.category)

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
