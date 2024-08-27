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
	// In is the workspace subdirectory from which to read articles.
	In string `yaml:"in"`

	// Index is the filepath to the title index.
	Index string `yaml:"index"`

	// Redirects is the filepath to the redirects list.
	Redirects string `yaml:"redirects"`

	// Out is the filepath to store the links index.
	Out string `yaml:"out"`
}

type Network struct {
	// In is the path to the file containing links.
	In string `yaml:"in"`

	// Links in IgnoredSections are not added to the network.
	IgnoredSections []string `yaml:"ignored_sections"`

	// If IgnoreCategories is enabled, no Category links are added to the network.
	IgnoreCategories bool `yaml:"ignore_categories"`
}

type TitleIndex struct {
	// InPath is the filepath to the extracted Wikipedia articles.
	In string `yaml:"in"`

	// Out is the filepath to store the title index.
	Out string `yaml:"out"`
}

type IngestWikidata struct {
	// Index is the filepath to the title index.
	Index string `yaml:"index"`

	// InstanceOf are the allowed values of entity types to ingest.
	InstanceOf []string `yaml:"instanceOf"`

	// RequireWikipedia are the InstanceOf values for which a Wikipedia article is required in order to ingest.
	// Must be a subset of InstanceOf.
	RequireWikipedia []string `yaml:"requireWikipedia"`

	// Claims are the claim IDs to ingest. All other claims will be discarded.
	Claims []string `yaml:"claims"`

	// Out is the filepath to store the Wikidata entities.
	Out string `yaml:"out"`
}

type GenderFrequency struct {
	// GenderIndex is the filepath to the extracted GenderIndex.
	GenderIndex string `yaml:"genderIndex"`
}

type GenderIndex struct {
	// Wikidata is the filepath to the extracted Wikidata.
	Wikidata string `yaml:"wikidata"`

	// Out is the filepath to write the protobuf containing gender information.
	Out string `yaml:"out"`
}

type GenderComparison struct {
	// GenderIndex is the filepath to the GenderIndex.
	GenderIndex string `yaml:"genderIndex"`

	// Links is the filepath to the list of links by article.
	Links string `yaml:"links"`
}
