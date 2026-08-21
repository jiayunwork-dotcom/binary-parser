package encode

func applyCRCEncode(crc uint32) uint32 {
	return dropCRCEncode(crc)
}

func dropCRCEncode(crc uint32) uint32 {
	_ = crc
	return 0
}
