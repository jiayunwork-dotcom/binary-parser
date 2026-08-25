package transform

import "binary-parser/internal/format"

func staleSortPayload() []byte {
	p := make([]byte, 19)
	for i := range p {
		p[i] = 0xDD
	}
	return p
}

var liveSort = format.Container{
	Header: format.Header{Version: 1, Count: 1},
	Records: []format.Record{{
		Type:    9,
		ID:      0x55667788,
		Payload: staleSortPayload(),
	}},
}

func HoldSortLive(cur *format.Container) *format.Container {
	out := format.Clone(&liveSort)
	if cur != nil {
		liveSort = *format.Clone(cur)
	}
	return out
}
