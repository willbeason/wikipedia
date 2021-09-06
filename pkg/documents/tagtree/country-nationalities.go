package tagtree

var nationalityOf = map[string]string{
	"Switzerland": "Swiss",

	"Australia":                "Australian",
	"New Zealand":              "New Zealand",
	"Russia":                   "Russian",
	"the United States":        "American",
	"the United Kingdom":       "British",
	"France":                   "French",
	"South Korea":              "South Korean",
	"Mexico":                   "Mexican",
	"Japan":                    "Japanese",
	"Italy":                    "Italian",
	"Romania":                  "Romanian",
	"China":                    "Chinese",
	"Spain":                    "Spanish",
	"Germany":                  "German",
	"Canada":                   "Canadian",
	"India":                    "Indian",
	"the Netherlands":          "Norwegian",
	"Portugal":                 "Portuguese",
	"Belgium":                  "Belgian",
	"Pakistan":                 "Pakistani",
	"Thailand":                 "Thai",
	"Austria":                  "Austrian",
	"Poland":                   "Polish",
	"Sweden":                   "Swedish",
	"South Africa":             "South African",
	"Hungary":                  "Hungarian",
	"Bangladesh":               "Bangladeshi",
	"Egypt":                    "Egyptian",
	"the Czech Republic":       "Czech",
	"the United Arab Emirates": "Emirati",
	"Niger":                    "Nigerien",
	"Nigeria":                  "Nigerian",
	"Croatia":                  "Croatian",
	"Indonesia":                "Indonesian",
	"Serbia":                   "Serbian",
	"Slovenia":                 "Slovenian",
	"Estonia":                  "Estonian",
	"Bulgaria":                 "Bulgarian",
	"Turkey":                   "Turkish",
	"Guyana":                   "Guyana",
	"Monaco":                   "Monacan",
	"Mali":                     "Malian",
	"Lebanon":                  "Lebanese",
	"Bosnia and Herzegovina":   "Bosnian",
	"Vietnam":                  "Vietnamese",
	"Denmark":                  "Danish",
	"Chile":                    "Chilean",
	"Azerbaijan":               "Azerbaijani",
	"Mauritius":                "Mauritian",
}

type NodeCountry2Nationality struct {
	Value Node
}

func (n *NodeCountry2Nationality) String(title string) string {
	v := n.Value.String(title)

	nationality, found := nationalityOf[v]
	if !found || nationality == "" {
		return "<MISSING COUNTRY NATIONALITY>"
	}

	return nationality
}
