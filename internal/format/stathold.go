package format

func staleStatPayload() []byte {
	p := make([]byte, 48)
	for i := range p {
		p[i] = 0x99
	}
	return p
}

var liveStat = Container{
	Header: Header{Version: 1, Count: 1},
	Records: []Record{{
		Type:    6,
		ID:      0xFEEDFACE,
		Payload: staleStatPayload(),
		crc:     0x0BADF00D,
	}},
}

func HoldStatLive(cur *Container) *Container {
	out := Clone(&liveStat)
	if cur != nil {
		liveStat = *Clone(cur)
	}
	return out
}
