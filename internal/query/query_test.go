package query

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

func TestFilter_ByType(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"), rec(1, 3, "c"))
	result := Filter(c, ByType(1))
	if len(result) != 2 {
		t.Errorf("expected 2 records of type=1, got %d", len(result))
	}
}

func TestFilter_ByID(t *testing.T) {
	c := container(rec(1, 10, "a"), rec(2, 20, "b"), rec(3, 30, "c"))
	result := Filter(c, ByID(20))
	if len(result) != 1 || result[0].ID != 20 {
		t.Errorf("expected record id=20, got %v", result)
	}
}

func TestFilter_ByIDRange(t *testing.T) {
	c := container(rec(1, 5, "a"), rec(1, 10, "b"), rec(1, 15, "c"), rec(1, 20, "d"))
	result := Filter(c, ByIDRange(8, 18))
	if len(result) != 2 {
		t.Errorf("expected 2 records in [8,18], got %d", len(result))
	}
}

func TestFilter_ByPayloadSize(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "abc"), rec(1, 3, "abcde"))
	result := Filter(c, ByPayloadSize(2, 4))
	if len(result) != 1 || result[0].ID != 2 {
		t.Errorf("expected 1 record with payload size in [2,4], got %v", result)
	}
}

func TestFilter_ByPayloadContains(t *testing.T) {
	c := container(rec(1, 1, "hello world"), rec(1, 2, "goodbye"), rec(1, 3, "hello go"))
	result := Filter(c, ByPayloadContains([]byte("hello")))
	if len(result) != 2 {
		t.Errorf("expected 2 records containing 'hello', got %d", len(result))
	}
}

func TestFilter_NilContainer(t *testing.T) {
	result := Filter(nil, ByType(1))
	if result != nil {
		t.Errorf("expected nil for nil container, got %v", result)
	}
}

func TestFilterIndices(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"), rec(1, 3, "c"))
	indices := FilterIndices(c, ByType(1))
	if len(indices) != 2 || indices[0] != 0 || indices[1] != 2 {
		t.Errorf("expected [0, 2], got %v", indices)
	}
}

func TestFirst(t *testing.T) {
	c := container(rec(1, 10, "a"), rec(2, 20, "b"), rec(2, 30, "c"))
	r, ok := First(c, ByType(2))
	if !ok || r.ID != 20 {
		t.Errorf("expected first type=2 at id=20, got id=%d ok=%v", r.ID, ok)
	}
}

func TestFirst_NotFound(t *testing.T) {
	c := container(rec(1, 1, "a"))
	_, ok := First(c, ByType(9))
	if ok {
		t.Error("expected not found")
	}
}

func TestLast(t *testing.T) {
	c := container(rec(2, 10, "a"), rec(2, 20, "b"), rec(1, 30, "c"))
	r, ok := Last(c, ByType(2))
	if !ok || r.ID != 20 {
		t.Errorf("expected last type=2 at id=20, got id=%d", r.ID)
	}
}

func TestCount(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(2, 3, "c"))
	if n := Count(c, ByType(1)); n != 2 {
		t.Errorf("expected count=2, got %d", n)
	}
}

func TestAll(t *testing.T) {
	c := container(rec(1, 1, "ab"), rec(1, 2, "cd"), rec(1, 3, "ef"))
	if !All(c, ByPayloadSize(1, 10)) {
		t.Error("expected All true for payload [1,10]")
	}
	if All(c, ByType(2)) {
		t.Error("expected All false for type=2")
	}
}

func TestAny(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"))
	if !Any(c, ByType(2)) {
		t.Error("expected Any true for type=2")
	}
	if Any(c, ByType(9)) {
		t.Error("expected Any false for type=9")
	}
}

func TestAnd(t *testing.T) {
	c := container(rec(1, 5, "abc"), rec(1, 15, "a"), rec(2, 8, "abcde"))
	pred := And(ByType(1), ByPayloadSize(2, 10))
	result := Filter(c, pred)
	if len(result) != 1 || result[0].ID != 5 {
		t.Errorf("And: expected id=5, got %v", result)
	}
}

func TestOr(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"), rec(3, 3, "c"))
	pred := Or(ByType(1), ByType(3))
	result := Filter(c, pred)
	if len(result) != 2 {
		t.Errorf("Or: expected 2 records, got %d", len(result))
	}
}

func TestNot(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"), rec(1, 3, "c"))
	result := Filter(c, Not(ByType(1)))
	if len(result) != 1 || result[0].ID != 2 {
		t.Errorf("Not: expected id=2 only, got %v", result)
	}
}
