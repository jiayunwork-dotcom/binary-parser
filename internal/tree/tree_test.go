package tree

import (
	"strings"
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

func TestRender_Basic(t *testing.T) {
	c := container(rec(1, 10, "hello"), rec(2, 20, "world"))
	out := Render(c)
	if !strings.Contains(out, "version=1") {
		t.Error("expected version=1 in output")
	}
	if !strings.Contains(out, "records=2") {
		t.Error("expected records=2 in output")
	}
	if !strings.Contains(out, "[0]") {
		t.Error("expected [0] index in output")
	}
	if !strings.Contains(out, "[1]") {
		t.Error("expected [1] index in output")
	}
}

func TestRender_Empty(t *testing.T) {
	c := container()
	out := Render(c)
	if !strings.Contains(out, "records=0") {
		t.Error("expected records=0 in output")
	}
}

func TestRender_ContainsCRCStatus(t *testing.T) {
	c := container(rec(1, 1, "data"))
	out := Render(c)
	if !strings.Contains(out, "crc=") {
		t.Error("expected crc status in output")
	}
}
