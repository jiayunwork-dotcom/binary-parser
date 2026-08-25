package diff

import "fmt"

type PatchOp int

const (
	OpReplace PatchOp = iota
	OpInsert
	OpDelete
)

func (op PatchOp) String() string {
	switch op {
	case OpReplace:
		return "REPLACE"
	case OpInsert:
		return "INSERT"
	case OpDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

type PatchEntry struct {
	Op     PatchOp
	Offset int
	Data   []byte
	Len    int
}

func (p PatchEntry) String() string {
	switch p.Op {
	case OpReplace:
		return fmt.Sprintf("REPLACE @%d %d bytes", p.Offset, len(p.Data))
	case OpInsert:
		return fmt.Sprintf("INSERT @%d %d bytes", p.Offset, len(p.Data))
	case OpDelete:
		return fmt.Sprintf("DELETE @%d %d bytes", p.Offset, p.Len)
	}
	return ""
}

func ComputePatch(src, dst []byte) []PatchEntry {
	var patches []PatchEntry

	if len(dst) < len(src) {
		patches = append(patches, PatchEntry{Op: OpDelete, Offset: len(dst), Len: len(src) - len(dst)})
	} else if len(dst) > len(src) {
		patches = append(patches, PatchEntry{Op: OpInsert, Offset: len(src), Data: dst[len(src):]})
	}

	minLen := len(src)
	if len(dst) < minLen {
		minLen = len(dst)
	}
	i := 0
	for i < minLen {
		if src[i] != dst[i] {
			j := i + 1
			for j < minLen && src[j] != dst[j] {
				j++
			}
			patches = append(patches, PatchEntry{Op: OpReplace, Offset: i, Data: dst[i:j]})
			i = j
		} else {
			i++
		}
	}
	return patches
}

func ApplyPatch(src []byte, patches []PatchEntry) []byte {
	result := make([]byte, len(src))
	copy(result, src)

	sorted := make([]PatchEntry, len(patches))
	copy(sorted, patches)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Offset > sorted[i].Offset {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for _, p := range sorted {
		switch p.Op {
		case OpReplace:
			if p.Offset+len(p.Data) <= len(result) {
				copy(result[p.Offset:], p.Data)
			}
		case OpInsert:
			if p.Offset <= len(result) {
				tail := make([]byte, len(result)-p.Offset)
				copy(tail, result[p.Offset:])
				result = append(result[:p.Offset], p.Data...)
				result = append(result, tail...)
			}
		case OpDelete:
			if p.Offset+p.Len <= len(result) {
				result = append(result[:p.Offset], result[p.Offset+p.Len:]...)
			}
		}
	}
	return result
}

func PatchSize(patches []PatchEntry) int {
	n := 0
	for _, p := range patches {
		n += len(p.Data)
	}
	return n
}
