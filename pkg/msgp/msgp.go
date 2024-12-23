package msgp

import (
	"io"
	"time"
)

type ParseFlags uint

const (
	IgnoreLen ParseFlags = 1 << iota
	AllowUnknownFields
	ZeroCopyString
	ZeroCopyBytes
	ZeroCopy = ZeroCopyString | ZeroCopyBytes
)

type AppendFlags uint

const (
	SortedMapKeys AppendFlags = 1 << iota
	ArrayEncodedStructs
)

func IsZero(v any) bool {
	switch v := v.(type) {
	case nil:
		return true
	case int:
		return v == 0
	case string:
		return v == ""
	case IsZeroer:
		return v.IsZero()
	default:
		return false
	}
}

type Decoder struct {
	r     io.Reader
	b     []byte
	flags ParseFlags
}

func NewDecoder(r io.Reader, flags ParseFlags) *Decoder   { panic("not reached") }
func NewDecoderBytes(b []byte, flags ParseFlags) *Decoder { return &Decoder{b: b, flags: flags} }
func (d *Decoder) PeekByte() (byte, error)                { return peekByte(d.b) }
func (d *Decoder) Read(b []byte) (int, error) {
	if len(d.b) == 0 {
		return 0, io.EOF
	}
	n := copy(b, d.b)
	d.b = d.b[n:]
	return n, nil
}
func (d *Decoder) DecodeInterface() (any, error) {
	v, b, err := ParseAny(d.b, d.flags)
	d.b = b
	return v, err
}
func (d *Decoder) DecodeBool() (bool, error) {
	v, b, err := ParseBool(d.b)
	d.b = b
	return v, err
}
func (d *Decoder) DecodeUint64() (uint64, error) {
	n, b, err := ParseUint64(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeInt64() (int64, error) {
	n, b, err := ParseInt64(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeFloat32() (float32, error) {
	n, b, err := ParseFloat32(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeFloat64() (float64, error) {
	n, b, err := ParseFloat64(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeString() (string, error) {
	s, b, err := ParseString(d.b, d.flags)
	d.b = b
	return s, err
}
func (d *Decoder) DecodeStringLen() (int, error) {
	n, b, err := ParseStringLen(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeBytes() ([]byte, error) {
	bs, b, err := ParseBytes(d.b, d.flags)
	d.b = b
	return bs, err
}
func (d *Decoder) DecodeBytesLen() (int, error) {
	n, b, err := ParseBytesLen(d.b)
	d.b = b
	return n, err
}
func (d *Decoder) DecodeTime() (time.Time, error) {
	tm, b, err := ParseTime(d.b)
	d.b = b
	return tm, err
}
func (d *Decoder) DecodeMapInterfaceInterface() (map[any]any, error) {
	m, b, err := ParseMapAnyAny(d.b, d.flags)
	d.b = b
	return m, err
}
func (d *Decoder) DecodeMapStringInterface() (map[string]any, error) {
	m, b, err := ParseMapStringAny(d.b, d.flags)
	d.b = b
	return m, err
}
func (d *Decoder) DecodeMapStringString() (map[string]string, error) {
	m, b, err := ParseMapStringString(d.b, d.flags)
	d.b = b
	return m, err
}
func (d *Decoder) DecodeMapStringBool() (map[string]bool, error) {
	m, b, err := ParseMapStringBool(d.b, d.flags)
	d.b = b
	return m, err
}
func (d *Decoder) DecodeMapLen() (int, error) {
	ln, b, err := ParseMapLen(d.b)
	d.b = b
	return ln, err
}
func (d *Decoder) DecodeSlice() ([]any, error) {
	v, b, err := ParseSlice(d.b, d.flags)
	d.b = b
	return v, err
}
func (d *Decoder) DecodeArrayLen() (int, error) {
	ln, b, err := ParseArrayLen(d.b)
	d.b = b
	return ln, err
}
func LenOrZero(ln int, flags ParseFlags) int {
	if flags&IgnoreLen != 0 {
		return 0
	}
	return ln
}
