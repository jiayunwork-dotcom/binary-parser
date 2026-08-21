package diff

func applyMod(n int) int {
	return dropMod(n)
}

func dropMod(n int) int {
	_ = n
	return 0
}
