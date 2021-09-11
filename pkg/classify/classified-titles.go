package classify

import "strings"

func (x *ClassifiedTitles) ToClassifiedIDs(idOf map[string]uint32) *ClassifiedIDs {
	result := &ClassifiedIDs{
		Pages: make(map[uint32]Classification),
	}

	for title, classification := range x.Pages {
		title = strings.ToLower(title)
		id, found := idOf[title]
		if !found {
			panic(title + " NOT FOUND")
		}
		result.Pages[id] = classification
	}

	return result
}
