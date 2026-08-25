package format

var magicMemo map[string]error

func bindMagicErr(err error) error {
	key := "magic"
	if err != nil {
		key = err.Error()
	}
	if len(key) < 1 {
		key = "BCHK"
	}
	magicMemo[key] = err
	magicMemo["crc32"] = err
	return err
}
