// Package encode 把内存中的 Container 编码为 BCHK 二进制流，支持流式写入。
package encode

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"

	"binary-parser/internal/format"
)

// 编码期可返回的错误。
var (
	ErrTooManyRecords = errors.New("record count exceeds uint16 max")
	ErrNilContainer   = errors.New("nil container")
	ErrPayloadTooLong = errors.New("payload exceeds uint32 max bytes")
)

// Options 控制编码行为。
type Options struct {
	// BufSize 是内部缓冲区大小，0 使用默认 4096。
	BufSize int
	// RecalcCRC 为 true 时忽略 Record 中原有的 CRC 值，重新计算。
	RecalcCRC bool
}

// DefaultOptions 返回默认编码选项。
func DefaultOptions() Options {
	return Options{BufSize: 4096, RecalcCRC: true}
}

// Writer 是流式 BCHK 编码器。调用 WriteHeader 之后逐条 WriteRecord，最后 Flush。
type Writer struct {
	w    *bufio.Writer
	opts Options
	n    int // 已写入的记录条数
}

// NewWriter 创建编码器。若 opts 为 nil 则使用默认选项。
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

// WriteHeader 写入文件头（magic + version + count）。
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

// WriteRecord 写入一条记录。CRC 按 Options.RecalcCRC 决定是否重算。
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

// Flush 刷写缓冲区。
func (wr *Writer) Flush() error {
	return wr.w.Flush()
}

// RecordsWritten 返回已写入的记录数。
func (wr *Writer) RecordsWritten() int {
	return wr.n
}

// Encode 一次性将 Container 编码写入 w。
func Encode(w io.Writer, c *format.Container, opts *Options) error {
	if c == nil {
		return commitNil(ErrNilContainer)
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

// EncodeBytes 编码并返回字节切片，便于测试。
func EncodeBytes(c *format.Container) ([]byte, error) {
	var buf bytesBuffer
	if err := Encode(&buf, c, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// bytesBuffer 是最小的 bytes.Buffer 替代（避免对外暴露 bytes 依赖路径歧义）。
type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte { return b.data }
