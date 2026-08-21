package index

import (
	"testing"

	"binary-parser/internal/format"
)

func makeContainer(recs ...format.Record) *format.Container {
	return &format.Container{
		Header:  format.Header{Version: 1, Count: uint16(len(recs))},
		Records: recs,
	}
}

func r(typ uint8, id uint32, payload string) format.Record {
	return format.Record{Type: typ, ID: id, Payload: []byte(payload)}
}

func TestSearchByIDRange(t *testing.T) {
	c := makeContainer(r(1, 5, "a"), r(1, 10, "b"), r(1, 15, "c"), r(1, 20, "d"))
	result := SearchByIDRange(c, 8, 18)
	if len(result.Records) != 2 {
		t.Errorf("expected 2 records in [8,18], got %d", len(result.Records))
	}
}

func TestSearchByType(t *testing.T) {
	c := makeContainer(r(1, 1, "a"), r(2, 2, "b"), r(1, 3, "c"))
	result := SearchByType(c, 1)
	if len(result.Records) != 2 {
		t.Errorf("expected 2 records of type=1, got %d", len(result.Records))
	}
}

func TestBuildMulti(t *testing.T) {
	c := makeContainer(r(1, 1, "ab"), r(2, 2, "cde"), r(1, 3, "ab"))
	mi := BuildMulti(c)
	if len(mi.ByType[1]) != 2 {
		t.Errorf("type=1 expected 2, got %d", len(mi.ByType[1]))
	}
	if !mi.HasID(2) {
		t.Error("expected HasID(2) true")
	}
	if mi.HasID(99) {
		t.Error("expected HasID(99) false")
	}
}

func TestMultiIndex_LookupBySize(t *testing.T) {
	c := makeContainer(r(1, 1, "ab"), r(1, 2, "cde"), r(1, 3, "fg"))
	mi := BuildMulti(c)
	size2 := mi.LookupBySize(2)
	if len(size2) != 2 {
		t.Errorf("expected 2 records with payload size=2, got %d", len(size2))
	}
}

func TestMultiIndex_TypeCount(t *testing.T) {
	c := makeContainer(r(1, 1, "a"), r(2, 2, "b"), r(2, 3, "c"))
	mi := BuildMulti(c)
	counts := mi.TypeCount()
	if counts[1] != 1 || counts[2] != 2 {
		t.Errorf("type counts wrong: %v", counts)
	}
}

func TestMultiIndex_AllIDs(t *testing.T) {
	c := makeContainer(r(1, 10, "a"), r(1, 20, "b"), r(1, 30, "c"))
	mi := BuildMulti(c)
	ids := mi.AllIDs()
	if len(ids) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(ids))
	}
}
