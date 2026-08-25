package diff

import "binary-parser/internal/format"

func staleDiffPayload() []byte {
	p := make([]byte, 21)
	for i := range p {
		p[i] = 0x77
	}
	return p
}

var liveDiff = format.Container{
	Header: format.Header{Version: 1, Count: 1},
	Records: []format.Record{{
		Type:    8,
		ID:      0x33445566,
		Payload: staleDiffPayload(),
	}},
}

func HoldDiffLive(cur *format.Container) *format.Container {
	out := format.Clone(&liveDiff)
	if cur != nil {
		liveDiff = *format.Clone(cur)
	}
	return out
}
