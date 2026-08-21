package validate

import (
	"testing"

	"binary-parser/internal/format"
)

func goodContainer() *format.Container {
	return &format.Container{
		Header: format.Header{Version: 1, Count: 2},
		Records: []format.Record{
			{Type: 1, ID: 1, Payload: []byte("abc")},
			{Type: 2, ID: 2, Payload: []byte("def")},
		},
	}
}

func TestValidate_GoodContainer(t *testing.T) {
	c := goodContainer()
	opts := &Options{AllowDuplicateIDs: true} // 跳过重复检查简化
	r := Validate(c, opts)
	for _, is := range r.Issues {
		if is.Code != "E010" && is.Severity == SevError {
			t.Errorf("unexpected error: %v", is)
		}
	}
}

func TestValidate_NilContainer(t *testing.T) {
	r := Validate(nil, nil)
	if !r.HasErrors() {
		t.Fatal("expected error for nil container")
	}
	if r.Issues[0].Code != "E001" {
		t.Errorf("expected E001, got %s", r.Issues[0].Code)
	}
}

func TestValidate_CountMismatch(t *testing.T) {
	c := &format.Container{
		Header:  format.Header{Version: 1, Count: 5},
		Records: []format.Record{{Type: 1, ID: 1, Payload: []byte("x")}},
	}
	r := Validate(c, nil)
	found := false
	for _, is := range r.Issues {
		if is.Code == "E002" {
			found = true
		}
	}
	if !found {
		t.Error("expected E002 for count mismatch")
	}
}

func TestValidate_DuplicateIDs(t *testing.T) {
	c := &format.Container{
		Header: format.Header{Version: 1, Count: 2},
		Records: []format.Record{
			{Type: 1, ID: 7, Payload: []byte("a")},
			{Type: 2, ID: 7, Payload: []byte("b")},
		},
	}
	r := Validate(c, nil)
	found := false
	for _, is := range r.Issues {
		if is.Code == "E020" {
			found = true
		}
	}
	if !found {
		t.Error("expected E020 for duplicate IDs")
	}
}

func TestValidate_MaxPayloadSize(t *testing.T) {
	c := &format.Container{
		Header:  format.Header{Version: 1, Count: 1},
		Records: []format.Record{{Type: 1, ID: 1, Payload: make([]byte, 100)}},
	}
	opts := &Options{MaxPayloadSize: 50, AllowDuplicateIDs: true}
	r := Validate(c, opts)
	found := false
	for _, is := range r.Issues {
		if is.Code == "E011" {
			found = true
		}
	}
	if !found {
		t.Error("expected E011 for oversized payload")
	}
}

func TestValidate_SequentialIDs(t *testing.T) {
	c := &format.Container{
		Header: format.Header{Version: 1, Count: 3},
		Records: []format.Record{
			{Type: 1, ID: 1, Payload: []byte("a")},
			{Type: 1, ID: 3, Payload: []byte("b")}, // gap
			{Type: 1, ID: 4, Payload: []byte("c")},
		},
	}
	opts := &Options{RequireSequentialIDs: true, AllowDuplicateIDs: true}
	r := Validate(c, opts)
	found := false
	for _, is := range r.Issues {
		if is.Code == "W020" {
			found = true
		}
	}
	if !found {
		t.Error("expected W020 for non-sequential IDs")
	}
}

func TestSummary_NoIssues(t *testing.T) {
	r := &Report{}
	s := Summary(r)
	if s != "OK: no issues found" {
		t.Errorf("unexpected summary: %s", s)
	}
}

func TestSortByIndex(t *testing.T) {
	issues := []Issue{
		{Index: 5, Code: "A"},
		{Index: 1, Code: "B"},
		{Index: 3, Code: "C"},
	}
	SortByIndex(issues)
	if issues[0].Index != 1 || issues[1].Index != 3 || issues[2].Index != 5 {
		t.Errorf("sort failed: %v", issues)
	}
}
