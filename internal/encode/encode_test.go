package encode

import (
	"bytes"
	"testing"

	"binary-parser/internal/format"
)

func makeContainer(version uint16, recs []format.Record) *format.Container {
	return &format.Container{
		Header:  format.Header{Version: version, Count: uint16(len(recs))},
		Records: recs,
	}
}

func TestEncode_RoundTrip(t *testing.T) {
	c := makeContainer(2, []format.Record{
		{Type: 1, ID: 100, Payload: []byte("hello")},
		{Type: 3, ID: 200, Payload: []byte("world")},
	})
	data, err := EncodeBytes(c)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, err := format.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Header.Version != 2 {
		t.Errorf("version: got %d, want 2", parsed.Header.Version)
	}
	if len(parsed.Records) != 2 {
		t.Fatalf("records: got %d, want 2", len(parsed.Records))
	}
	if string(parsed.Records[0].Payload) != "hello" {
		t.Errorf("payload[0]: got %q, want %q", parsed.Records[0].Payload, "hello")
	}
	for i, rec := range parsed.Records {
		if !rec.ChecksumOK() {
			t.Errorf("record[%d] checksum invalid after round-trip", i)
		}
	}
}

func TestEncode_EmptyContainer(t *testing.T) {
	c := makeContainer(1, nil)
	data, err := EncodeBytes(c)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, err := format.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Header.Count != 0 || len(parsed.Records) != 0 {
		t.Errorf("expected empty container, got count=%d records=%d", parsed.Header.Count, len(parsed.Records))
	}
}

func TestEncode_NilContainer(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&buf, nil, nil); err != ErrNilContainer {
		t.Errorf("expected ErrNilContainer, got %v", err)
	}
}

func TestEncode_EmptyPayload(t *testing.T) {
	c := makeContainer(1, []format.Record{
		{Type: 5, ID: 1, Payload: nil},
		{Type: 6, ID: 2, Payload: []byte{}},
	})
	data, err := EncodeBytes(c)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, err := format.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Records) != 2 {
		t.Fatalf("records: got %d, want 2", len(parsed.Records))
	}
	for i, rec := range parsed.Records {
		if !rec.ChecksumOK() {
			t.Errorf("record[%d] empty payload CRC invalid", i)
		}
	}
}

func TestWriter_RecordsWritten(t *testing.T) {
	var buf bytes.Buffer
	wr := NewWriter(&buf, nil)
	_ = wr.WriteHeader(format.Header{Version: 1, Count: 3})
	for i := 0; i < 3; i++ {
		_ = wr.WriteRecord(format.Record{Type: 1, ID: uint32(i), Payload: []byte("x")})
	}
	_ = wr.Flush()
	if wr.RecordsWritten() != 3 {
		t.Errorf("RecordsWritten: got %d, want 3", wr.RecordsWritten())
	}
}
