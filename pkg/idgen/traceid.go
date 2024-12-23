package idgen

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bun/schema"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/msgp"
	"github.com/uptrace/pkg/unsafeconv"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"
)

const (
	traceIDLen    = 16
	traceIDHexLen = 32
	uuidPrettyLen = 36
)

var (
	traceIDRandMu sync.Mutex
	traceIDRand   = rand.NewPCG(seed1(), seed2())
)

type TraceID [traceIDLen]byte

func RandTraceID() TraceID { return RandTraceIDTime(time.Now()) }
func RandTraceIDTime(tm time.Time) TraceID {
	traceIDRandMu.Lock()
	rnd := traceIDRand.Uint64()
	traceIDRandMu.Unlock()
	var id TraceID
	binary.BigEndian.PutUint64(id[:8], uint64(tm.UnixNano()))
	binary.BigEndian.PutUint64(id[8:], rnd)
	return id
}
func TraceIDFromBytes(b []byte) TraceID {
	var id TraceID
	if len(b) == 24 {
		s := base64.RawStdEncoding.EncodeToString(b)
		var err error
		b, err = hex.DecodeString(s)
		if err != nil {
			slog.Error("can't parse trace id", slog.String("data", hex.EncodeToString(b)), slog.Any("err", err))
			return id
		}
	}
	switch len(b) {
	case 0:
		return id
	case 16:
		copy(id[:], b)
		return id
	default:
		slog.Error("invalid trace id length", slog.Int("length", len(b)))
		return id
	}
}
func NewTraceIDLowHigh(low, high uint64) TraceID {
	var u TraceID
	binary.BigEndian.PutUint64(u[:8], low)
	binary.BigEndian.PutUint64(u[8:], high)
	return u
}
func MustParseTraceID(s string) TraceID {
	id, err := ParseTraceID(s)
	if err != nil {
		panic(err)
	}
	return id
}
func ParseTraceID(s string) (TraceID, error) { return ParseTraceIDBytes([]byte(s)) }
func ParseTraceIDBytes(b []byte) (TraceID, error) {
	var u TraceID
	return u, u.UnmarshalText(b)
}
func (u TraceID) IsZero() bool { return u == TraceID{} }
func (u TraceID) String() string {
	b := appendHex(nil, u[:])
	return string(b)
}
func (u TraceID) Low() uint64  { return binary.BigEndian.Uint64(u[:8]) }
func (u TraceID) High() uint64 { return binary.BigEndian.Uint64(u[8:]) }

var _ schema.QueryAppender = (*TraceID)(nil)

func (u TraceID) AppendQuery(fmter schema.Formatter, b []byte) ([]byte, error) {
	b = append(b, '\'')
	b = appendHex(b, u[:])
	b = append(b, '\'')
	return b, nil
}

var _ driver.Valuer = (*TraceID)(nil)

func (u TraceID) Value() (driver.Value, error) { return u.String(), nil }

var _ sql.Scanner = (*TraceID)(nil)

func (id *TraceID) Scan(src any) error {
	if src == nil {
		for i := range id {
			id[i] = 0
		}
		return nil
	}
	switch src := src.(type) {
	case []byte:
		return id.UnmarshalBinary(src)
	case string:
		return id.UnmarshalText(unsafeconv.Bytes(src))
	case chschema.UUID:
		copy(id[:], src[:])
		return nil
	default:
		return fmt.Errorf("unsupported TraceID source: %T", src)
	}
}

var _ encoding.BinaryMarshaler = (*TraceID)(nil)

func (id TraceID) MarshalBinary() ([]byte, error) { return id[:], nil }

var _ encoding.BinaryUnmarshaler = (*TraceID)(nil)

func (id *TraceID) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case traceIDLen:
		copy(id[:], b)
		return nil
	case traceIDHexLen:
		_, err := hex.Decode(id[:], b)
		return err
	case uuidPrettyLen:
		return id.unmarshalPretty(b)
	}
	return fmt.Errorf("can't parse TraceID: %q", b)
}

var _ encoding.TextMarshaler = (*TraceID)(nil)

func (u TraceID) MarshalText() ([]byte, error) { return appendHex(nil, u[:]), nil }

var _ encoding.TextUnmarshaler = (*TraceID)(nil)

func (u *TraceID) UnmarshalText(buf []byte) error {
	switch len(buf) {
	case traceIDHexLen:
		_, err := hex.Decode(u[:], buf)
		return err
	case uuidPrettyLen:
		return u.unmarshalPretty(buf)
	case traceIDHexLen + 1:
		if buf[8] == '-' {
			return u.unmarshalAmazon(buf)
		}
	case 4:
		if bytes.Equal(buf, []byte("null")) {
			return nil
		}
	}
	return fmt.Errorf("can't parse TraceID: %q", buf)
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

func (u *TraceID) UnmarshalJSON(buf []byte) error {
	if len(buf) >= 2 && buf[0] == '"' && buf[len(buf)-1] == '"' {
		buf = buf[1 : len(buf)-1]
	}
	return u.UnmarshalText(buf)
}

var _ msgp.Sizer = (*TraceID)(nil)

func (TraceID) MsgpackSize() int { return 16 }

var _ msgp.Appender = (*TraceID)(nil)

func (u TraceID) AppendMsgpack(buf []byte, flags msgp.AppendFlags) (_ []byte, err error) {
	return msgp.AppendBytes(buf, u[:]), nil
}

var _ msgp.Parser = (*TraceID)(nil)

func (u *TraceID) ParseMsgpack(b []byte, flags msgp.ParseFlags) (_ []byte, err error) {
	bs, b, err := msgp.ParseBytes(b, msgp.ZeroCopyBytes)
	if err != nil {
		return nil, err
	}
	copy(u[:], bs)
	return b, nil
}
func appendHex(b []byte, u []byte) []byte {
	b = append(b, make([]byte, traceIDHexLen)...)
	hex.Encode(b[len(b)-traceIDHexLen:], u)
	return b
}
func (id *TraceID) UUID() string {
	dest := make([]byte, 36)
	encodePretty(dest, id[:])
	return unsafeconv.String(dest)
}
func encodePretty(dest []byte, id []byte) {
	hex.Encode(dest, id[:4])
	dest[8] = '-'
	hex.Encode(dest[9:13], id[4:6])
	dest[13] = '-'
	hex.Encode(dest[14:18], id[6:8])
	dest[18] = '-'
	hex.Encode(dest[19:23], id[8:10])
	dest[23] = '-'
	hex.Encode(dest[24:], id[10:])
}
func ShouldSample(traceID TraceID, fraction float64) bool {
	traceIDUpperBound := uint64(fraction * (1 << 63))
	x := binary.BigEndian.Uint64(traceID[8:16]) >> 1
	return x < traceIDUpperBound
}
