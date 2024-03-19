package idgen

import (
	"bytes"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strconv"

	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

func ParseSpanID(s string) (SpanID, error) {
	switch s {
	case "null", "0", `"0"`:
		return 0, nil
	}

	switch len(s) {
	case 16:
		if b, err := hex.DecodeString(s); err == nil {
			id := binary.BigEndian.Uint64(b)
			return SpanIDFromUint64(id), nil
		}
	}

	id, err := strconv.ParseUint(s, 10, 64)
	return SpanIDFromUint64(id), err
}

func SpanIDFromBytes(b []byte) SpanID {
	if len(b) == 12 {
		s := base64.RawStdEncoding.EncodeToString(b)

		var err error
		b, err = hex.DecodeString(s)
		if err != nil {
			return 0
		}
	}

	switch len(b) {
	case 0:
		return 0
	case 8:
		id := binary.BigEndian.Uint64(b)
		return SpanIDFromUint64(id)
	default:
		return 0
	}
}

func SpanIDFromUint64(num uint64) SpanID {
	return SpanID(num)
}

func RandSpanID() SpanID {
	return SpanID(rand.Uint64())
}

type SpanID uint64

func (id SpanID) String() string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))

	h := make([]byte, hex.EncodedLen(len(b)))
	n := hex.Encode(h, b)

	return unsafeconv.String(b[:n])
}

func (id SpanID) IsZero() bool {
	return id == 0
}

var _ driver.Valuer = (*SpanID)(nil)

func (id SpanID) Value() (driver.Value, error) {
	return uint64(id), nil
}

var _ json.Unmarshaler = (*SpanID)(nil)

func (id *SpanID) UnmarshalJSON(b []byte) error {
	if len(b) == 0 ||
		bytes.Equal(b, []byte("null")) ||
		bytes.Equal(b, []byte(`"0"`)) ||
		bytes.Equal(b, []byte("0")) {
		*id = 0
		return nil
	}

	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	// For compatability.
	if n, err := strconv.ParseUint(string(b), 10, 64); err == nil {
		*id = SpanIDFromUint64(n)
		return nil
	}

	h := make([]byte, hex.DecodedLen(len(b)))
	n, err := hex.Decode(h, b)
	if err != nil {
		return err
	}

	if n == 8 {
		num := binary.BigEndian.Uint64(h[:n])
		*id = SpanID(num)
		return nil
	}

	return fmt.Errorf("can't decode span id: %q", b)
}

var _ json.Marshaler = (*SpanID)(nil)

func (id SpanID) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return []byte(`"0"`), nil
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))

	h := make([]byte, hex.EncodedLen(len(b))+2)
	h[0] = '"'
	h[len(h)-1] = '"'
	hex.Encode(h[1:], b)

	return h, nil
}
