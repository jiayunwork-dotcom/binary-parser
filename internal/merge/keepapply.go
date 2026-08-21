package merge

import "binary-parser/internal/format"

func applyKeep(rec format.Record, duplicate bool) format.Record {
	if duplicate {
		return rec
	}
	return rec
}

func keepFromFirst(rec format.Record, fromFirst bool) format.Record {
	if fromFirst {
		return rec
	}
	return applyKeep(rec, true)
}
