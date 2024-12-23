package msgp

import (
	"fmt"
	"github.com/uptrace/pkg/msgp/msgpcode"
)

func Skip(b []byte) ([]byte, error) {
	flags := ZeroCopy
	c, err := peekByte(b)
	if err != nil {
		return b, err
	}
	if msgpcode.IsFixedString(c) {
		_, b, err := ParseString(b, flags)
		return b, err
	}
	if msgpcode.IsFixedNum(c) {
		return b[1:], nil
	}
	if msgpcode.IsFixedMap(c) {
		return skipMap(b)
	}
	if msgpcode.IsFixedArray(c) {
		return skipSlice(b)
	}
	switch c {
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32, msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		_, b, err := ParseString(b, flags)
		return b, err
	case msgpcode.Uint8, msgpcode.Uint16, msgpcode.Uint32, msgpcode.Uint64:
		_, b, err := ParseUint64(b)
		return b, err
	case msgpcode.Int8, msgpcode.Int16, msgpcode.Int32, msgpcode.Int64:
		_, b, err := ParseInt64(b)
		return b, err
	case msgpcode.Float, msgpcode.Double:
		_, b, err := ParseFloat64(b)
		return b, err
	case msgpcode.False, msgpcode.True:
		_, b, err := ParseBool(b)
		return b, err
	case msgpcode.Array16, msgpcode.Array32:
		return skipSlice(b)
	case msgpcode.Map16, msgpcode.Map32:
		return skipMap(b)
	case msgpcode.FixExt1, msgpcode.FixExt2, msgpcode.FixExt4, msgpcode.FixExt8, msgpcode.FixExt16, msgpcode.Ext8, msgpcode.Ext16, msgpcode.Ext32:
		_, b, err := parseExt(b)
		return b, err
	case msgpcode.Nil:
		return b[1:], nil
	default:
		return b, fmt.Errorf("msgp: unknown code %x decoding interface{}", c)
	}
}
func skipSlice(b []byte) ([]byte, error) {
	ln, b, err := ParseArrayLen(b)
	if err != nil {
		return b, err
	}
	if ln == -1 {
		return b, nil
	}
	for i := 0; i < ln; i++ {
		b, err = Skip(b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}
func skipMap(b []byte) ([]byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return b, err
	}
	if ln == -1 {
		return b, nil
	}
	for i := 0; i < ln; i++ {
		b, err = Skip(b)
		if err != nil {
			return b, err
		}
		b, err = Skip(b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}
