package bunutil

import (
	"github.com/segmentio/encoding/json"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

func IsJSON(s string) (map[string]any, bool) {
	if len(s) < 2 {
		return nil, false
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return nil, false
	}

	m := make(map[string]any)
	_, err := json.Parse(unsafeconv.Bytes(s), &m, 0)
	if err != nil {
		return nil, false
	}
	return m, true
}
