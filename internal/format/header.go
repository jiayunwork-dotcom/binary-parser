package format

import "fmt"

// HeaderSize 是固定文件头占用的字节数（magic 4 + version 2 + count 2 = 8）。
const HeaderSize = 8

// RecordHeaderSize 是单条记录头的固定字节数（type 1 + id 4 + length 4 = 9）。
const RecordHeaderSize = 9

// CRCSize 是 CRC32 字段占用的字节数。
const CRCSize = 4

// RecordOverhead 是单条记录除载荷外的固定开销（头 + CRC）。
const RecordOverhead = RecordHeaderSize + CRCSize

// ContainerSize 估算容器序列化后的总字节数。
func ContainerSize(c *Container) int {
	if c == nil {
		return 0
	}
	size := HeaderSize
	for _, rec := range c.Records {
		size += RecordOverhead + len(rec.Payload)
	}
	return size
}

// RecordSize 返回单条记录序列化后占用的字节数。
func RecordSize(rec Record) int {
	return RecordOverhead + len(rec.Payload)
}

// String 返回 Header 的调试文本表示。
func (h Header) String() string {
	return fmt.Sprintf("Header{version=%d, count=%d}", h.Version, h.Count)
}

// String 返回 Record 的调试文本表示。
func (r Record) String() string {
	return fmt.Sprintf("Record{type=%d, id=%d, payload=%d bytes}", r.Type, r.ID, len(r.Payload))
}

// String 返回 Container 的调试文本表示。
func (c *Container) String() string {
	if c == nil {
		return "Container{nil}"
	}
	return fmt.Sprintf("Container{%s, records=%d}", c.Header.String(), len(c.Records))
}

// Clone 深拷贝容器。
func Clone(c *Container) *Container {
	if c == nil {
		return nil
	}
	out := &Container{
		Header:  c.Header,
		Records: make([]Record, len(c.Records)),
	}
	for i, rec := range c.Records {
		out.Records[i] = CloneRecord(rec)
	}
	return out
}

// CloneRecord 深拷贝单条记录。
func CloneRecord(rec Record) Record {
	cp := rec
	if rec.Payload != nil {
		cp.Payload = make([]byte, len(rec.Payload))
		copy(cp.Payload, rec.Payload)
	}
	return cp
}

// Empty 返回一个空容器。
func Empty(version uint16) *Container {
	return &Container{Header: Header{Version: version, Count: 0}}
}
