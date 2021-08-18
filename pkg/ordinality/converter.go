package ordinality

import (
	"sort"
	"strings"

	"github.com/willbeason/extract-wikipedia/pkg/documents"
	"github.com/willbeason/extract-wikipedia/pkg/nlp"
)

type WordBagConverter struct {
	nlp.Tokenizer
	WordOrdinality
}

func (c *WordBagConverter) ToDocumentWordBag(d *documents.Document) *DocumentWordBag {
	result := &DocumentWordBag{}
	result.Pages = make([]*PageWordBag, len(d.Pages))

	for i, page := range d.Pages {
		result.Pages[i] = c.ToPageWordBag(page)
	}

	return result
}

func (c *WordBagConverter) ToPageWordBag(p documents.Page) *PageWordBag {
	result := &PageWordBag{
		Id: uint32(p.ID),
		Title: p.Title,
	}

	tokenCounts := make(map[uint32]uint32)
	for _, line := range strings.Split(p.Revision.Text, "\n") {
		for _, token := range c.Tokenizer.Tokenize(line) {
			tokenID := c.WordOrdinality[token]
			if tokenID == 0 {
				// This word is not in our dictionary.
				continue
			}

			tokenCounts[tokenID]++
		}
	}

	result.Words = make([]*WordCount, len(tokenCounts))

	i := 0
	for id, count := range tokenCounts {
		result.Words[i] = &WordCount{Word: id, Count: count}
		i++
	}

	sort.Slice(result.Words, func(i, j int) bool {
		return result.Words[i].Word < result.Words[j].Word
	})

	return result
}
