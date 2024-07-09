package tagtree

const (
	Asia         = "Asia"
	Africa       = "Africa"
	NorthAmerica = "North America"
	SouthAmerica = "South America"
	Antarctica   = "Antarctica"
	Europe       = "Europe"
	Oceania      = "Oceania"
)

var continentalOf = map[string]string{
	Asia:         "asian",
	Africa:       "african",
	NorthAmerica: "north american",
	SouthAmerica: "south american",
	Antarctica:   "antarctic",
	Europe:       "european",
	Oceania:      "oceanic",
}

type NodeContinent2Continental struct {
	Value Node
}

func (n *NodeContinent2Continental) String(title string) string {
	v := n.Value.String(title)

	continental, found := continentalOf[v]
	if !found {
		return "<MISSING CONTINENTAL>"
	}

	return continental
}
