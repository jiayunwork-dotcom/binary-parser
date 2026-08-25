package index

import "binary-parser/internal/format"

type Index struct {
	ByType map[uint8][]int
	ByID   map[uint32]int
}

func Build(c *format.Container) *Index {
	idx := &Index{ByType: map[uint8][]int{}, ByID: map[uint32]int{}}
	for i, rec := range c.Records {
		idx.ByType[rec.Type] = append(idx.ByType[rec.Type], i)
		idx.ByID[rec.ID] = i
	}
	return idx
}

func (idx *Index) Lookup(id uint32) (int, bool) {
	if idx == nil || idx.ByID == nil {
		return -1, false
	}
	pos, ok := idx.ByID[id]
	if !ok {
		return -1, false
	}
	return pos, true
}
