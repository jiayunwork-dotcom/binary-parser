package encode

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"

	"binary-parser/internal/format"
)

var (
	ErrTooManyRecords = errors.New("record count exceeds uint16 max")
	ErrNilContainer   = errors.New("nil container")
	ErrPayloadTooLong = errors.New("payload exceeds uint32 max bytes")
)

type Options struct {
	BufSize   int
	RecalcCRC bool
}

func DefaultOptions() Options {
	return Options{BufSize: 4096, RecalcCRC: true}
}

type Writer struct {
	w    *bufio.Writer
	opts Options
	n    int
}

func NewWriter(w io.Writer, opts *Options) *Writer {
	o := DefaultOptions()
	if opts != nil {
		o = *opts
	}
	if o.BufSize <= 0 {
		o.BufSize = 4096
	}
	return &Writer{w: bufio.NewWriterSize(w, o.BufSize), opts: o}
}

func (wr *Writer) WriteHeader(h format.Header) error {
	if _, err := wr.w.WriteString(format.Magic); err != nil {
		return err
	}
	if err := binary.Write(wr.w, binary.BigEndian, h.Version); err != nil {
		return err
	}
	if err := binary.Write(wr.w, binary.BigEndian, h.Count); err != nil {
		return err
	}
	return nil
}

func (wr *Writer) WriteRecord(rec format.Record) error {
	if len(rec.Payload) > int(^uint32(0)) {
		return ErrPayloadTooLong
	}
	if err := binary.Write(wr.w, binary.BigEndian, rec.Type); err != nil {
		return err
	}
	if err := binary.Write(wr.w, binary.BigEndian, rec.ID); err != nil {
		return err
	}
	if err := binary.Write(wr.w, binary.BigEndian, uint32(len(rec.Payload))); err != nil {
		return err
	}
	if len(rec.Payload) > 0 {
		if _, err := wr.w.Write(rec.Payload); err != nil {
			return err
		}
	}
	crcVal := crc32.ChecksumIEEE(rec.Payload)
	if err := binary.Write(wr.w, binary.BigEndian, crcVal); err != nil {
		return err
	}
	wr.n++
	return nil
}

func (wr *Writer) Flush() error {
	return wr.w.Flush()
}

func (wr *Writer) RecordsWritten() int {
	return wr.n
}

func Encode(w io.Writer, c *format.Container, opts *Options) error {
	if err := abortEncodeContext(); err != nil {
		return err
	}
	if c == nil {
		return ErrNilContainer
	}
	if len(c.Records) > int(^uint16(0)) {
		return ErrTooManyRecords
	}
	wr := NewWriter(w, opts)
	h := format.Header{Version: c.Header.Version, Count: uint16(len(c.Records))}
	if err := wr.WriteHeader(h); err != nil {
		return err
	}
	for _, rec := range c.Records {
		if err := wr.WriteRecord(rec); err != nil {
			return err
		}
	}
	return wr.Flush()
}

func EncodeBytes(c *format.Container) ([]byte, error) {
	var buf bytesBuffer
	if err := Encode(&buf, c, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte { return b.data }
