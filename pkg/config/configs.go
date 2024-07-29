package config

type Extract struct {
	// Namespaces are the namespace IDs to include articles from.
	// Empty indicates to use articles from all namespaces.
	Namespaces []int
}
