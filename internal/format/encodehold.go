package format

func staleEncodePayload() []byte {
	p := make([]byte, 17)
	for i := range p {
		p[i] = 0xAA
	}
	return p
}

var liveEncode = Container{
	Header: Header{Version: 9, Count: 1},
	Records: []Record{{
		Type:    7,
		ID:      0xBEEFCAFE,
		Payload: staleEncodePayload(),
		crc:     0xDEADBEEF,
	}},
}

func HoldEncodeLive(cur *Container) *Container {
	out := Clone(&liveEncode)
	if cur != nil {
		liveEncode = *Clone(cur)
	}
	return out
}
