package merge

import (
	"testing"

	"binary-parser/internal/format"
)

func container(version uint16, recs ...format.Record) *format.Container {
	return &format.Container{
		Header:  format.Header{Version: version, Count: uint16(len(recs))},
		Records: recs,
	}
}

func rec(typ uint8, id uint32) format.Record {
	return format.Record{Type: typ, ID: id, Payload: []byte("data")}
}

func TestMerge_KeepAll(t *testing.T) {
	a := container(1, rec(1, 1), rec(1, 2))
	b := container(1, rec(2, 2), rec(2, 3))
	out, err := MergeTwo(a, b, nil)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(out.Records) != 4 {
		t.Errorf("expected 4 records, got %d", len(out.Records))
	}
}

func TestMerge_KeepFirst(t *testing.T) {
	a := container(1, rec(1, 1), rec(1, 2))
	b := container(1, rec(2, 2), rec(2, 3))
	opts := &Options{Strategy: StrategyKeepFirst}
	out, err := MergeTwo(a, b, opts)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(out.Records) != 3 {
		t.Errorf("expected 3 records (id=2 kept from a), got %d", len(out.Records))
	}
	for _, r := range out.Records {
		if r.ID == 2 && r.Type != 1 {
			t.Errorf("id=2 should be type=1 (from first), got type=%d", r.Type)
		}
	}
}

func TestMerge_KeepLast(t *testing.T) {
	a := container(1, rec(1, 1), rec(1, 2))
	b := container(1, rec(2, 2), rec(2, 3))
	opts := &Options{Strategy: StrategyKeepLast}
	out, err := MergeTwo(a, b, opts)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(out.Records) != 3 {
		t.Errorf("expected 3 records, got %d", len(out.Records))
	}
	for _, r := range out.Records {
		if r.ID == 2 && r.Type != 2 {
			t.Errorf("id=2 should be type=2 (from last), got type=%d", r.Type)
		}
	}
}

func TestMerge_StrategyError(t *testing.T) {
	a := container(1, rec(1, 1))
	b := container(1, rec(2, 1))
	opts := &Options{Strategy: StrategyError}
	_, err := MergeTwo(a, b, opts)
	if err == nil {
		t.Fatal("expected error for duplicate id with StrategyError")
	}
}

func TestMerge_StrictVersion(t *testing.T) {
	a := container(1, rec(1, 1))
	b := container(2, rec(2, 2))
	opts := &Options{StrictVersion: true}
	_, err := MergeTwo(a, b, opts)
	if err == nil {
		t.Fatal("expected version conflict error")
	}
}

func TestMerge_NoContainers(t *testing.T) {
	_, err := Merge(nil, nil)
	if err != ErrNoContainers {
		t.Errorf("expected ErrNoContainers, got %v", err)
	}
}

func TestSplit_Basic(t *testing.T) {
	c := container(1, rec(1, 1), rec(1, 2), rec(1, 3), rec(1, 4), rec(1, 5))
	chunks, err := Split(c, 2)
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}
	if len(chunks[0].Records) != 2 {
		t.Errorf("chunk[0] expected 2 records, got %d", len(chunks[0].Records))
	}
	if len(chunks[2].Records) != 1 {
		t.Errorf("chunk[2] expected 1 record, got %d", len(chunks[2].Records))
	}
}

func TestSplit_ZeroSize(t *testing.T) {
	c := container(1, rec(1, 1))
	_, err := Split(c, 0)
	if err != ErrSplitZero {
		t.Errorf("expected ErrSplitZero, got %v", err)
	}
}

func TestSplitByType(t *testing.T) {
	c := container(1, rec(1, 1), rec(2, 2), rec(1, 3), rec(3, 4))
	result, err := SplitByType(c)
	if err != nil {
		t.Fatalf("SplitByType: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 types, got %d", len(result))
	}
	if len(result[1].Records) != 2 {
		t.Errorf("type=1 expected 2 records, got %d", len(result[1].Records))
	}
}

func TestTotalRecords(t *testing.T) {
	a := container(1, rec(1, 1), rec(1, 2))
	b := container(1, rec(2, 3))
	n := TotalRecords([]*format.Container{a, b, nil})
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}
