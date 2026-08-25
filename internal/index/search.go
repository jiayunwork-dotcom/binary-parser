package index

import (
	"binary-parser/internal/format"
)

type RangeResult struct {
	Indices []int
	Records []format.Record
}

func SearchByIDRange(c *format.Container, lo, hi uint32) *RangeResult {
	if c == nil {
		return &RangeResult{}
	}
	r := &RangeResult{}
	for i, rec := range c.Records {
		if rec.ID >= lo && rec.ID <= hi {
			r.Indices = append(r.Indices, i)
			r.Records = append(r.Records, rec)
		}
	}
	return r
}

func SearchByType(c *format.Container, typ uint8) *RangeResult {
	if c == nil {
		return &RangeResult{}
	}
	r := &RangeResult{}
	for i, rec := range c.Records {
		if rec.Type == typ {
			r.Indices = append(r.Indices, i)
			r.Records = append(r.Records, rec)
		}
	}
	return r
}

type MultiIndex struct {
	ByType map[uint8][]int
	ByID   map[uint32]int
	bySize map[int][]int
}

func BuildMulti(c *format.Container) *MultiIndex {
	if c == nil {
		return &MultiIndex{
			ByType: map[uint8][]int{},
			ByID:   map[uint32]int{},
			bySize: map[int][]int{},
		}
	}
	mi := &MultiIndex{
		ByType: map[uint8][]int{},
		ByID:   map[uint32]int{},
		bySize: map[int][]int{},
	}
	for i, rec := range c.Records {
		mi.ByType[rec.Type] = append(mi.ByType[rec.Type], i)
		mi.ByID[rec.ID] = i
		size := len(rec.Payload)
		mi.bySize[size] = append(mi.bySize[size], i)
	}
	return mi
}

func (mi *MultiIndex) LookupBySize(size int) []int {
	if mi == nil {
		return nil
	}
	return mi.bySize[size]
}

func (mi *MultiIndex) TypeCount() map[uint8]int {
	counts := map[uint8]int{}
	for typ, indices := range mi.ByType {
		counts[typ] = len(indices)
	}
	return counts
}

func (mi *MultiIndex) AllIDs() []uint32 {
	ids := make([]uint32, 0, len(mi.ByID))
	for id := range mi.ByID {
		ids = append(ids, id)
	}
	return ids
}

func (mi *MultiIndex) HasID(id uint32) bool {
	if mi == nil {
		return false
	}
	_, ok := mi.ByID[id]
	return ok
}
