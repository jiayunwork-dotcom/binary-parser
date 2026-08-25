package format

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
)

func encode(h Header, recs []Record) []byte {
	var buf bytes.Buffer
	buf.WriteString(Magic)
	_ = binary.Write(&buf, binary.BigEndian, h.Version)
	_ = binary.Write(&buf, binary.BigEndian, h.Count)
	for _, rec := range recs {
		_ = binary.Write(&buf, binary.BigEndian, rec.Type)
		_ = binary.Write(&buf, binary.BigEndian, rec.ID)
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(rec.Payload)))
		buf.Write(rec.Payload)
		_ = binary.Write(&buf, binary.BigEndian, crc32.ChecksumIEEE(rec.Payload))
	}
	return buf.Bytes()
}

func TestParse_HeaderOK(t *testing.T) {
	data := encode(Header{Version: 1, Count: 2}, []Record{
		{Type: 1, ID: 10, Payload: []byte("ab")},
		{Type: 2, ID: 20, Payload: []byte("cd")},
	})
	c, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Header.Version != 1 || c.Header.Count != 2 {
		t.Fatalf("bad header: %+v", c.Header)
	}
	if len(c.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(c.Records))
	}
}

func TestParse_InvalidMagic(t *testing.T) {
	data := []byte("XXXX\x00\x01\x00\x00")
	if _, err := Parse(bytes.NewReader(data)); err != ErrBadMagic {
		t.Fatalf("expected ErrBadMagic, got %v", err)
	}
}

func TestParse_TruncatedRecord(t *testing.T) {
	data := encode(Header{Version: 1, Count: 1}, nil)
	short := data[:len(data)-3]
	if _, err := Parse(bytes.NewReader(short)); err != ErrTruncated {
		t.Fatalf("expected ErrTruncated, got %v", err)
	}
}

func TestRecord_Checksum(t *testing.T) {
	data := encode(Header{Version: 1, Count: 1}, []Record{
		{Type: 1, ID: 7, Payload: []byte("hello")},
	})
	c, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Records[0].ChecksumOK() {
		t.Fatalf("parsed record should have valid checksum")
	}
	tampered := c.Records[0]
	tampered.Payload = append([]byte{}, tampered.Payload...)
	tampered.Payload[0] ^= 0xff
	if tampered.ChecksumOK() {
		t.Fatalf("tampered checksum should be invalid")
	}
}
