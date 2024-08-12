package config

type Ingest struct {
	// Namespaces are the namespace IDs to include articles from.
	Namespaces []int `yaml:"namespaces"`
}

type Clean struct {
	// In is the workspace subdirectory from which to clean articles.
	In string `yaml:"in"`

	// Out is the workspace subdirectory in which to store the cleaned articles.
	Out string `yaml:"out"`
}

type Links struct {
	// In is the workspace subdirectory from which to clean articles.
	In string `yaml:"in"`

	// Index is the filepath to the title index.
	Index string `yaml:"index"`

	// Out is the filepath to store the links index.
	Out string `yaml:"out"`
}

type TitleIndex struct {
	// InPath is the filepath to the extracted Wikipedia articles.
	In string `yaml:"in"`

	// Out is the filepath to store the title index.
	Out string `yaml:"out"`
}
