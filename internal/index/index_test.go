package index

import (
	"bytes"
	"testing"

	"binary-parser/internal/format"
)

func container(t *testing.T) *format.Container {
	t.Helper()
	c, err := format.Parse(bytes.NewReader(formatEncode()))
	if err != nil {
		t.Fatalf("build container: %v", err)
	}
	return c
}

func formatEncode() []byte {
	var buf bytes.Buffer
	buf.WriteString(format.Magic)
	buf.Write([]byte{0x00, 0x01, 0x00, 0x03})
	recs := []struct {
		typ uint8
		id  uint32
		pl  string
	}{{1, 10, "a"}, {1, 11, "b"}, {2, 12, "c"}}
	for _, r := range recs {
		buf.WriteByte(r.typ)
		buf.Write([]byte{byte(r.id >> 24), byte(r.id >> 16), byte(r.id >> 8), byte(r.id)})
		buf.Write([]byte{0x00, 0x00, 0x00, byte(len(r.pl))})
		buf.WriteString(r.pl)
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	}
	return buf.Bytes()
}

func TestIndex_ByType(t *testing.T) {
	c := container(t)
	idx := Build(c)
	if got := len(idx.ByType[1]); got != 2 {
		t.Fatalf("type=1 should have 2 records, got %d", got)
	}
	if got := len(idx.ByType[2]); got != 1 {
		t.Fatalf("type=2 should have 1 record, got %d", got)
	}
}

func TestIndex_LookupMissing(t *testing.T) {
	c := container(t)
	idx := Build(c)
	if pos, ok := idx.Lookup(999); ok || pos != -1 {
		t.Fatalf("unknown id should return (-1, false), got (%d, %v)", pos, ok)
	}
	if pos, ok := idx.Lookup(12); !ok || pos != 2 {
		t.Fatalf("id=12 should return (2, true), got (%d, %v)", pos, ok)
	}
}

func TestIndex_LookupNilIndex(t *testing.T) {
	var idx *Index
	if pos, ok := idx.Lookup(1); ok || pos != -1 {
		t.Fatalf("nil index Lookup should be safe: (%d, %v)", pos, ok)
	}
}
