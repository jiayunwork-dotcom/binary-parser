package index

import (
	"binary-parser/internal/format"
)

// RangeResult 是范围查询结果。
type RangeResult struct {
	Indices []int
	Records []format.Record
}

// SearchByIDRange 返回 ID 在 [lo, hi] 范围内的所有记录。
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

// SearchByType 返回指定类型的所有记录。
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

// MultiIndex 支持同时按 ID 和 Type 双重索引。
type MultiIndex struct {
	ByType map[uint8][]int
	ByID   map[uint32]int
	bySize map[int][]int // 按载荷大小分组
}

// BuildMulti 构建多维索引。
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

// LookupBySize 按载荷精确大小查找记录下标。
func (mi *MultiIndex) LookupBySize(size int) []int {
	if mi == nil {
		return nil
	}
	return mi.bySize[size]
}

// TypeCount 返回该索引中各类型的记录数量。
func (mi *MultiIndex) TypeCount() map[uint8]int {
	counts := map[uint8]int{}
	for typ, indices := range mi.ByType {
		counts[typ] = len(indices)
	}
	return counts
}

// AllIDs 返回索引中所有已知 ID（无序）。
func (mi *MultiIndex) AllIDs() []uint32 {
	ids := make([]uint32, 0, len(mi.ByID))
	for id := range mi.ByID {
		ids = append(ids, id)
	}
	return ids
}

// HasID 检查 ID 是否在索引中。
func (mi *MultiIndex) HasID(id uint32) bool {
	if mi == nil {
		return false
	}
	_, ok := mi.ByID[id]
	return ok
}
