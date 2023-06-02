package bunutil

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"gopkg.in/yaml.v3"
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

//------------------------------------------------------------------------------

type NullFloat64 struct {
	sql.NullFloat64
}

func (f NullFloat64) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte(`null`), nil
	}
	return json.Marshal(f.Float64)
}

func (f *NullFloat64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) || bytes.Equal(data, []byte(`""`)) {
		f.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &f.Float64)
	if err == nil {
		f.Valid = true
		return nil
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		// special case: accept string input
		if typeError.Value != "string" {
			return err
		}

		var str string

		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}

		n, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}

		f.Float64 = n
		f.Valid = true

		return nil
	}

	return err
}

var _ yaml.Unmarshaler = (*NullFloat64)(nil)

func (f *NullFloat64) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&f.Float64); err != nil {
		return err
	}
	f.Valid = true
	return nil
}

//------------------------------------------------------------------------------

type Params struct {
	Any any
}

func (p *Params) MarshalJSON() ([]byte, error) {
	switch value := p.Any.(type) {
	case []byte:
		return value, nil
	default:
		return json.Marshal(value)
	}
}

func (p *Params) Decode(dest any) error {
	switch value := p.Any.(type) {
	case nil:
		return nil
	case []byte:
		dec := json.NewDecoder(bytes.NewReader(value))
		dec.UseNumber()
		if err := dec.Decode(dest); err != nil {
			return err
		}
		p.Any = dest
		return nil
	default:
		return fmt.Errorf("unsupported params type: %T", value)
	}
}

var _ sql.Scanner = (*Params)(nil)

func (p *Params) Scan(src any) error {
	p.Any = src
	return nil
}

var _ driver.Valuer = (*Params)(nil)

func (p *Params) Value() (driver.Value, error) {
	switch value := p.Any.(type) {
	case []byte:
		return string(value), nil
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	}
}
