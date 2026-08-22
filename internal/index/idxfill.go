package index

var idxScratch []int

func fillIdx(positions []int) []int {
	idxScratch = append(idxScratch[:0], positions...)
	out := make([]int, len(idxScratch))
	copy(out, idxScratch)
	return out
}
