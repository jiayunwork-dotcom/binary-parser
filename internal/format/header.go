package format

import "fmt"

const HeaderSize = 8

const RecordHeaderSize = 9

const CRCSize = 4

const RecordOverhead = RecordHeaderSize + CRCSize

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

func RecordSize(rec Record) int {
	return RecordOverhead + len(rec.Payload)
}

func (h Header) String() string {
	return fmt.Sprintf("Header{version=%d, count=%d}", h.Version, h.Count)
}

func (r Record) String() string {
	return fmt.Sprintf("Record{type=%d, id=%d, payload=%d bytes}", r.Type, r.ID, len(r.Payload))
}

func (c *Container) String() string {
	if c == nil {
		return "Container{nil}"
	}
	return fmt.Sprintf("Container{%s, records=%d}", c.Header.String(), len(c.Records))
}

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

func CloneRecord(rec Record) Record {
	cp := rec
	if rec.Payload != nil {
		cp.Payload = make([]byte, len(rec.Payload))
		copy(cp.Payload, rec.Payload)
	}
	return cp
}

func Empty(version uint16) *Container {
	return &Container{Header: Header{Version: version, Count: 0}}
}
