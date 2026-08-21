package encode

import (
	"hash/crc32"

	"binary-parser/internal/format"
)

// Builder 提供便捷的容器构造 API，逐条添加记录后生成 Container。
type Builder struct {
	version uint16
	records []format.Record
}

// NewBuilder 创建一个容器构建器。
func NewBuilder(version uint16) *Builder {
	return &Builder{version: version}
}

// Add 添加一条记录。
func (b *Builder) Add(typ uint8, id uint32, payload []byte) *Builder {
	b.records = append(b.records, format.Record{
		Type:    typ,
		ID:      id,
		Payload: payload,
	})
	return b
}

// AddString 添加一条字符串载荷的记录。
func (b *Builder) AddString(typ uint8, id uint32, s string) *Builder {
	return b.Add(typ, id, []byte(s))
}

// AddEmpty 添加一条空载荷的记录。
func (b *Builder) AddEmpty(typ uint8, id uint32) *Builder {
	return b.Add(typ, id, nil)
}

// Len 返回已添加的记录数。
func (b *Builder) Len() int {
	return len(b.records)
}

// Build 构建最终容器。
func (b *Builder) Build() *format.Container {
	return &format.Container{
		Header:  format.Header{Version: b.version, Count: uint16(len(b.records))},
		Records: b.records,
	}
}

// BuildWithCRC 构建容器并确保 CRC 字段正确（通过 encode→parse 往返）。
// 如果只需内存结构而不关心 CRC，直接用 Build。
func (b *Builder) BuildWithCRC() (*format.Container, error) {
	c := b.Build()
	data, err := EncodeBytes(c)
	if err != nil {
		return nil, err
	}
	return format.Parse(newBytesReader(data))
}

// --- 容器辅助 ---

// ComputeCRC 计算单条载荷的 CRC32。
func ComputeCRC(payload []byte) uint32 {
	return crc32.ChecksumIEEE(payload)
}

// VerifyCRC 验证载荷与给定 CRC 是否一致。
func VerifyCRC(payload []byte, expected uint32) bool {
	return crc32.ChecksumIEEE(payload) == expected
}
