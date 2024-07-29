package config

type Extract struct {
	// ArticlesPath is the exact filepath to the pages-articles-multistream
	// dump of Wikipedia. Must be compressed as the associated Index points
	// to specific bytes of the compressed format.
	ArticlesPath string

	// IndexPath is the exact path to the index corresponding to the dump.
	// Accepts bz2 or uncompressed.
	IndexPath string

	// Namespaces are the namespace IDs to include articles from.
	// Empty indicates to use articles from all namespaces.
	Namespaces []int
}
