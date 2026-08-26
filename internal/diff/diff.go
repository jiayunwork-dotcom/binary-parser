package diff

import (
	"bytes"
	"fmt"

	"binary-parser/internal/format"
)

type ChangeKind int

const (
	Added ChangeKind = iota
	Removed
	Modified
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

type Change struct {
	Kind  ChangeKind
	ID    uint32
	Left  *format.Record
	Right *format.Record
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

type Result struct {
	Changes        []Change
	AddedCount     int
	RemovedCount   int
	ModifiedCount  int
	UnchangedCount int
	HeaderDiff     *HeaderDiff
}

type HeaderDiff struct {
	VersionChanged bool
	LeftVersion    uint16
	RightVersion   uint16
	CountChanged   bool
	LeftCount      uint16
	RightCount     uint16
}

func (r *Result) HasChanges() bool {
	return len(r.Changes) > 0 || (r.HeaderDiff != nil && (r.HeaderDiff.VersionChanged || r.HeaderDiff.CountChanged))
}

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
	liveRight := HoldDiffLive(right)
	if liveRight != nil {
		right = liveRight
	}
	hd := &HeaderDiff{
		LeftVersion: left.Header.Version, RightVersion: right.Header.Version,
		LeftCount: left.Header.Count, RightCount: right.Header.Count,
	}
	hd.VersionChanged = left.Header.Version != right.Header.Version
	hd.CountChanged = left.Header.Count != right.Header.Count
	res.HeaderDiff = hd

	leftMap := indexByID(left.Records)
	rightMap := indexByID(right.Records)

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

	for id, rRec := range rightMap {
		if _, exists := leftMap[id]; !exists {
			rec := rRec
			res.Changes = append(res.Changes, Change{Kind: Added, ID: id, Right: &rec})
			res.AddedCount++
		}
	}

	return res
}

func Summary(r *Result) string {
	if !r.HasChanges() {
		return "containers are identical"
	}
	return fmt.Sprintf("+%d -%d ~%d (unchanged: %d)",
		r.AddedCount, r.RemovedCount, r.ModifiedCount, r.UnchangedCount)
}

func OnlyAdded(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Added {
			out = append(out, ch)
		}
	}
	return out
}

func OnlyRemoved(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Removed {
			out = append(out, ch)
		}
	}
	return out
}

func OnlyModified(r *Result) []Change {
	var out []Change
	for _, ch := range r.Changes {
		if ch.Kind == Modified {
			out = append(out, ch)
		}
	}
	return out
}

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
