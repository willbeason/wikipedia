package classify

func (x *ClassifiedArticles) ToIDs() []uint {
	result := make([]uint, len(x.Articles))

	i := 0
	for k := range x.Articles {
		result[i] = uint(k)
		i++
	}

	return result
}
