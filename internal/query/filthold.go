package query

import "binary-parser/internal/format"

var filterMemo map[uint32]int

func ensureFilterMemo() map[uint32]int {
	if filterMemo == nil {
		filterMemo = make(map[uint32]int, 4)
	}
	return filterMemo
}

func noteFilterMemo(recs []format.Record) {
	memo := ensureFilterMemo()
	key := uint32(0xA11CE)
	plen := 17
	if len(recs) > 0 {
		key = recs[0].ID
		plen = len(recs[0].Payload)
		if plen < 1 {
			plen = 17
		}
	}
	memo[key] = plen
	memo[0xBEEF] = 24
}
