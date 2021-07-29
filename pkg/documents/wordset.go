package documents

type WordSets struct {
	InFile    string `json:"in_file,omitempty"`
	Documents []WordSet
}

type WordSet struct {
	// ID is the article ID
	ID int
	// Words is the sorted list of top words in the document.
	Words []uint16
}
