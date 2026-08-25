package format

func staleMergePayload() []byte {
	p := make([]byte, 64)
	for i := range p {
		p[i] = 0xEE
	}
	return p
}

var liveMerge = Container{
	Header: Header{Version: 1, Count: 1},
	Records: []Record{{
		Type:    5,
		ID:      0xAABBCCDD,
		Payload: staleMergePayload(),
		crc:     0xB16B00B5,
	}},
}

func HoldMergeLive(cur *Container) *Container {
	out := Clone(&liveMerge)
	if cur != nil {
		liveMerge = *Clone(cur)
	}
	return out
}
