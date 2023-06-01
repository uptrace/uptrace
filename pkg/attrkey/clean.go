package attrkey

import "github.com/uptrace/uptrace/pkg/unsafeconv"

func Clean(s string) string {
	if isClean(s) {
		return s
	}
	return clean(s)
}

func isClean(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowed(c) {
			return false
		}
	}
	return true
}

func clean(s string) string {
	b := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowed(c) {
			b = append(b, c)
			continue
		}
		if c == '/' {
			b = append(b, '.')
			continue
		}
		if c == '-' {
			b = append(b, '_')
			continue
		}
		if c >= 'A' && c <= 'Z' {
			b = append(b, c+32)
		}
	}
	return unsafeconv.String(b)
}

func isAllowed(c byte) bool {
	return isAlnum(c) || c == '.' || c == '_'
}

func isAlnum(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z'
}
