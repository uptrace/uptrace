package logparser

import (
	"strconv"
	"strings"

	"github.com/segmentio/encoding/json"
)

func IsJSON(s string) (map[string]any, bool) {
	if len(s) < 2 {
		return nil, false
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return nil, false
	}

	m := make(map[string]any)
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, false
	}
	return m, true
}

func IsLogfmt(s string) (string, string, bool) {
	idx := strings.IndexByte(s, '=')
	if idx == -1 {
		return "", "", false
	}

	key := s[:idx]
	if !isIdent(key) {
		return "", "", false
	}

	value := s[idx+1:]
	if len(value) == 0 {
		return key, value, true
	}

	switch value[0] {
	case '"':
		value, err := strconv.Unquote(value)
		if err != nil {
			return "", "", false
		}
		return key, value, true
	}

	if strings.IndexByte(value, ' ') == -1 {
		return key, value, true
	}

	return "", "", false
}
