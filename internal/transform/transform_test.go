package transform

import (
	"testing"

	"binary-parser/internal/format"
)

func container(recs ...format.Record) *format.Container {
	return &format.Container{
		Header:  format.Header{Version: 1, Count: uint16(len(recs))},
		Records: recs,
	}
}

func rec(typ uint8, id uint32, payload string) format.Record {
	return format.Record{Type: typ, ID: id, Payload: []byte(payload)}
}

func TestSortByID(t *testing.T) {
	c := container(rec(1, 30, "c"), rec(1, 10, "a"), rec(1, 20, "b"))
	sorted := SortByID(c)
	if sorted.Records[0].ID != 10 || sorted.Records[1].ID != 20 || sorted.Records[2].ID != 30 {
		t.Errorf("sort by ID failed: %v", sorted.Records)
	}
	if c.Records[0].ID != 30 {
		t.Error("original container modified")
	}
}

func TestSortByType(t *testing.T) {
	c := container(rec(3, 1, "a"), rec(1, 2, "b"), rec(2, 3, "c"))
	sorted := SortByType(c)
	if sorted.Records[0].Type != 1 || sorted.Records[1].Type != 2 || sorted.Records[2].Type != 3 {
		t.Errorf("sort by type failed: %v", sorted.Records)
	}
}

func TestSortByPayloadSize(t *testing.T) {
	c := container(rec(1, 1, "abcde"), rec(1, 2, "a"), rec(1, 3, "abc"))
	sorted := SortByPayloadSize(c)
	if len(sorted.Records[0].Payload) != 1 || len(sorted.Records[2].Payload) != 5 {
		t.Error("sort by payload size failed")
	}
}

func TestReverse(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(1, 3, "c"))
	rev := Reverse(c)
	if rev.Records[0].ID != 3 || rev.Records[1].ID != 2 || rev.Records[2].ID != 1 {
		t.Errorf("reverse failed: %v", rev.Records)
	}
}

func TestSlice(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(1, 3, "c"), rec(1, 4, "d"))
	s := Slice(c, 1, 3)
	if len(s.Records) != 2 || s.Records[0].ID != 2 || s.Records[1].ID != 3 {
		t.Errorf("slice [1,3) failed: %v", s.Records)
	}
}

func TestSlice_OutOfBounds(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"))
	s := Slice(c, -5, 100)
	if len(s.Records) != 2 {
		t.Errorf("out of bounds slice should clamp: got %d records", len(s.Records))
	}
}

func TestSlice_Empty(t *testing.T) {
	c := container(rec(1, 1, "a"))
	s := Slice(c, 5, 10)
	if len(s.Records) != 0 {
		t.Errorf("empty slice expected, got %d records", len(s.Records))
	}
}

func TestTake(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(1, 3, "c"))
	taken := Take(c, 2)
	if len(taken.Records) != 2 {
		t.Errorf("take 2 expected 2, got %d", len(taken.Records))
	}
}

func TestSkip(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(1, 3, "c"))
	skipped := Skip(c, 1)
	if len(skipped.Records) != 2 || skipped.Records[0].ID != 2 {
		t.Errorf("skip 1 failed: %v", skipped.Records)
	}
}

func TestDedup(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 1, "b"), rec(1, 2, "c"), rec(3, 2, "d"))
	deduped := Dedup(c)
	if len(deduped.Records) != 2 {
		t.Errorf("dedup expected 2 records, got %d", len(deduped.Records))
	}
	if deduped.Records[0].Type != 1 || deduped.Records[1].Type != 1 {
		t.Error("dedup should keep first occurrence")
	}
}

func TestDedupByKey(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(2, 3, "c"), rec(2, 4, "d"))
	deduped := DedupByKey(c, func(r format.Record) uint64 { return uint64(r.Type) })
	if len(deduped.Records) != 2 {
		t.Errorf("expected 2 records (one per type), got %d", len(deduped.Records))
	}
}

func TestMap(t *testing.T) {
	c := container(rec(1, 1, "hello"), rec(2, 2, "world"))
	mapped := Map(c, func(r format.Record, _ int) format.Record {
		r.Type = r.Type + 10
		return r
	})
	if mapped.Records[0].Type != 11 || mapped.Records[1].Type != 12 {
		t.Error("map type+10 failed")
	}
}

func TestReID(t *testing.T) {
	c := container(rec(1, 99, "a"), rec(1, 88, "b"), rec(1, 77, "c"))
	reIDed := ReID(c, 100)
	for i, r := range reIDed.Records {
		if r.ID != uint32(100+i) {
			t.Errorf("record[%d] expected id=%d, got %d", i, 100+i, r.ID)
		}
	}
}

func TestSetType(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"), rec(3, 3, "c"))
	unified := SetType(c, 5)
	for i, r := range unified.Records {
		if r.Type != 5 {
			t.Errorf("record[%d] expected type=5, got %d", i, r.Type)
		}
	}
}

func TestTruncatePayload(t *testing.T) {
	c := container(rec(1, 1, "abcdefgh"), rec(1, 2, "ab"))
	truncated := TruncatePayload(c, 4)
	if len(truncated.Records[0].Payload) != 4 {
		t.Errorf("expected payload len=4, got %d", len(truncated.Records[0].Payload))
	}
	if string(truncated.Records[0].Payload) != "abcd" {
		t.Errorf("expected 'abcd', got %q", truncated.Records[0].Payload)
	}
	if len(truncated.Records[1].Payload) != 2 {
		t.Errorf("short payload should not change: got len=%d", len(truncated.Records[1].Payload))
	}
}

func TestPadPayload(t *testing.T) {
	c := container(rec(1, 1, "ab"), rec(1, 2, "abcde"))
	padded := PadPayload(c, 5, 0x00)
	if len(padded.Records[0].Payload) != 5 {
		t.Errorf("expected padded len=5, got %d", len(padded.Records[0].Payload))
	}
	if padded.Records[0].Payload[2] != 0x00 {
		t.Error("pad byte should be 0x00")
	}
	if len(padded.Records[1].Payload) != 5 {
		t.Errorf("long payload should stay: got len=%d", len(padded.Records[1].Payload))
	}
}

func TestNilContainer(t *testing.T) {
	if SortByID(nil) != nil {
		t.Error("SortByID(nil) should be nil")
	}
	if Reverse(nil) != nil {
		t.Error("Reverse(nil) should be nil")
	}
	if Slice(nil, 0, 1) != nil {
		t.Error("Slice(nil) should be nil")
	}
	if Dedup(nil) != nil {
		t.Error("Dedup(nil) should be nil")
	}
	if Map(nil, func(r format.Record, _ int) format.Record { return r }) != nil {
		t.Error("Map(nil) should be nil")
	}
}
