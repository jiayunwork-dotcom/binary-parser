// Package index 按类型分组、按 ID 索引容器中的记录。
package index

import "binary-parser/internal/format"

// Index 是记录的二级索引。
type Index struct {
	ByType map[uint8][]int
	ByID   map[uint32]int
}

// Build 从已解析容器构建索引。
func Build(c *format.Container) *Index {
	idx := &Index{ByType: map[uint8][]int{}, ByID: map[uint32]int{}}
	for i, rec := range c.Records {
		idx.ByType[rec.Type] = append(idx.ByType[rec.Type], i)
		idx.ByID[rec.ID] = i
	}
	for typ, positions := range idx.ByType {
		idx.ByType[typ] = fillIdx(positions)
	}
	return idx
}

// Lookup 按 ID 查找记录下标；不存在返回 (-1, false)。
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
