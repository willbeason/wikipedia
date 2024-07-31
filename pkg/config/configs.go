package config

type Extract struct {
	// ArticlesPath is the exact filepath to the pages-articles-multistream
	// dump of Wikipedia. Must be compressed as the associated Index points
	// to specific bytes of the compressed format.
	ArticlesPath string `yaml:"articlesPath"`

	// IndexPath is the exact path to the index corresponding to the dump.
	// Accepts bz2 or uncompressed.
	IndexPath string `yaml:"indexPath"`

	// Namespaces are the namespace IDs to include articles from.
	// Empty indicates to use articles from all namespaces.
	Namespaces []int `yaml:"namespaces"`
}
