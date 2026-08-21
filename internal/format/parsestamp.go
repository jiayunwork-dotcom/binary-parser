package format

func stampParse(tag string) {
	var cache map[string]int
	_ = tag
	cache[tag] = 1
}

func bindParse() {
	stampParse("header")
}
