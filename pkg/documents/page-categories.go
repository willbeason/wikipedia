package documents

import (
	"sort"

	"github.com/gogo/protobuf/sortkeys"
)

func (x *PageCategories) Add(child, parent uint32) {
	if x.Pages == nil {
		x.Pages = make(map[uint32]*Categories)
	}

	page, exists := x.Pages[child]
	if !exists {
		page = &Categories{}
	}

	page.Add(parent)
	x.Pages[child] = page
}

func (x *Categories) Add(parent uint32) {
	sortkeys.Uint32s(x.Categories)

	pos := sort.Search(len(x.Categories), func(i int) bool {
		return x.Categories[i] >= parent
	})

	if pos < len(x.Categories) && x.Categories[pos] == parent {
		return
	}

	x.Categories = append(x.Categories, 0)
	copy(x.Categories[pos+1:], x.Categories[pos:])
	x.Categories[pos] = parent
}
