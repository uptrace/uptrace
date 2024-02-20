package idgen

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/uptrace/bun/schema"
)

const (
	traceIDLen    = 16
	traceIDHexLen = 32
	uuidPrettyLen = 36
)

var (
	traceIDRandMu sync.Mutex
	traceIDRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type TraceID [traceIDLen]byte

func RandTraceID() TraceID {
	var id TraceID

	traceIDRandMu.Lock()
	traceIDRand.Read(id[:])
	traceIDRandMu.Unlock()

	return id
}

func TraceIDFromBytes(b []byte) TraceID {
	var id TraceID

	if len(b) == 24 {
		s := base64.RawStdEncoding.EncodeToString(b)

		var err error
		b, err = hex.DecodeString(s)
		if err != nil {
			return id
		}
	}

	switch len(b) {
	case 0:
		// This is okay when trace id is absent.
		return id
	case 16:
		copy(id[:], b)
		return id
	default:
		return id
	}
}

func MustParseTraceID(s string) TraceID {
	id, err := ParseTraceID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func ParseTraceID(s string) (TraceID, error) {
	return ParseTraceIDBytes([]byte(s))
}

func ParseTraceIDBytes(b []byte) (TraceID, error) {
	var u TraceID
	return u, u.UnmarshalText(b)
}

func (u TraceID) IsZero() bool {
	return u == TraceID{}
}

func (u TraceID) String() string {
	b := appendHex(nil, u[:])
	return string(b)
}

func (u TraceID) Low() uint64 {
	return binary.BigEndian.Uint64(u[:8])
}

func (u TraceID) High() uint64 {
	return binary.BigEndian.Uint64(u[8:])
}

var _ schema.QueryAppender = (*TraceID)(nil)

func (u TraceID) AppendQuery(fmter schema.Formatter, b []byte) ([]byte, error) {
	b = append(b, '\'')
	b = appendHex(b, u[:])
	b = append(b, '\'')
	return b, nil
}

var _ driver.Valuer = (*TraceID)(nil)

func (u TraceID) Value() (driver.Value, error) {
	return u.String(), nil
}

var _ sql.Scanner = (*TraceID)(nil)

func (u *TraceID) Scan(src any) error {
	if src == nil {
		for i := range u {
			u[i] = 0
		}
		return nil
	}

	var uuid TraceID
	var err error

	switch src := src.(type) {
	case []byte:
		uuid, err = ParseTraceIDBytes(src)
	case string:
		uuid, err = ParseTraceID(src)
	}
	if err != nil {
		return err
	}

	copy(u[:], uuid[:])

	return nil
}

var _ encoding.BinaryMarshaler = (*TraceID)(nil)

func (u TraceID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

var _ encoding.BinaryUnmarshaler = (*TraceID)(nil)

func (u *TraceID) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case traceIDLen:
		copy(u[:], b)
		return nil
	case traceIDHexLen:
		_, err := hex.Decode(u[:], b)
		return err
	}
	return fmt.Errorf("can't parse TraceID: %q", b)
}

var _ encoding.TextMarshaler = (*TraceID)(nil)

func (u TraceID) MarshalText() ([]byte, error) {
	return appendHex(nil, u[:]), nil
}

var _ encoding.TextUnmarshaler = (*TraceID)(nil)

func (u *TraceID) UnmarshalText(b []byte) error {
	switch len(b) {
	case traceIDHexLen:
		_, err := hex.Decode(u[:], b)
		return err
	case uuidPrettyLen:
		return u.unmarshalPretty(b)
	case traceIDHexLen + 1:
		if b[8] == '-' {
			return u.unmarshalAmazon(b)
		}
	}
	return fmt.Errorf("can't parse TraceID: %q", b)
}

func (u *TraceID) unmarshalPretty(b []byte) error {
	if _, err := hex.Decode(u[:4], b[:8]); err != nil {
		return err
	}
	if _, err := hex.Decode(u[4:6], b[9:13]); err != nil {
		return err
	}
	if _, err := hex.Decode(u[6:8], b[14:18]); err != nil {
		return err
	}
	if _, err := hex.Decode(u[8:10], b[19:23]); err != nil {
		return err
	}
	if _, err := hex.Decode(u[10:], b[24:]); err != nil {
		return err
	}
	return nil
}

func (u *TraceID) unmarshalAmazon(b []byte) error {
	if _, err := hex.Decode(u[:4], b[:8]); err != nil {
		return err
	}
	if _, err := hex.Decode(u[4:], b[9:]); err != nil {
		return err
	}
	return nil
}

var _ json.Marshaler = (*TraceID)(nil)

func (u TraceID) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte("null"), nil
	}

	b := make([]byte, 0, traceIDHexLen+2)
	b = append(b, '"')
	b = appendHex(b, u[:])
	b = append(b, '"')
	return b, nil
}

var _ json.Unmarshaler = (*TraceID)(nil)

func (u *TraceID) UnmarshalJSON(b []byte) error {
	if len(b) >= 2 {
		b = b[1 : len(b)-1]
	}
	return u.UnmarshalText(b)
}

func appendHex(b []byte, u []byte) []byte {
	b = append(b, make([]byte, traceIDHexLen)...)
	hex.Encode(b[len(b)-traceIDHexLen:], u)
	return b
}
