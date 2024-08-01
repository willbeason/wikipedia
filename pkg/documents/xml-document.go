package documents

// XMLDocument solely exists for extracting from pages-articles-multistream.
// Individual XML documents from the compressed file slices may contain one
// or more Pages.
type XMLDocument struct {
	Pages []XMLPage `xml:"page"`
}

func (d *XMLDocument) ToProto() *Document {
	result := &Document{
		Pages: make([]*Page, len(d.Pages)),
	}

	for i, page := range d.Pages {
		result.Pages[i] = page.ToProto()
	}

	return result
}

type XMLPage struct {
	Title string `xml:"title"`
	// NS is the Wikipedia Namespace the page is categorized into.
	NS       Namespace   `xml:"ns"`
	ID       uint32      `xml:"id"`
	Redirect XMLRedirect `yaml:",omitempty" xml:"redirect"`
	Revision XMLRevision `xml:"revision"`
}

func (p *XMLPage) ToProto() *Page {
	return &Page{
		Id:    p.ID,
		Title: p.Title,
		Text:  p.Revision.Text,
	}
}

type XMLRedirect struct {
	Title string `yaml:",omitempty" xml:"title,attr"`
}

type XMLRevision struct {
	Text string `xml:"text"`
}
