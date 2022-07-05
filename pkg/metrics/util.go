package metrics

import "github.com/uptrace/uptrace/pkg/unsafeconv"

func listToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//------------------------------------------------------------------------------

func cleanPromName(s string) string {
	if isValidPromName(s) {
		return s
	}

	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowedPromNameChar(c) {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return unsafeconv.String(r)
}

func isValidPromName(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowedPromNameChar(c) {
			return false
		}
	}
	return true
}

func isAllowedPromNameChar(c byte) bool {
	return isAlpha(c) || isDigit(c) || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
