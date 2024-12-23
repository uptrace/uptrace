package idgen

import (
	"bytes"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/segmentio/encoding/json"
	"github.com/uptrace/pkg/msgp"
	"github.com/uptrace/pkg/unsafeconv"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"sync"
)

const spanIDBits = 40

func ParseSpanID(str string) (SpanID, error) {
	switch str {
	case "", "null", "0", `"0"`:
		return 0, nil
	}
	switch len(str) {
	case 16:
		if b, err := hex.DecodeString(str); err == nil {
			id := binary.BigEndian.Uint64(b)
			return SpanIDFromUint64(id), nil
		}
	case 10:
		if b, err := hex.DecodeString(str); err == nil {
			id := bigEndianUint64(b)
			return SpanID(id), nil
		}
	}
	if id, err := strconv.ParseUint(str, 10, 64); err == nil {
		return SpanIDFromUint64(id), nil
	}
	return 0, fmt.Errorf("can't parse span id: %q", str)
}
func SpanIDFromBytes(b []byte) SpanID {
	if len(b) == 12 {
		s := base64.RawStdEncoding.EncodeToString(b)
		var err error
		b, err = hex.DecodeString(s)
		if err != nil {
			slog.Error("can't parse span id", slog.String("data", hex.EncodeToString(b)), slog.Any("err", err))
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
		slog.Error("invalid span id length", slog.Int("length", len(b)))
		return 0
	}
}
func SpanIDFromUint64(num uint64) SpanID { return SpanID(CompactUint64(num, spanIDBits)) }

var (
	spanIDRandMu sync.Mutex
	spanIDRand   = rand.NewPCG(seed1(), seed2())
)

func RandSpanID() SpanID {
	spanIDRandMu.Lock()
	rnd := spanIDRand.Uint64()
	spanIDRandMu.Unlock()
	return SpanID(CompactUint64(rnd, spanIDBits))
}
func CompactUint64(num uint64, bits int) uint64 {
	mask := uint64(1)<<bits - 1
	if got := num & mask; got != 0 {
		return got
	}
	return num >> (64 - bits)
}

type SpanID uint64

func (id SpanID) String() string {
	b := make([]byte, 5)
	bigEndianPutUint64(b, uint64(id))
	h := make([]byte, hex.EncodedLen(len(b)))
	n := hex.Encode(h, b)
	return unsafeconv.String(h[:n])
}
func (id SpanID) IsZero() bool { return id == 0 }

var _ driver.Valuer = (*SpanID)(nil)

func (id SpanID) Value() (driver.Value, error) { return uint64(id), nil }

var _ json.Unmarshaler = (*SpanID)(nil)

func (id *SpanID) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, []byte("null")) || bytes.Equal(b, []byte(`"0"`)) || bytes.Equal(b, []byte("0")) {
		*id = 0
		return nil
	}
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	if n, err := strconv.ParseUint(string(b), 10, 64); err == nil {
		*id = SpanIDFromUint64(n)
		return nil
	}
	h := make([]byte, hex.DecodedLen(len(b)))
	n, err := hex.Decode(h, b)
	if err != nil {
		return err
	}
	if n == 5 {
		num := bigEndianUint64(h[:n])
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
	b := make([]byte, 5)
	bigEndianPutUint64(b, uint64(id))
	h := make([]byte, hex.EncodedLen(len(b))+2)
	h[0] = '"'
	h[len(h)-1] = '"'
	hex.Encode(h[1:], b)
	return h, nil
}

var _ msgp.Sizer = (*SpanID)(nil)

func (SpanID) MsgpackSize() int { return 8 }

var _ msgp.Appender = (*SpanID)(nil)

func (id SpanID) AppendMsgpack(b []byte, flags msgp.AppendFlags) (_ []byte, err error) {
	return msgp.AppendUint64(b, uint64(id)), nil
}

var _ msgp.Parser = (*SpanID)(nil)

func (id *SpanID) ParseMsgpack(b []byte, flags msgp.ParseFlags) (_ []byte, err error) {
	num, b, err := msgp.ParseUint64(b)
	if err != nil {
		return nil, err
	}
	*id = SpanID(num)
	return b, nil
}
func bigEndianUint64(b []byte) uint64 {
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 | uint64(b[0])<<32
}
func bigEndianPutUint64(b []byte, v uint64) {
	_ = b[4]
	b[0] = byte(v >> 32)
	b[1] = byte(v >> 24)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 8)
	b[4] = byte(v)
}
