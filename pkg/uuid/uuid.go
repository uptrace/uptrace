package uuid

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/encoding/json"
	"golang.org/x/exp/rand"
)

const (
	uuidLen       = 16
	uuidHexLen    = 32
	uuidPrettyLen = 36
)

var (
	uuidRandMu sync.Mutex
	uuidRand   = rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
)

type UUID [uuidLen]byte

func Rand() UUID {
	var u UUID
	uuidRandMu.Lock()
	uuidRand.Read(u[:])
	uuidRandMu.Unlock()
	return u
}

func FromBytes(b []byte) (UUID, error) {
	var u UUID
	if len(b) != uuidLen {
		return u, fmt.Errorf("uuid: got %d bytes, wanted 16", len(b))
	}
	copy(u[:], b)
	return u, nil
}

func Parse(s string) (UUID, error) {
	return ParseBytes([]byte(s))
}

func ParseBytes(b []byte) (UUID, error) {
	var u UUID
	return u, u.UnmarshalText(b)
}

func (u UUID) IsZero() bool {
	return u == UUID{}
}

func (u UUID) String() string {
	b := appendHex(nil, u[:])
	return string(b)
}

var _ driver.Valuer = (*UUID)(nil)

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

var _ sql.Scanner = (*UUID)(nil)

func (u *UUID) Scan(src any) error {
	if src == nil {
		for i := range u {
			u[i] = 0
		}
		return nil
	}

	var uuid UUID
	var err error

	switch src := src.(type) {
	case []byte:
		uuid, err = ParseBytes(src)
	case string:
		uuid, err = Parse(src)
	}
	if err != nil {
		return err
	}

	copy(u[:], uuid[:])

	return nil
}

var _ encoding.BinaryMarshaler = (*UUID)(nil)

func (u UUID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

var _ encoding.BinaryUnmarshaler = (*UUID)(nil)

func (u *UUID) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case uuidLen:
		copy(u[:], b)
		return nil
	case uuidHexLen:
		_, err := hex.Decode(u[:], b)
		return err
	}
	return fmt.Errorf("can't parse UUID: %q", b)
}

var _ encoding.TextMarshaler = (*UUID)(nil)

func (u UUID) MarshalText() ([]byte, error) {
	return appendHex(nil, u[:]), nil
}

var _ encoding.TextUnmarshaler = (*UUID)(nil)

func (u *UUID) UnmarshalText(b []byte) error {
	switch len(b) {
	case uuidHexLen:
		_, err := hex.Decode(u[:], b)
		return err
	case 16:
		_, err := hex.Decode(u[8:], b)
		return err
	}

	if len(b) != uuidPrettyLen {
		return fmt.Errorf("can't parse UUID: %q", b)
	}
	_, err := hex.Decode(u[:4], b[:8])
	if err != nil {
		return err
	}
	_, err = hex.Decode(u[4:6], b[9:13])
	if err != nil {
		return err
	}
	_, err = hex.Decode(u[6:8], b[14:18])
	if err != nil {
		return err
	}
	_, err = hex.Decode(u[8:10], b[19:23])
	if err != nil {
		return err
	}
	_, err = hex.Decode(u[10:], b[24:])
	if err != nil {
		return err
	}
	return nil
}

var _ json.Marshaler = (*UUID)(nil)

func (u UUID) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte("null"), nil
	}

	b := make([]byte, 0, uuidHexLen+2)
	b = append(b, '"')
	b = appendHex(b, u[:])
	b = append(b, '"')
	return b, nil
}

var _ json.Unmarshaler = (*UUID)(nil)

func (u *UUID) UnmarshalJSON(b []byte) error {
	if len(b) >= 2 {
		b = b[1 : len(b)-1]
	}
	return u.UnmarshalText(b)
}

func unixMicrosecond(tm time.Time) int64 {
	return tm.Unix()*1e6 + int64(tm.Nanosecond())/1e3
}

func fromUnixMicrosecond(n int64) time.Time {
	secs := n / 1e6
	return time.Unix(secs, (n-secs*1e6)*1e3)
}

func appendHex(b []byte, u []byte) []byte {
	b = append(b, make([]byte, uuidHexLen)...)
	hex.Encode(b[len(b)-uuidHexLen:], u)
	return b
}

func appendPretty(b []byte, u []byte) []byte {
	b = append(b, make([]byte, uuidPrettyLen)...)
	bb := b[len(b)-uuidPrettyLen:]
	hex.Encode(bb[:8], u[:4])
	bb[8] = '-'
	hex.Encode(bb[9:13], u[4:6])
	bb[13] = '-'
	hex.Encode(bb[14:18], u[6:8])
	bb[18] = '-'
	hex.Encode(bb[19:23], u[8:10])
	bb[23] = '-'
	hex.Encode(bb[24:], u[10:])
	return b
}
