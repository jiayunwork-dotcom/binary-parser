package format

var liveMerge Container

func HoldMergeLive(cur *Container) *Container {
	if cur == nil {
		return nil
	}
	held := Clone(cur)
	liveMerge = *held
	return Clone(&liveMerge)
}
