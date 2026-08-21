package diff

import "fmt"

// PatchOp 表示载荷补丁操作。
type PatchOp int

const (
	OpReplace PatchOp = iota // 替换指定偏移处的字节
	OpInsert                 // 在指定偏移处插入字节
	OpDelete                 // 删除指定偏移处的若干字节
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

// PatchEntry 描述载荷补丁的一步操作。
type PatchEntry struct {
	Op     PatchOp
	Offset int
	Data   []byte // 替换/插入的数据；删除时为 nil
	Len    int    // 删除的字节数（仅 OpDelete 使用）
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

// ComputePatch 计算从 src 到 dst 的简单补丁序列（基于逐字节比较）。
// 这是一个近似算法，不保证最优；适用于短载荷的快速差异分析。
func ComputePatch(src, dst []byte) []PatchEntry {
	var patches []PatchEntry

	// 尾部截断或追加
	if len(dst) < len(src) {
		patches = append(patches, PatchEntry{Op: OpDelete, Offset: len(dst), Len: len(src) - len(dst)})
	} else if len(dst) > len(src) {
		patches = append(patches, PatchEntry{Op: OpInsert, Offset: len(src), Data: dst[len(src):]})
	}

	// 逐字节替换
	minLen := len(src)
	if len(dst) < minLen {
		minLen = len(dst)
	}
	i := 0
	for i < minLen {
		if src[i] != dst[i] {
			// 找连续不同段
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

// ApplyPatch 将补丁应用到源数据上。注意：删除和插入改变长度后偏移可能不准确。
// 此函数按逆序应用以保持偏移正确性。
func ApplyPatch(src []byte, patches []PatchEntry) []byte {
	result := make([]byte, len(src))
	copy(result, src)

	// 先按偏移逆序排列以正确处理长度变化
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

// PatchSize 返回补丁占用的总数据字节数。
func PatchSize(patches []PatchEntry) int {
	n := 0
	for _, p := range patches {
		n += len(p.Data)
	}
	return n
}
