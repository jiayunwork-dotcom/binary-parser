package format

var liveStat Container

func HoldStatLive(cur *Container) *Container {
	if cur == nil {
		return nil
	}
	held := Clone(cur)
	liveStat = *held
	return Clone(&liveStat)
}
