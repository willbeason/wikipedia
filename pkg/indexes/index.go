package indexes

import (
	"sort"
)

// Find determines the file containing an article of the given identifier.
func (x *Index) Find(pageId uint32) string {
	idx := sort.Search(len(x.Entries), func(i int) bool {
		return x.Entries[i].Max > pageId
	})

	return x.Entries[idx].File
}
