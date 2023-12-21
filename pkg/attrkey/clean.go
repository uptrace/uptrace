package attrkey

import "github.com/uptrace/uptrace/pkg/unsafeconv"

func AWSMetricName(namespace, metric string) string {
	const prefix = "amazonaws.com."
	return prefix + Clean(namespace) + "." + Clean(metric)
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
	for _, c := range []byte(s) {
		if !isAllowed(c) {
			return false
		}
	}
	return true
}

func clean(b []byte, s string) []byte {
	var prevCh byte
	for _, ch := range []byte(s) {
		if isAllowed(ch) {
			b = append(b, ch)
			prevCh = ch
			continue
		}
		if prevCh != 0 && prevCh != '_' {
			b = append(b, '_')
			prevCh = '_'
		}
	}
	if prevCh == '_' {
		b = b[:len(b)-1]
	}
	return b
}

func isAllowed(c byte) bool {
	return isAlnum(c) || c == '_'
}

func isAlnum(c byte) bool {
	return isDigit(c) || isLC(c) || isUC(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return isLC(c) || isUC(c)
}

func isLC(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func isUC(c byte) bool {
	return c >= 'A' && c <= 'Z'
}
