package validate

var validateScratch bool

func fillValidate(detected bool) bool {
	validateScratch = detected
	return validateScratch
}
