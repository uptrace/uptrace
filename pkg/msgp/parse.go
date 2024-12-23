package msgp

import (
	"fmt"
	"github.com/uptrace/pkg/msgp/msgpcode"
	"io"
	"math"
	"time"
	"unsafe"
)

func ParseAny(b []byte, flags ParseFlags) (any, []byte, error) {
	c, err := peekByte(b)
	if err != nil {
		return nil, nil, err
	}
	if msgpcode.IsFixedString(c) {
		s, b, err := ParseString(b, flags)
		return s, b, err
	}
	if msgpcode.IsFixedNum(c) {
		return int64(int8(c)), b[1:], nil
	}
	if msgpcode.IsFixedMap(c) {
		return parseMapDefault(b, flags)
	}
	if msgpcode.IsFixedArray(c) {
		return ParseSlice(b, flags)
	}
	switch c {
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32, msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		s, b, err := ParseString(b, flags)
		return s, b, err
	case msgpcode.Uint8, msgpcode.Uint16, msgpcode.Uint32, msgpcode.Uint64:
		return ParseUint64(b)
	case msgpcode.Int8, msgpcode.Int16, msgpcode.Int32, msgpcode.Int64:
		return ParseInt64(b)
	case msgpcode.Float, msgpcode.Double:
		return ParseFloat64(b)
	case msgpcode.False, msgpcode.True:
		return ParseBool(b)
	case msgpcode.Array16, msgpcode.Array32:
		return ParseSlice(b, flags)
	case msgpcode.Map16, msgpcode.Map32:
		return parseMapDefault(b, flags)
	case msgpcode.FixExt1, msgpcode.FixExt2, msgpcode.FixExt4, msgpcode.FixExt8, msgpcode.FixExt16, msgpcode.Ext8, msgpcode.Ext16, msgpcode.Ext32:
		return parseExt(b)
	case msgpcode.Nil:
		return nil, b[1:], nil
	}
	return nil, b, fmt.Errorf("msgp: unknown code %x decoding interface{}", c)
}
func ParseBool(b []byte) (bool, []byte, error) {
	saved := b
	c, b, err := readByte(b)
	if err != nil {
		return false, b, err
	}
	switch c {
	case msgpcode.True:
		return true, b, nil
	case msgpcode.False:
		return false, b, nil
	case msgpcode.Nil:
		return false, b, nil
	default:
		return false, saved, fmt.Errorf("msgp: unexpected code %x decoding bool", c)
	}
}
func ParseUint(b []byte) (uint, []byte, error) {
	n, b, err := ParseUint64(b)
	return uint(n), b, err
}
func ParseUint8(b []byte) (uint8, []byte, error) {
	n, b, err := ParseUint64(b)
	return uint8(n), b, err
}
func ParseUint16(b []byte) (uint16, []byte, error) {
	n, b, err := ParseInt64(b)
	return uint16(n), b, err
}
func ParseUint32(b []byte) (uint32, []byte, error) {
	n, b, err := ParseUint64(b)
	return uint32(n), b, err
}
func ParseUint64(b []byte) (uint64, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c == msgpcode.Nil {
		return 0, b, nil
	}
	if msgpcode.IsFixedNum(c) {
		return uint64(int8(c)), b, nil
	}
	switch c {
	case msgpcode.Uint8, msgpcode.Int8:
		n, b, err := readByte(b)
		return uint64(n), b, err
	case msgpcode.Uint16, msgpcode.Int16:
		n, b, err := parseUint16(b)
		return uint64(n), b, err
	case msgpcode.Uint32, msgpcode.Int32:
		n, b, err := parseUint32(b)
		return uint64(n), b, err
	case msgpcode.Uint64, msgpcode.Int64:
		return parseUint64(b)
	default:
		return 0, b, fmt.Errorf("msgp: unexpected code %x decoding uint", c)
	}
}
func ParseInt(b []byte) (int, []byte, error) {
	n, b, err := ParseInt64(b)
	return int(n), b, err
}
func ParseInt8(b []byte) (int8, []byte, error) {
	n, b, err := ParseInt64(b)
	return int8(n), b, err
}
func ParseInt16(b []byte) (int16, []byte, error) {
	n, b, err := ParseInt64(b)
	return int16(n), b, err
}
func ParseInt32(b []byte) (int32, []byte, error) {
	n, b, err := ParseInt64(b)
	return int32(n), b, err
}
func ParseInt64(b []byte) (int64, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c == msgpcode.Nil {
		return 0, b, nil
	}
	if msgpcode.IsFixedNum(c) {
		return int64(int8(c)), b, nil
	}
	switch c {
	case msgpcode.Int8:
		n, b, err := readByte(b)
		return int64(int8(n)), b, err
	case msgpcode.Int16:
		n, b, err := parseUint16(b)
		return int64(int16(n)), b, err
	case msgpcode.Int32:
		n, b, err := parseUint32(b)
		return int64(int32(n)), b, err
	case msgpcode.Uint8:
		n, b, err := readByte(b)
		return int64(n), b, err
	case msgpcode.Uint16:
		n, b, err := parseUint16(b)
		return int64(n), b, err
	case msgpcode.Uint32:
		n, b, err := parseUint32(b)
		return int64(n), b, err
	case msgpcode.Uint64, msgpcode.Int64:
		n, b, err := parseUint64(b)
		return int64(n), b, err
	default:
		return 0, b, fmt.Errorf("msgp: unexpected code %x decoding int", c)
	}
}
func parseUint16(b []byte) (uint16, []byte, error) {
	if len(b) < 2 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	_ = b[1]
	n := uint16(b[1]) | uint16(b[0])<<8
	return n, b[2:], nil
}
func parseUint32(b []byte) (uint32, []byte, error) {
	if len(b) < 4 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	_ = b[3]
	n := uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	return n, b[4:], nil
}
func parseUint64(b []byte) (uint64, []byte, error) {
	if len(b) < 8 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	_ = b[7]
	n := uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 | uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	return n, b[8:], nil
}
func ParseFloat32(b []byte) (float32, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c != msgpcode.Float {
		return 0, b, fmt.Errorf("msgp: unexpected code %x decoding float32", c)
	}
	n, b, err := parseUint32(b)
	if err != nil {
		return 0, b, err
	}
	return math.Float32frombits(n), b, nil
}
func ParseFloat64(b []byte) (float64, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	switch c {
	case msgpcode.Float:
		n, b, err := parseUint32(b)
		if err != nil {
			return 0, b, err
		}
		return float64(math.Float32frombits(n)), b, nil
	case msgpcode.Double:
		n, b, err := parseUint64(b)
		if err != nil {
			return 0, b, err
		}
		return math.Float64frombits(n), b, nil
	default:
		return 0, b, fmt.Errorf("msgp: unexpected code %x decoding float64", c)
	}
}
func ParseString(b []byte, flags ParseFlags) (string, []byte, error) {
	ln, b, err := ParseStringLen(b)
	if err != nil {
		return "", b, err
	}
	if ln <= 0 {
		return "", b, nil
	}
	if len(b) < ln {
		return "", b, fmt.Errorf("msgp: wanted %d bytes decoding string, got %d", ln, len(b))
	}
	var s string
	if flags&ZeroCopyString != 0 {
		bs := b[:ln]
		s = *(*string)(unsafe.Pointer(&bs))
	} else {
		s = string(b[:ln])
	}
	return s, b[ln:], nil
}
func ParseBytes(b []byte, flags ParseFlags) ([]byte, []byte, error) {
	ln, b, err := ParseStringLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln <= 0 {
		return nil, b, nil
	}
	if len(b) < ln {
		return nil, b, fmt.Errorf("msgp: wanted %d bytes decoding bytes, got %d", ln, len(b))
	}
	var bs []byte
	if flags&ZeroCopyBytes != 0 {
		bs = b[:ln]
	} else {
		bs = make([]byte, ln)
		copy(bs, b)
	}
	return bs, b[ln:], nil
}
func ParseStringLen(b []byte) (int, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c == msgpcode.Nil {
		return -1, b, nil
	}
	if msgpcode.IsFixedString(c) {
		return int(c & msgpcode.FixedStrMask), b, nil
	}
	switch c {
	case msgpcode.Str8, msgpcode.Bin8:
		c, b, err := readByte(b)
		return int(c), b, err
	case msgpcode.Str16, msgpcode.Bin16:
		n, b, err := parseUint16(b)
		return int(n), b, err
	case msgpcode.Str32, msgpcode.Bin32:
		n, b, err := parseUint32(b)
		return int(n), b, err
	}
	return 0, b, fmt.Errorf("msgp: unexpected code %x decoding string len", c)
}
func ParseBytesLen(b []byte) (int, []byte, error) { return ParseStringLen(b) }
func ParseTime(b []byte) (time.Time, []byte, error) {
	extID, size, b, err := parseExtHeader(b)
	if err != nil {
		return time.Time{}, b, err
	}
	switch extID {
	case msgpcode.Nil:
		return time.Time{}, b, nil
	case timeExtID:
		return parseTimeData(b, size)
	default:
		return time.Time{}, b, fmt.Errorf("msgp: unexpected code %x decoding time ext id", extID)
	}
}
func parseTimeData(b []byte, size int) (time.Time, []byte, error) {
	if len(b) < size {
		return time.Time{}, b, fmt.Errorf("msgp: wanted %d bytes decoding time, got %d", size, len(b))
	}
	switch size {
	case 4:
		sec, b, err := parseUint32(b)
		if err != nil {
			return time.Time{}, b, err
		}
		return time.Unix(int64(sec), 0), b, nil
	case 8:
		sec, b, err := parseUint64(b)
		if err != nil {
			return time.Time{}, b, err
		}
		nsec := int64(sec >> 34)
		sec &= 0x00000003ffffffff
		return time.Unix(int64(sec), nsec), b, nil
	case 12:
		nsec, b, err := parseUint32(b)
		if err != nil {
			return time.Time{}, b, err
		}
		sec, b, err := parseUint64(b)
		if err != nil {
			return time.Time{}, b, err
		}
		return time.Unix(int64(sec), int64(nsec)), b, nil
	default:
		return time.Time{}, b, fmt.Errorf("msgp: unknown time data size: %d", size)
	}
}
func parseMapDefault(b []byte, flags ParseFlags) (any, []byte, error) {
	if m, b, err := ParseMapStringAny(b, flags); err == nil {
		return m, b, nil
	}
	return ParseMapAnyAny(b, flags)
}
func ParseMapAnyAny(b []byte, flags ParseFlags) (map[any]any, []byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln == -1 {
		return nil, b, nil
	}
	m := make(map[any]any, LenOrZero(ln, flags))
	for i := 0; i < ln; i++ {
		var key any
		key, b, err = ParseAny(b, flags)
		if err != nil {
			return nil, b, err
		}
		var val any
		val, b, err = ParseAny(b, flags)
		if err != nil {
			return nil, b, err
		}
		mapSetSafe(m, key, val)
	}
	return m, b, nil
}
func mapSetSafe(m map[any]any, key, val any) {
	defer func() { _ = recover() }()
	m[key] = val
}
func ParseMapStringAny(b []byte, flags ParseFlags) (map[string]any, []byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln == -1 {
		return nil, b, nil
	}
	m := make(map[string]any, LenOrZero(ln, flags))
	for i := 0; i < ln; i++ {
		var key string
		key, b, err = ParseString(b, flags)
		if err != nil {
			return nil, b, err
		}
		var value any
		value, b, err = ParseAny(b, flags)
		if err != nil {
			return nil, b, err
		}
		m[key] = value
	}
	return m, b, nil
}
func ParseMapStringString(b []byte, flags ParseFlags) (map[string]string, []byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln == -1 {
		return nil, b, nil
	}
	m := make(map[string]string, LenOrZero(ln, flags))
	for i := 0; i < ln; i++ {
		var key string
		key, b, err = ParseString(b, flags)
		if err != nil {
			return nil, b, err
		}
		var val string
		val, b, err = ParseString(b, flags)
		if err != nil {
			return nil, b, err
		}
		m[key] = val
	}
	return m, b, nil
}
func ParseMapStringBool(b []byte, flags ParseFlags) (map[string]bool, []byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln == -1 {
		return nil, b, nil
	}
	m := make(map[string]bool, LenOrZero(ln, 0))
	for i := 0; i < ln; i++ {
		var key string
		key, b, err = ParseString(b, flags)
		if err != nil {
			return nil, b, err
		}
		var val bool
		val, b, err = ParseBool(b)
		if err != nil {
			return nil, b, err
		}
		m[key] = val
	}
	return m, b, nil
}
func ParseMapLen(b []byte) (int, []byte, error) {
	ln, b, err := parseMapLen(b)
	if err != nil {
		return 0, nil, err
	}
	if len(b) < ln {
		return 0, nil, fmt.Errorf("msgp: wanted %d bytes decoding map len, got %d", ln, len(b))
	}
	return ln, b, nil
}
func parseMapLen(b []byte) (int, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c == msgpcode.Nil {
		return -1, b, nil
	}
	if c >= msgpcode.FixedMapLow && c <= msgpcode.FixedMapHigh {
		return int(c & msgpcode.FixedMapMask), b, nil
	}
	switch c {
	case msgpcode.Map16:
		n, b, err := parseUint16(b)
		return int(n), b, err
	case msgpcode.Map32:
		n, b, err := parseUint32(b)
		return int(n), b, err
	default:
		return 0, b, fmt.Errorf("msgp: unexpected code %x decoding map len", c)
	}
}
func ParseSlice(b []byte, flags ParseFlags) ([]any, []byte, error) {
	var slice []any
	b, err := Parse(b, &slice, flags)
	return slice, b, err
}
func ParseStringSlice(b []byte, flags ParseFlags) ([]string, []byte, error) {
	ln, b, err := ParseArrayLen(b)
	if err != nil {
		return nil, b, err
	}
	if ln == -1 {
		return nil, b, nil
	}
	ss := make([]string, LenOrZero(ln, flags))
	for i := 0; i < ln; i++ {
		var str string
		str, b, err = ParseString(b, flags)
		if err != nil {
			return nil, b, err
		}
		ss[i] = str
	}
	return ss, b, nil
}
func ParseArrayLen(b []byte) (int, []byte, error) {
	ln, b, err := parseArrayLen(b)
	if err != nil {
		return 0, nil, err
	}
	if len(b) < ln {
		return 0, nil, fmt.Errorf("msgp: wanted %d bytes decoding array len, got %d", ln, len(b))
	}
	return ln, b, nil
}
func parseArrayLen(b []byte) (int, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, b, err
	}
	if c == msgpcode.Nil {
		return -1, b, nil
	}
	if c >= msgpcode.FixedArrayLow && c <= msgpcode.FixedArrayHigh {
		return int(c & msgpcode.FixedArrayMask), b, nil
	}
	switch c {
	case msgpcode.Array16:
		n, b, err := parseUint16(b)
		return int(n), b, err
	case msgpcode.Array32:
		n, b, err := parseUint32(b)
		return int(n), b, err
	default:
		return 0, b, fmt.Errorf("msgp: unexpected code=%x decoding array length", c)
	}
}
func parseStructLen(b []byte) (int, bool, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, false, b, err
	}
	if c == msgpcode.Nil {
		return -1, false, b, nil
	}
	if c >= msgpcode.FixedMapLow && c <= msgpcode.FixedMapHigh {
		return int(c & msgpcode.FixedMapMask), true, b, nil
	}
	if c >= msgpcode.FixedArrayLow && c <= msgpcode.FixedArrayHigh {
		return int(c & msgpcode.FixedArrayMask), false, b, nil
	}
	switch c {
	case msgpcode.Map16:
		n, b, err := parseUint16(b)
		return int(n), true, b, err
	case msgpcode.Array16:
		n, b, err := parseUint16(b)
		return int(n), false, b, err
	case msgpcode.Map32:
		n, b, err := parseUint32(b)
		return int(n), true, b, err
	case msgpcode.Array32:
		n, b, err := parseUint32(b)
		return int(n), false, b, err
	default:
		return 0, false, b, fmt.Errorf("msgp: unexpected code=%x decoding map length", c)
	}
}
func parseExt(b []byte) (any, []byte, error) {
	extID, size, b, err := parseExtHeader(b)
	if err != nil {
		return nil, b, err
	}
	if extID == timeExtID {
		return parseTimeData(b, size)
	}
	return nil, b, fmt.Errorf("msgp: unknown ext id %d", extID)
}
func parseExtHeader(b []byte) (byte, int, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, 0, b, err
	}
	switch c {
	case msgpcode.Nil:
		return msgpcode.Nil, 0, b, nil
	case msgpcode.FixExt4:
		extID, b, err := readByte(b)
		if err != nil {
			return 0, 0, b, err
		}
		return extID, 4, b, nil
	case msgpcode.FixExt8:
		extID, b, err := readByte(b)
		if err != nil {
			return 0, 0, b, err
		}
		return extID, 8, b, nil
	case msgpcode.Ext8:
		size, b, err := readByte(b)
		if err != nil {
			return 0, 0, b, err
		}
		extID, b, err := readByte(b)
		if err != nil {
			return 0, 0, b, err
		}
		return extID, int(size), b, nil
	default:
		return 0, 0, b, fmt.Errorf("msgp: unexpected code %x decoding ext header", c)
	}
}
func readByte(b []byte) (byte, []byte, error) {
	if len(b) == 0 {
		return 0, b, io.ErrUnexpectedEOF
	}
	return b[0], b[1:], nil
}
func peekByte(b []byte) (byte, error) {
	if len(b) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	return b[0], nil
}
