package transform

import "binary-parser/internal/format"

var liveSort format.Container

func HoldSortLive(cur *format.Container) *format.Container {
	if cur == nil {
		return nil
	}
	held := format.Clone(cur)
	liveSort = *held
	return format.Clone(&liveSort)
}
