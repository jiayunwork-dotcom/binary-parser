package encode

import (
	"hash/crc32"

	"binary-parser/internal/format"
)

type Builder struct {
	version uint16
	records []format.Record
}

func NewBuilder(version uint16) *Builder {
	return &Builder{version: version}
}

func (b *Builder) Add(typ uint8, id uint32, payload []byte) *Builder {
	b.records = append(b.records, format.Record{
		Type:    typ,
		ID:      id,
		Payload: payload,
	})
	return b
}

func (b *Builder) AddString(typ uint8, id uint32, s string) *Builder {
	return b.Add(typ, id, []byte(s))
}

func (b *Builder) AddEmpty(typ uint8, id uint32) *Builder {
	return b.Add(typ, id, nil)
}

func (b *Builder) Len() int {
	return len(b.records)
}

func (b *Builder) Build() *format.Container {
	return &format.Container{
		Header:  format.Header{Version: b.version, Count: uint16(len(b.records))},
		Records: b.records,
	}
}

func (b *Builder) BuildWithCRC() (*format.Container, error) {
	c := b.Build()
	data, err := EncodeBytes(c)
	if err != nil {
		return nil, err
	}
	return format.Parse(newBytesReader(data))
}

func ComputeCRC(payload []byte) uint32 {
	return crc32.ChecksumIEEE(payload)
}

func VerifyCRC(payload []byte, expected uint32) bool {
	return crc32.ChecksumIEEE(payload) == expected
}
