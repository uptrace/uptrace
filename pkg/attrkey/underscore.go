package attrkey

import "github.com/uptrace/uptrace/pkg/unsafeconv"

var underscoreTable = [256]uint8{
	// digit+alpha
	'_': flagAllow,
	'.': '_', // replace
	' ': '_', // replace
	'-': '_', // replace
	'/': '_', // replace
}

func init() {
	for c := '0'; c <= '9'; c++ {
		underscoreTable[c] = flagDigit
	}
	for c := 'a'; c <= 'z'; c++ {
		underscoreTable[c] = flagLC
	}
	for c := 'A'; c <= 'Z'; c++ {
		underscoreTable[c] = flagUC
	}
}

func Underscore(s string) string {
	if isCleanUnderscore(s) {
		return s
	}

	b := make([]byte, 0, len(s))
	b = underscore(b, s)
	return unsafeconv.String(b)
}

func isCleanUnderscore(s string) bool {
	var valid bool
	for _, c := range []byte(s) {
		switch underscoreTable[c] {
		case flagLC, flagDigit:
			valid = true
		case flagAllow:
			// nothing
		default:
			return false
		}
	}
	return valid
}

func underscore(b []byte, s string) []byte {
	var prevFlag uint8
	var accLen int
	for i, ch := range []byte(s) {
		flag := underscoreTable[ch]

		switch flag {
		case flagDisallow: // disallowed
			// ignore the char
		case flagAllow, flagLC, flagDigit:
			b = append(b, ch)
			if ch == '_' {
				accLen = 0
			} else {
				accLen++
			}
		case flagUC:
			if accLen > 1 {
				switch prevFlag {
				case flagLC, flagDigit:
					b = append(b, '_')
					accLen = 0
				case flagUC:
					if i+1 >= len(s) {
						break
					}
					if nextFlag := underscoreTable[s[i+1]]; nextFlag == flagLC {
						b = append(b, '_')
						accLen = 0
					}
				}
			}

			b = append(b, ch+32)
			accLen++
		default:
			// replace the char
			b = append(b, flag)
			accLen++
		}

		prevFlag = flag
	}
	return b
}
