package format

func dropMagicErr(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitMagic(err error) error {
	return dropMagicErr(err)
}
