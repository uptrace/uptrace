package attrkey

import "github.com/uptrace/pkg/unsafeconv"

var promTable = [256]uint8{
	// digit+alpha
	'_': flagAllow,
	'.': '_',
}

func init() {
	for c := '0'; c <= '9'; c++ {
		promTable[c] = flagAllow
	}
	for c := 'a'; c <= 'z'; c++ {
		promTable[c] = flagAlpha
	}
	for c := 'A'; c <= 'Z'; c++ {
		promTable[c] = flagAlpha
	}
}

func Clean(s string) string {
	if isClean(s) {
		return s
	}

	b := make([]byte, 0, len(s))
	b = clean(b, s)
	return unsafeconv.String(b)
}

func isClean(s string) bool {
	var valid bool
	for _, c := range []byte(s) {
		switch promTable[c] {
		case flagAlpha:
			valid = true
		case flagAllow:
			// nothing
		default:
			return false
		}
	}
	return valid
}

func clean(b []byte, s string) []byte {
	var prevCh byte
	for _, ch := range []byte(s) {
		switch flag := promTable[ch]; flag {
		case flagAllow, flagAlpha:
			b = append(b, ch)
			prevCh = ch
		case flagDisallow:
			flag = '_'
			fallthrough
		default:
			if prevCh != 0 && flag != prevCh {
				b = append(b, flag)
				prevCh = flag
			}
		}
	}
	if prevCh == '_' {
		b = b[:len(b)-1]
	}
	return b
}

func Valid(s string) bool {
	var valid bool
	for _, c := range []byte(s) {
		switch promTable[c] {
		case flagAlpha:
			valid = true
		case flagAllow:
			// nothing
		default:
			return false
		}
	}
	return valid
}

func isAllowed(flag uint8) bool {
	return flag == flagAllow || flag == flagAlpha
}
