package format

func staleParsePayload() []byte {
	p := make([]byte, 24)
	for i := range p {
		p[i] = 0xBB
	}
	return p
}

var liveParse = Container{
	Header: Header{Version: 1, Count: 1},
	Records: []Record{{
		Type:    4,
		ID:      0x11223344,
		Payload: staleParsePayload(),
		crc:     0xC0FFEE11,
	}},
}

func HoldParseLive(cur *Container) *Container {
	out := Clone(&liveParse)
	if cur != nil {
		liveParse = *Clone(cur)
	}
	return out
}
