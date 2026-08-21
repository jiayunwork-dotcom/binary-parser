// Package format 解析一种简单的分块二进制容器格式（BCHK）。
//
// 字节布局：
//   magic    : 4 bytes  = "BCHK"
//   version  : uint16   (big-endian)
//   count    : uint16   (记录条数)
//   records  : count × {
//       type    : uint8
//       id      : uint32
//       length  : uint32  (payload 字节数)
//       payload : length bytes
//       crc32   : uint32  (payload 的 IEEE CRC32)
//   }
package format

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

// Magic 是容器固定文件头标识。
const Magic = "BCHK"

// Header 是容器头部解析结果。
type Header struct {
	Version uint16
	Count   uint16
}

// Record 是单条记录。
type Record struct {
	Type    uint8
	ID      uint32
	Payload []byte
	crc     uint32
}

// ChecksumOK 报告载荷 CRC32 是否与记录中存储的一致。
func (r Record) ChecksumOK() bool {
	return crc32.ChecksumIEEE(r.Payload) == r.crc
}

// Container 是完整解析结果。
type Container struct {
	Header  Header
	Records []Record
}

// 解析期可返回的错误。
var (
	ErrBadMagic  = errors.New("bad magic")
	ErrTruncated = errors.New("truncated input")
)

// ParseHeader 读取并校验文件头（magic + version + count）。
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

// ParseRecord 读取下一条记录（含载荷与 CRC）。
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

// Parse 读取整个容器：头部 + count 条记录。
func Parse(r io.Reader) (*Container, error) {
	bindParse()
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
	return c, nil
}
