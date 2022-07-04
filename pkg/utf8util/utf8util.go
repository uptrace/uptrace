package utf8util

import "unicode/utf8"

func Trunc(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	maxLen -= 3
	for i := maxLen; i > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		if r == utf8.RuneError {
			i -= size
		} else {
			return s[:i] + "..."
		}
	}
	return "<invalid UTF-8>"
}

func TruncSmall(s string) string {
	return Trunc(s, 100)
}

func TruncMedium(s string) string {
	return Trunc(s, 255)
}

func TruncLarge(s string) string {
	return Trunc(s, 500)
}
