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

	var prevCh uint8
	for i, ch := range []byte(s) {
		switch ch {
		case '/':
			b = append(b, '.')
		case '-':
			b = append(b, '_')
		case '.', '_':
			b = append(b, ch)
		default:
			if isDigit(ch) {
				b = append(b, ch)
				break
			}

			if isLC(ch) {
				b = append(b, ch)
				break
			}

			if isUC(ch) {
				if isAlnum(prevCh) && i+1 < len(s) {
					if nextCh := s[i+1]; isAlpha(nextCh) && !isUC(nextCh) {
						b = append(b, '_')
					}
				}
				b = append(b, ch+32)
				break
			}
		}

		prevCh = ch
	}

	return unsafeconv.String(b)
}

func isAllowed(c byte) bool {
	return isDigit(c) || isLC(c) || c == '.' || c == '_'
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
