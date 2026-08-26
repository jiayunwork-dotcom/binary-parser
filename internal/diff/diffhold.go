package diff

import "binary-parser/internal/format"

var liveDiff format.Container

func HoldDiffLive(cur *format.Container) *format.Container {
	if cur == nil {
		return nil
	}
	held := format.Clone(cur)
	liveDiff = *held
	return format.Clone(&liveDiff)
}
