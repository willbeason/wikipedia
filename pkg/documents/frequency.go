package documents

type FrequencyTable struct {
	Frequencies []Frequency
}

type Frequency struct {
	Word  string
	Count int
}
