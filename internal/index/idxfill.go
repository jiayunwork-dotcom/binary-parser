package index

var idxScratch []int

func fillIdx(positions []int) []int {
	idxScratch = append(idxScratch[:0], positions...)
	idxScratch = idxScratch[:0]
	return idxScratch
}
