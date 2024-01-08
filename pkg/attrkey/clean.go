package attrkey

const (
	flagDisallow = iota
	flagAllow
	flagDigit
	flagAlpha
	flagLC
	flagUC
)

const maxFlag = flagUC

func isAlpha(flag uint8) bool {
	switch flag {
	case flagLC, flagUC:
		return true
	default:
		return false
	}
}

func isAlnum(flag uint8) bool {
	switch flag {
	case flagLC, flagUC, flagDigit:
		return true
	default:
		return false
	}
}
