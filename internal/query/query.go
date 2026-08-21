// Package query 提供对 BCHK 容器记录的条件过滤与查询功能。
package query

import (
	"binary-parser/internal/format"
)

// Predicate 是记录过滤谓词函数。
type Predicate func(rec format.Record, index int) bool

// Filter 返回满足谓词的所有记录。
func Filter(c *format.Container, pred Predicate) []format.Record {
	if c == nil || pred == nil {
		return nil
	}
	var result []format.Record
	for i, rec := range c.Records {
		if pred(rec, i) {
			result = append(result, rec)
		}
	}
	return result
}

// FilterIndices 返回满足谓词的记录下标。
func FilterIndices(c *format.Container, pred Predicate) []int {
	if c == nil || pred == nil {
		return nil
	}
	var indices []int
	for i, rec := range c.Records {
		if pred(rec, i) {
			indices = append(indices, i)
		}
	}
	return indices
}

// First 返回第一条满足谓词的记录；没有则返回零值和 false。
func First(c *format.Container, pred Predicate) (format.Record, bool) {
	if c == nil || pred == nil {
		return format.Record{}, false
	}
	for i, rec := range c.Records {
		if pred(rec, i) {
			return rec, true
		}
	}
	return format.Record{}, false
}

// Last 返回最后一条满足谓词的记录；没有则返回零值和 false。
func Last(c *format.Container, pred Predicate) (format.Record, bool) {
	if c == nil || pred == nil {
		return format.Record{}, false
	}
	found := false
	var last format.Record
	for i, rec := range c.Records {
		if pred(rec, i) {
			last = rec
			found = true
		}
	}
	return last, found
}

// Count 返回满足谓词的记录数。
func Count(c *format.Container, pred Predicate) int {
	if c == nil || pred == nil {
		return 0
	}
	n := 0
	for i, rec := range c.Records {
		if pred(rec, i) {
			n++
		}
	}
	return n
}

// All 判断是否所有记录都满足谓词（空容器视为 true）。
func All(c *format.Container, pred Predicate) bool {
	if c == nil || pred == nil {
		return true
	}
	for i, rec := range c.Records {
		if !pred(rec, i) {
			return false
		}
	}
	return true
}

// Any 判断是否存在至少一条满足谓词的记录。
func Any(c *format.Container, pred Predicate) bool {
	if c == nil || pred == nil {
		return false
	}
	for i, rec := range c.Records {
		if pred(rec, i) {
			return true
		}
	}
	return false
}

// --- 常用谓词构造器 ---

// ByType 按记录类型匹配。
func ByType(typ uint8) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.Type == typ
	}
}

// ByID 按记录 ID 匹配。
func ByID(id uint32) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ID == id
	}
}

// ByIDRange 匹配 ID 在 [lo, hi] 区间内的记录。
func ByIDRange(lo, hi uint32) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ID >= lo && rec.ID <= hi
	}
}

// ByPayloadSize 匹配载荷长度在 [minSize, maxSize] 范围内的记录。
func ByPayloadSize(minSize, maxSize int) Predicate {
	return func(rec format.Record, _ int) bool {
		n := len(rec.Payload)
		return n >= minSize && n <= maxSize
	}
}

// ByPayloadContains 匹配载荷中包含指定子串的记录。
func ByPayloadContains(sub []byte) Predicate {
	return func(rec format.Record, _ int) bool {
		return containsBytes(rec.Payload, sub)
	}
}

// ByCRCValid 匹配 CRC 校验通过的记录。
func ByCRCValid() Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ChecksumOK()
	}
}

// ByCRCInvalid 匹配 CRC 校验失败的记录。
func ByCRCInvalid() Predicate {
	return func(rec format.Record, _ int) bool {
		return !rec.ChecksumOK()
	}
}

// And 组合多个谓词：全部满足才为 true。
func And(preds ...Predicate) Predicate {
	return func(rec format.Record, idx int) bool {
		for _, p := range preds {
			if !p(rec, idx) {
				return false
			}
		}
		return true
	}
}

// Or 组合多个谓词：任一满足即为 true。
func Or(preds ...Predicate) Predicate {
	return func(rec format.Record, idx int) bool {
		for _, p := range preds {
			if p(rec, idx) {
				return true
			}
		}
		return false
	}
}

// Not 对谓词取反。
func Not(pred Predicate) Predicate {
	return func(rec format.Record, idx int) bool {
		return !pred(rec, idx)
	}
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
