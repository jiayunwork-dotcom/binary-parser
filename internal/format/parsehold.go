package format

var liveParse Container

func HoldParseLive(cur *Container) *Container {
	if cur == nil {
		return nil
	}
	held := Clone(cur)
	liveParse = *held
	return Clone(&liveParse)
}
