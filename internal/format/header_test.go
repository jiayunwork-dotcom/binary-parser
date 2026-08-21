package format

import (
	"strings"
	"testing"
)

func TestContainerSize(t *testing.T) {
	c := &Container{
		Header: Header{Version: 1, Count: 2},
		Records: []Record{
			{Type: 1, ID: 1, Payload: []byte("ab")},
			{Type: 2, ID: 2, Payload: []byte("cde")},
		},
	}
	// 8 + (13 + 2) + (13 + 3) = 8 + 15 + 16 = 39
	expected := HeaderSize + (RecordOverhead + 2) + (RecordOverhead + 3)
	got := ContainerSize(c)
	if got != expected {
		t.Errorf("ContainerSize: got %d, want %d", got, expected)
	}
}

func TestContainerSize_Nil(t *testing.T) {
	if ContainerSize(nil) != 0 {
		t.Error("nil container size should be 0")
	}
}

func TestRecordSize(t *testing.T) {
	rec := Record{Type: 1, ID: 1, Payload: []byte("hello")}
	expected := RecordOverhead + 5
	if RecordSize(rec) != expected {
		t.Errorf("RecordSize: got %d, want %d", RecordSize(rec), expected)
	}
}

func TestHeader_String(t *testing.T) {
	h := Header{Version: 3, Count: 10}
	s := h.String()
	if !strings.Contains(s, "version=3") || !strings.Contains(s, "count=10") {
		t.Errorf("unexpected Header.String: %s", s)
	}
}

func TestRecord_String(t *testing.T) {
	rec := Record{Type: 2, ID: 42, Payload: []byte("xyz")}
	s := rec.String()
	if !strings.Contains(s, "type=2") || !strings.Contains(s, "id=42") || !strings.Contains(s, "3 bytes") {
		t.Errorf("unexpected Record.String: %s", s)
	}
}

func TestContainer_String(t *testing.T) {
	c := &Container{Header: Header{Version: 1, Count: 2}, Records: make([]Record, 2)}
	s := c.String()
	if !strings.Contains(s, "records=2") {
		t.Errorf("unexpected Container.String: %s", s)
	}
}

func TestClone(t *testing.T) {
	c := &Container{
		Header:  Header{Version: 1, Count: 1},
		Records: []Record{{Type: 1, ID: 1, Payload: []byte("hello")}},
	}
	cp := Clone(c)
	cp.Records[0].Payload[0] = 'X'
	if c.Records[0].Payload[0] == 'X' {
		t.Error("Clone should deep copy payload")
	}
}

func TestClone_Nil(t *testing.T) {
	if Clone(nil) != nil {
		t.Error("Clone(nil) should be nil")
	}
}

func TestEmpty(t *testing.T) {
	c := Empty(5)
	if c.Header.Version != 5 || c.Header.Count != 0 || len(c.Records) != 0 {
		t.Errorf("unexpected Empty: %+v", c)
	}
}
