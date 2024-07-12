package config

type Config struct {
	// WikiPath is where to store intermediate outputs, such as the article database.
	// Should be an exact path.
	WikiPath string

	Extract Extract

	PageRank PageRank
}

// Extract is options related to extracting Wikipedia from the compressed format.
type Extract struct {
	// Namespaces are the namespace IDs to include articles from.
	// Empty indicates to use articles from all namespaces.
	Namespaces []int
}

// PageRank is options related to calculating the PageRank of various articles.
type PageRank struct {
	// Iterations is how many times to iterate the PageRank calculation before timing out.
	Iterations int

	// ConvergeThreshold is when to stop iterating the PageRank calculation, if the PageRank vector is changing
	// slowly enough.
	ConvergeThreshold float64

	// Filter is the exact text string to require all articles to contain, case-insensitive.
	// If unset, calculates PageRank for all articles.
	Filter string
}
