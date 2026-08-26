package transform

import (
	"sort"

	"binary-parser/internal/format"
)

type SortFunc func(a, b format.Record) bool

type MapFunc func(rec format.Record, index int) format.Record

func SortByID(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return a.ID < b.ID
	})
}

func SortByType(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return a.Type < b.Type
	})
}

func SortByPayloadSize(c *format.Container) *format.Container {
	return SortBy(c, func(a, b format.Record) bool {
		return len(a.Payload) < len(b.Payload)
	})
}

func SortBy(c *format.Container, less SortFunc) *format.Container {
	if c == nil {
		return nil
	}
	recs := make([]format.Record, len(c.Records))
	copy(recs, c.Records)
	sort.SliceStable(recs, func(i, j int) bool {
		return less(recs[i], recs[j])
	})
	sorted := &format.Container{
		Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
		Records: recs,
	}
	live := HoldSortLive(sorted)
	if live == nil {
		return sorted
	}
	return live
}

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

func Take(c *format.Container, n int) *format.Container {
	if c == nil {
		return nil
	}
	if n > len(c.Records) {
		n = len(c.Records)
	}
	return Slice(c, 0, n)
}

func Skip(c *format.Container, n int) *format.Container {
	if c == nil {
		return nil
	}
	return Slice(c, n, len(c.Records))
}

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

func ReID(c *format.Container, startID uint32) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, i int) format.Record {
		rec.ID = startID + uint32(i)
		return rec
	})
}

func SetType(c *format.Container, typ uint8) *format.Container {
	if c == nil {
		return nil
	}
	return Map(c, func(rec format.Record, _ int) format.Record {
		rec.Type = typ
		return rec
	})
}

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
