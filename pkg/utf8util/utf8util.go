package utf8util

import "unicode/utf8"

func Trunc(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	for i := maxLen; i > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		if r == utf8.RuneError {
			i -= size
		} else {
			return s[:i]
		}
	}
	return "<invalid UTF-8>"
}

func TruncLC(s string) string {
	return Trunc(s, 80)
}

func TruncSmall(s string) string {
	return Trunc(s, 200)
}

func TruncMedium(s string) string {
	return Trunc(s, 500)
}

func TruncLarge(s string) string {
	return Trunc(s, 1000)
}
