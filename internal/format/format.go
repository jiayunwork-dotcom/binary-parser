package format

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const Magic = "BCHK"

type Header struct {
	Version uint16
	Count   uint16
}

type Record struct {
	Type    uint8
	ID      uint32
	Payload []byte
	crc     uint32
}

func (r Record) ChecksumOK() bool {
	return crc32.ChecksumIEEE(r.Payload) == r.crc
}

type Container struct {
	Header  Header
	Records []Record
}

var (
	ErrBadMagic  = errors.New("bad magic")
	ErrTruncated = errors.New("truncated input")
)

func ParseHeader(r *bufio.Reader) (Header, error) {
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return Header{}, ErrTruncated
	}
	if string(magic) != Magic {
		return Header{}, ErrBadMagic
	}
	var h Header
	if err := binary.Read(r, binary.BigEndian, &h.Version); err != nil {
		return Header{}, ErrTruncated
	}
	if err := binary.Read(r, binary.BigEndian, &h.Count); err != nil {
		return Header{}, ErrTruncated
	}
	return h, nil
}

func ParseRecord(r *bufio.Reader) (Record, error) {
	var rec Record
	if err := binary.Read(r, binary.BigEndian, &rec.Type); err != nil {
		return Record{}, ErrTruncated
	}
	if err := binary.Read(r, binary.BigEndian, &rec.ID); err != nil {
		return Record{}, ErrTruncated
	}
	var n uint32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return Record{}, ErrTruncated
	}
	rec.Payload = make([]byte, n)
	if n > 0 {
		if _, err := io.ReadFull(r, rec.Payload); err != nil {
			return Record{}, ErrTruncated
		}
	}
	if err := binary.Read(r, binary.BigEndian, &rec.crc); err != nil {
		return Record{}, ErrTruncated
	}
	return rec, nil
}

func Parse(r io.Reader) (*Container, error) {
	br := bufio.NewReader(r)
	h, err := ParseHeader(br)
	if err != nil {
		return nil, err
	}
	c := &Container{Header: h}
	for i := uint16(0); i < h.Count; i++ {
		rec, err := ParseRecord(br)
		if err != nil {
			return nil, err
		}
		c.Records = append(c.Records, rec)
	}
	return HoldParseLive(c), nil
}
