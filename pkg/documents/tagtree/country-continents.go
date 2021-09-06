package tagtree

import (
	"fmt"
	"strings"
)

var continentOf = map[string]string{
	"Australia":          Oceania,
	"Bangladesh":         Asia,
	"Canada":             NorthAmerica,
	"China":              Asia,
	"France":             Europe,
	"Germany":            Europe,
	"India":              Asia,
	"Italy":              Europe,
	"Japan":              Asia,
	"Mexico":             NorthAmerica,
	"New Zealand":        Oceania,
	"North Korea":        Asia,
	"the Philippines":    Asia,
	"Romania":            Europe,
	"Russia":             Asia,
	"Spain":              Europe,
	"South Korea":        Asia,
	"Thailand":           Asia,
	"the United Kingdom": Europe,
	"the United States":  NorthAmerica,

	"the Netherlands":          Europe,
	"Portugal":                 Europe,
	"Switzerland":              Europe,
	"Belgium":                  Europe,
	"Poland":                   Europe,
	"Pakistan":                 Asia,
	"Austria":                  Europe,
	"South Africa":             Africa,
	"Sweden":                   Europe,
	"Hungary":                  Europe,
	"Egypt":                    Africa,
	"Niger":                    Africa,
	"Nigeria":                  Africa,
	"Turkey":                   Asia,
	"Kazakhstan":               Asia,
	"the Czech Republic":       Europe,
	"Indonesia":                Asia,
	"Serbia":                   Europe,
	"Croatia":                  Europe,
	"Bulgaria":                 Europe,
	"the United Arab Emirates": Asia,
	"Vietnam":                  Asia,
	"Malaysia":                 Asia,
	"Slovenia":                 Europe,
	"Brunei":                   Asia,
	"Syria":                    Asia,
	"Estonia":                  Europe,
	"Malta":                    Europe,
	"Mongolia":                 Asia,
	"Iraq":                     Asia,
	"Kenya":                    Africa,
	"Montenegro":               Europe,
	"Chile":                    SouthAmerica,
	"Guyana":                   SouthAmerica,
	"Burundi":                  Africa,
	"Chad":                     Africa,
	"Mozambique":               Africa,
	"South Sudan":              Africa,
	"Moldova":                  Europe,
	"Botswana":                 Africa,
	"Algeria":                  Africa,
	"Togo":                     Africa,
	"Tajikistan":               Asia,
	"Lebanon":                  Asia,
	"Cameroon":                 Africa,
	"Mauritius":                Africa,
	"Ghana":                    Africa,
	"Denmark":                  Europe,
	"Uruguay":                  SouthAmerica,
	"Nepal":                    Asia,
	"Israel":                   Asia,
	"Monaco":                   Europe,
	"Senegal":                  Africa,
	"Republic of the Congo":    Africa,
}

type NodeCountry2Continent struct {
	Value Node
}

func (n *NodeCountry2Continent) String(title string) string {
	v := n.Value.String(title)
	v = strings.Trim(v, "*[]")

	continent, found := continentOf[v]
	if !found {
		return fmt.Sprintf("<MISSING COUNTRY CONTINENT %q>", v)
	}

	return continent
}
