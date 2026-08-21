// Package transform 实现 BCHK 容器记录的变换操作：排序、切片、去重、映射。
package transform

import (
	"sort"

	"binary-parser/internal/format"
)

// SortFunc 是记录排序的比较函数。返回 true 表示 a 排在 b 前面。
type SortFunc func(a, b format.Record) bool

// MapFunc 对单条记录执行变换，返回变换后的记录。
type MapFunc func(rec format.Record, index int) format.Record

// SortByID 按 ID 升序排序记录（不改变原容器，返回新容器）。
func SortByID(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return a.ID < b.ID
	})
}

// SortByType 按类型升序排序。
func SortByType(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return a.Type < b.Type
	})
}

// SortByPayloadSize 按载荷长度升序排序。
func SortByPayloadSize(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return len(a.Payload) < len(b.Payload)
	})
}

// SortBy 使用自定义比较函数排序。
func SortBy(c *format.Container, less SortFunc) *format.Container {
	if c == nil {
		return nil
	}
	recs := make([]format.Record, len(c.Records))
	copy(recs, c.Records)
	sort.SliceStable(recs, func(i, j int) bool {
		return less(recs[i], recs[j])
	})
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
}

// Reverse 反转记录顺序。
func Reverse(c *format.Container) *format.Container {
	if c == nil {
		return nil
	}
	n := len(c.Records)
	recs := make([]format.Record, n)
	for i, rec := range c.Records {
		recs[n-1-i] = rec
	}
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(n)},
		Records: recs,
	}
}

// Slice 截取记录的子区间 [start, end)。
func Slice(c *format.Container, start, end int) *format.Container {
	if c == nil {
		return nil
	}
	n := len(c.Records)
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if start >= end {
		return &format.Container{Header: format.Header{Version: c.Header.Version, Count: 0}}
	}
	recs := make([]format.Record, end-start)
	copy(recs, c.Records[start:end])
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
}

// Take 取前 n 条记录。
func Take(c *format.Container, n int) *format.Container {
	if c == nil {
		return nil
	}
	if n > len(c.Records) {
		n = len(c.Records)
	}
	return Slice(c, 0, n)
}

// Skip 跳过前 n 条记录。
func Skip(c *format.Container, n int) *format.Container {
	if c == nil {
		return nil
	}
	return Slice(c, n, len(c.Records))
}

// Dedup 按 ID 去重，保留每个 ID 第一次出现的记录。
func Dedup(c *format.Container) *format.Container {
	if c == nil {
		return nil
	}
	seen := map[uint32]bool{}
	var recs []format.Record
	for _, rec := range c.Records {
		if !seen[rec.ID] {
			seen[rec.ID] = true
			recs = append(recs, rec)
		}
	}
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
}

// DedupByKey 按自定义 key 去重，保留每个 key 首次出现的记录。
func DedupByKey(c *format.Container, keyFn func(format.Record) uint64) *format.Container {
	if c == nil {
		return nil
	}
	seen := map[uint64]bool{}
	var recs []format.Record
	for _, rec := range c.Records {
		k := keyFn(rec)
		if !seen[k] {
			seen[k] = true
			recs = append(recs, rec)
		}
	}
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
}

// Map 对每条记录执行变换，返回新容器。
func Map(c *format.Container, fn MapFunc) *format.Container {
	if c == nil || fn == nil {
		return nil
	}
	recs := make([]format.Record, len(c.Records))
	for i, rec := range c.Records {
		recs[i] = fn(rec, i)
	}
	return &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
}

// ReID 为所有记录重新分配从 startID 开始的连续 ID。
func ReID(c *format.Container, startID uint32) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, i int) format.Record {
		rec.ID = startID + uint32(i)
		return rec
	})
}

// SetType 将所有记录的类型统一设置为指定值。
func SetType(c *format.Container, typ uint8) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, _ int) format.Record {
		rec.Type = typ
		return rec
	})
}

// TruncatePayload 截断载荷到指定最大长度。
func TruncatePayload(c *format.Container, maxLen int) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, _ int) format.Record {
		if len(rec.Payload) > maxLen {
			truncated := make([]byte, maxLen)
			copy(truncated, rec.Payload[:maxLen])
			rec.Payload = truncated
		}
		return rec
	})
}

// PadPayload 将短于 minLen 的载荷用指定字节填充到 minLen。
func PadPayload(c *format.Container, minLen int, padByte byte) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, _ int) format.Record {
		if len(rec.Payload) < minLen {
			padded := make([]byte, minLen)
			copy(padded, rec.Payload)
			for i := len(rec.Payload); i < minLen; i++ {
				padded[i] = padByte
			}
			rec.Payload = padded
		}
		return rec
	})
}
