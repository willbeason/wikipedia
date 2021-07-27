package documents

type Document struct {
	Pages []Page `xml:"page"`
}

type Page struct {
	ID       int      `xml:"id"`
	Title    string   `xml:"title"`
	Redirect Redirect `xml:"redirect"`
	Revision Revision `xml:"revision"`
}

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Revision struct {
	Text string `xml:"text"`
}
