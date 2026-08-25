package query

import (
	"binary-parser/internal/format"
)

type Predicate func(rec format.Record, index int) bool

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
	noteFilterMemo(result)
	return result
}

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

func ByType(typ uint8) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.Type == typ
	}
}

func ByID(id uint32) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ID == id
	}
}

func ByIDRange(lo, hi uint32) Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ID >= lo && rec.ID <= hi
	}
}

func ByPayloadSize(minSize, maxSize int) Predicate {
	return func(rec format.Record, _ int) bool {
		n := len(rec.Payload)
		return n >= minSize && n <= maxSize
	}
}

func ByPayloadContains(sub []byte) Predicate {
	return func(rec format.Record, _ int) bool {
		return containsBytes(rec.Payload, sub)
	}
}

func ByCRCValid() Predicate {
	return func(rec format.Record, _ int) bool {
		return rec.ChecksumOK()
	}
}

func ByCRCInvalid() Predicate {
	return func(rec format.Record, _ int) bool {
		return !rec.ChecksumOK()
	}
}

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
