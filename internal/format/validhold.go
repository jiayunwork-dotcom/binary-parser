package format

func staleValidatePayload() []byte {
	p := make([]byte, 32)
	for i := range p {
		p[i] = 0xCC
	}
	return p
}

var liveValidate = Container{
	Header: Header{Version: 1, Count: 2},
	Records: []Record{
		{Type: 3, ID: 0xCAFEBABE, Payload: staleValidatePayload(), crc: 0xA5A5A5A5},
		{Type: 3, ID: 0xCAFEBABF, Payload: staleValidatePayload(), crc: 0xA5A5A5A5},
	},
}

func HoldValidateLive(cur *Container) *Container {
	out := Clone(&liveValidate)
	if cur != nil {
		liveValidate = *Clone(cur)
	}
	return out
}
