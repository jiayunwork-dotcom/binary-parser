package diff

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

func TestCompare_Identical(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(2, 2, "b"))
	r := Compare(c, c)
	if r.HasChanges() {
		t.Error("identical containers should have no changes")
	}
	if r.UnchangedCount != 2 {
		t.Errorf("expected 2 unchanged, got %d", r.UnchangedCount)
	}
}

func TestCompare_Added(t *testing.T) {
	left := container(rec(1, 1, "a"))
	right := container(rec(1, 1, "a"), rec(2, 2, "b"))
	r := Compare(left, right)
	if r.AddedCount != 1 {
		t.Errorf("expected 1 added, got %d", r.AddedCount)
	}
	added := OnlyAdded(r)
	if len(added) != 1 || added[0].ID != 2 {
		t.Errorf("added record should be id=2, got %v", added)
	}
}

func TestCompare_Removed(t *testing.T) {
	left := container(rec(1, 1, "a"), rec(2, 2, "b"))
	right := container(rec(1, 1, "a"))
	r := Compare(left, right)
	if r.RemovedCount != 1 {
		t.Errorf("expected 1 removed, got %d", r.RemovedCount)
	}
	removed := OnlyRemoved(r)
	if len(removed) != 1 || removed[0].ID != 2 {
		t.Errorf("removed record should be id=2, got %v", removed)
	}
}

func TestCompare_Modified(t *testing.T) {
	left := container(rec(1, 1, "hello"))
	right := container(rec(1, 1, "world"))
	r := Compare(left, right)
	if r.ModifiedCount != 1 {
		t.Errorf("expected 1 modified, got %d", r.ModifiedCount)
	}
	mods := OnlyModified(r)
	if len(mods) != 1 || mods[0].ID != 1 {
		t.Errorf("modified record should be id=1, got %v", mods)
	}
}

func TestCompare_TypeChange(t *testing.T) {
	left := container(rec(1, 5, "data"))
	right := container(rec(2, 5, "data"))
	r := Compare(left, right)
	if r.ModifiedCount != 1 {
		t.Errorf("type change should count as modified, got modified=%d", r.ModifiedCount)
	}
}

func TestCompare_BothNil(t *testing.T) {
	r := Compare(nil, nil)
	if r.HasChanges() {
		t.Error("both nil should have no changes")
	}
}

func TestCompare_LeftNil(t *testing.T) {
	right := container(rec(1, 1, "a"))
	r := Compare(nil, right)
	if r.AddedCount != 1 {
		t.Errorf("expected 1 added, got %d", r.AddedCount)
	}
}

func TestCompare_HeaderDiff(t *testing.T) {
	left := &format.Container{Header: format.Header{Version: 1, Count: 0}}
	right := &format.Container{Header: format.Header{Version: 2, Count: 0}}
	r := Compare(left, right)
	if !r.HeaderDiff.VersionChanged {
		t.Error("expected version change detected")
	}
}

func TestSummary_Identical(t *testing.T) {
	r := &Result{}
	s := Summary(r)
	if s != "containers are identical" {
		t.Errorf("unexpected summary: %s", s)
	}
}

func TestPayloadDiff_Same(t *testing.T) {
	off := PayloadDiff([]byte("hello"), []byte("hello"))
	if off != -1 {
		t.Errorf("expected -1, got %d", off)
	}
}

func TestPayloadDiff_DifferentByte(t *testing.T) {
	off := PayloadDiff([]byte("hello"), []byte("hxllo"))
	if off != 1 {
		t.Errorf("expected offset 1, got %d", off)
	}
}

func TestPayloadDiff_DifferentLength(t *testing.T) {
	off := PayloadDiff([]byte("hi"), []byte("hi!"))
	if off != 2 {
		t.Errorf("expected offset 2, got %d", off)
	}
}
