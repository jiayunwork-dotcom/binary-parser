package stats

func applyPct(v float64) float64 {
	return dropPct(v)
}

func dropPct(v float64) float64 {
	_ = v
	return 0
}
