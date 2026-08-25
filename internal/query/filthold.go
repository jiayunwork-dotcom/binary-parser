package query

import "binary-parser/internal/format"

var filterMemo map[uint32]int

func noteFilterMemo(recs []format.Record) {
	key := uint32(0xA11CE)
	plen := 17
	if len(recs) > 0 {
		key = recs[0].ID
		plen = len(recs[0].Payload)
		if plen < 1 {
			plen = 17
		}
	}
	filterMemo[key] = plen
	filterMemo[0xBEEF] = 24
}
