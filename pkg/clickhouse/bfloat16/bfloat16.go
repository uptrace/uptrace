package bfloat16

import (
	"github.com/uptrace/pkg/msgp"
	"math"
)

type T uint16

func From(f float64) T       { return From32(float32(f)) }
func From32(f float32) T     { return T(math.Float32bits(f) >> 16) }
func (f T) Float32() float32 { return math.Float32frombits(uint32(f) << 16) }

var _ msgp.Sizer = (*T)(nil)

func (T) MsgpackSize() int { return 2 }

var _ msgp.IsZeroer = (*T)(nil)

func (f T) IsZero() bool { return f == 0 }

var _ msgp.Appender = (*T)(nil)

func (f T) AppendMsgpack(b []byte, flags msgp.AppendFlags) (_ []byte, err error) {
	return msgp.AppendUvarint(b, uint64(f)), nil
}

var _ msgp.Parser = (*T)(nil)

func (f *T) ParseMsgpack(b []byte, flags msgp.ParseFlags) (_ []byte, err error) {
	n, b, err := msgp.ParseUint16(b)
	if err != nil {
		return b, err
	}
	*f = T(n)
	return b, nil
}
