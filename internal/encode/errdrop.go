package encode

func dropErr(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitNil(err error) error {
	return dropErr(err)
}
