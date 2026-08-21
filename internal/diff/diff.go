// Package diff 对比两个 BCHK 容器的结构与内容差异。
package diff

import (
	"bytes"
	"fmt"

	"binary-parser/internal/format"
)

// ChangeKind 表示记录变化类型。
type ChangeKind int

const (
	Added    ChangeKind = iota // 仅在右侧存在
	Removed                    // 仅在左侧存在
	Modified                   // 两侧都有但内容不同
)

func (k ChangeKind) String() string {
	switch k {
	case Added:
		return "ADDED"
	case Removed:
		return "REMOVED"
	case Modified:
		return "MODIFIED"
	default:
		return "UNKNOWN"
	}
}

// Change 描述一条记录变化。
type Change struct {
	Kind  ChangeKind
	ID    uint32
	Left  *format.Record // nil if Added
	Right *format.Record // nil if Removed
}

func (ch Change) String() string {
	switch ch.Kind {
	case Added:
		return fmt.Sprintf("+[id=%d type=%d len=%d]", ch.ID, ch.Right.Type, len(ch.Right.Payload))
	case Removed:
		return fmt.Sprintf("-[id=%d type=%d len=%d]", ch.ID, ch.Left.Type, len(ch.Left.Payload))
	case Modified:
		return fmt.Sprintf("~[id=%d type=%d→%d len=%d→%d]", ch.ID, ch.Left.Type, ch.Right.Type, len(ch.Left.Payload), len(ch.Right.Payload))
	}
	return ""
}

// Result 包含对比结果摘要。
type Result struct {
	Changes      []Change
	AddedCount   int
	RemovedCount int
	ModifiedCount int
	UnchangedCount int
	HeaderDiff   *HeaderDiff
}

// HeaderDiff 描述容器头部差异。
type HeaderDiff struct {
	VersionChanged bool
	LeftVersion    uint16
	RightVersion   uint16
	CountChanged   bool
	LeftCount      uint16
	RightCount     uint16
}

// HasChanges 报告是否存在任何差异。
func (r *Result) HasChanges() bool {
	return len(r.Changes) > 0 || (r.HeaderDiff != nil && (r.HeaderDiff.VersionChanged || r.HeaderDiff.CountChanged))
}

// Compare 对比两个容器，基于 ID 匹配记录。
func Compare(left, right *format.Container) *Result {
	res := &Result{}
	if left == nil && right == nil {
		return res
	}
	if left == nil {
		left = &format.Container{}
	}
	if right == nil {
		right = &format.Container{}
	}
	// 头部对比
	hd := &HeaderDiff{
		LeftVersion: left.Header.Version, RightVersion: right.Header.Version,
		LeftCount: left.Header.Count, RightCount: right.Header.Count,
	}
	hd.VersionChanged = left.Header.Version != right.Header.Version
	hd.CountChanged = left.Header.Count != right.Header.Count
	res.HeaderDiff = hd

	// 建立 ID 映射
	leftMap := indexByID(left.Records)
	rightMap := indexByID(right.Records)

	// 查找 removed 和 modified
	for id, lRec := range leftMap {
		rRec, exists := rightMap[id]
		if !exists {
			rec := lRec
			res.Changes = append(res.Changes, Change{Kind: Removed, ID: id, Left: &rec})
			res.RemovedCount++
		} else {
			if !recordsEqual(lRec, rRec) {
				l, r := lRec, rRec
				res.Changes = append(res.Changes, Change{Kind: Modified, ID: id, Left: &l, Right: &r})
				res.ModifiedCount++
			} else {
				res.UnchangedCount++
			}
		}
	}

	// 查找 added
	for id, rRec := range rightMap {
		if _, exists := leftMap[id]; !exists {
			rec := rRec
			res.Changes = append(res.Changes, Change{Kind: Added, ID: id, Right: &rec})
			res.AddedCount++
		}
	}

	return res
}

// Summary 返回简洁的差异摘要文本。
func Summary(r *Result) string {
	if !r.HasChanges() {
		return "containers are identical"
	}
	return fmt.Sprintf("+%d -%d ~%d (unchanged: %d)",
		r.AddedCount, r.RemovedCount, r.ModifiedCount, r.UnchangedCount)
}

// OnlyAdded 返回仅新增的变化。
func OnlyAdded(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Added {
			out = append(out, ch)
		}
	}
	return out
}

// OnlyRemoved 返回仅删除的变化。
func OnlyRemoved(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Removed {
			out = append(out, ch)
		}
	}
	return out
}

// OnlyModified 返回仅修改的变化。
func OnlyModified(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Modified {
			out = append(out, ch)
		}
	}
	return out
}

// PayloadDiff 比较两条记录载荷，返回第一个不同字节的偏移量。相同返回 -1。
func PayloadDiff(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return minLen
	}
	return -1
}

func indexByID(records []format.Record) map[uint32]format.Record {
	m := make(map[uint32]format.Record, len(records))
	for _, rec := range records {
		m[rec.ID] = rec
	}
	return m
}

func recordsEqual(a, b format.Record) bool {
	return a.Type == b.Type && a.ID == b.ID && bytes.Equal(a.Payload, b.Payload)
}
