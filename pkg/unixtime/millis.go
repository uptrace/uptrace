package unixtime

import (
	"encoding/json"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

// Millis is a JavaScript style time.Duration.
type Millis float64

func MillisOf(d time.Duration) Millis {
	return Millis(float64(d.Microseconds()) / 1000)
}

func ParseMillis(s string) (Millis, error) {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return Millis(f), nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return MillisOf(d), nil
}

func (ms Millis) Duration() time.Duration {
	return time.Duration(float64(ms) * float64(time.Millisecond))
}

var _ json.Unmarshaler = (*Millis)(nil)

func (ms *Millis) UnmarshalJSON(b []byte) error {
	var str string

	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
	} else {
		str = unsafeconv.String(b)
	}

	got, err := ParseMillis(str)
	if err != nil {
		return err
	}
	*ms = got
	return nil
}

var _ yaml.Unmarshaler = (*Millis)(nil)

func (ms *Millis) UnmarshalYAML(value *yaml.Node) error {
	got, err := ParseMillis(value.Value)
	if err != nil {
		return err
	}
	*ms = got
	return nil
}
