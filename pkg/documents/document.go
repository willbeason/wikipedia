package documents

type Document struct {
	Path string `yaml:",omitempty"`

	Pages []Page `xml:"page"`
}

type Page struct {
	ID       int      `xml:"id"`
	Title    string   `xml:"title"`
	Redirect Redirect `yaml:",omitempty" xml:"redirect"`
	Revision Revision `xml:"revision"`
}

type Redirect struct {
	Title string `yaml:",omitempty" xml:"title,attr"`
}

type Revision struct {
	Text string `xml:"text"`
}
