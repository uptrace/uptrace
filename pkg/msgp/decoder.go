package msgp

import (
	"encoding"
	"errors"
	"fmt"
	"github.com/uptrace/pkg/msgp/msgpcode"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

func Parse(b []byte, x any, flags ParseFlags) ([]byte, error) {
	t := reflect.TypeOf(x)
	p := (*eface)(unsafe.Pointer(&x)).ptr
	if t == nil || p == nil || t.Kind() != reflect.Ptr {
		return b, fmt.Errorf("msgp: Parse(%T)", x)
	}
	t = t.Elem()
	b, err := CodecFor(t).Parse(b, p, flags)
	runtime.KeepAlive(x)
	return b, err
}
func Unmarshal(b []byte, x any, flags ParseFlags) error {
	b, err := Parse(b, x, flags)
	if err != nil {
		return err
	}
	if len(b) > 0 {
		return fmt.Errorf("msgp: buffer has unread data: %.100x", b)
	}
	return nil
}

type decoder struct{ flags ParseFlags }

func (d decoder) decodeNil(b []byte, p unsafe.Pointer) ([]byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return b, err
	}
	if c != msgpcode.Nil {
		return b, fmt.Errorf("msgp: got %x, wanted %x", c, msgpcode.Nil)
	}
	return b, nil
}
func (d decoder) decodeBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	v, b, err := ParseBool(b)
	if err != nil {
		return b, err
	}
	*(*bool)(p) = v
	return b, nil
}
func (d decoder) decodeInt(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*int)(p) = int(n)
	return b, nil
}
func (d decoder) decodeInt8(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*int8)(p) = int8(n)
	return b, nil
}
func (d decoder) decodeInt16(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*int16)(p) = int16(n)
	return b, nil
}
func (d decoder) decodeInt32(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*int32)(p) = int32(n)
	return b, nil
}
func (d decoder) decodeInt64(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*int64)(p) = int64(n)
	return b, nil
}
func (d decoder) decodeUint(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uint)(p) = uint(n)
	return b, nil
}
func (d decoder) decodeUintptr(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uintptr)(p) = uintptr(n)
	return b, nil
}
func (d decoder) decodeUint8(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uint8)(p) = uint8(n)
	return b, nil
}
func (d decoder) decodeUint16(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uint16)(p) = uint16(n)
	return b, nil
}
func (d decoder) decodeUint32(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uint32)(p) = uint32(n)
	return b, nil
}
func (d decoder) decodeUint64(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseUint64(b)
	if err != nil {
		return b, err
	}
	*(*uint64)(p) = uint64(n)
	return b, nil
}
func (d decoder) decodeFloat32(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseFloat32(b)
	if err != nil {
		return b, err
	}
	*(*float32)(p) = n
	return b, nil
}
func (d decoder) decodeFloat64(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseFloat64(b)
	if err != nil {
		return b, err
	}
	*(*float64)(p) = n
	return b, nil
}
func (d decoder) decodeString(b []byte, p unsafe.Pointer) ([]byte, error) {
	s, b, err := ParseString(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*string)(p) = s
	return b, nil
}
func (d decoder) decodeBytes(b []byte, p unsafe.Pointer) ([]byte, error) {
	bs, b, err := ParseBytes(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*[]byte)(p) = bs
	return b, nil
}
func (d decoder) decodeDuration(b []byte, p unsafe.Pointer) ([]byte, error) {
	n, b, err := ParseInt64(b)
	if err != nil {
		return b, err
	}
	*(*time.Duration)(p) = time.Duration(n)
	return b, nil
}
func (d decoder) decodeTime(b []byte, p unsafe.Pointer) ([]byte, error) {
	tm, b, err := ParseTime(b)
	if err != nil {
		return b, err
	}
	*(*time.Time)(p) = tm
	return b, nil
}
func (d decoder) decodeArray(b []byte, p unsafe.Pointer, n int, size uintptr, t reflect.Type, decode decodeFunc) ([]byte, error) {
	ln, b, err := ParseArrayLen(b)
	if err != nil {
		return b, err
	}
	if ln != n {
		return b, fmt.Errorf("msgp: got %d array elements, wanted %d", ln, n)
	}
	for i := 0; i < ln; i++ {
		b, err = decode(d, b, unsafe.Pointer(uintptr(p)+(uintptr(i)*size)))
		if err != nil {
			return b, err
		}
	}
	return b, nil
}
func (d decoder) decodeStringSlice(b []byte, p unsafe.Pointer) ([]byte, error) {
	ss, b, err := ParseStringSlice(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*[]string)(p) = ss
	return b, nil
}

var empty struct{}

func (d decoder) decodeSlice(b []byte, p unsafe.Pointer, size uintptr, t reflect.Type, decode decodeFunc) ([]byte, error) {
	ln, b, err := ParseArrayLen(b)
	if err != nil {
		return b, err
	}
	s := (*slice)(p)
	s.len = 0
	switch ln {
	case -1:
		s.data = nil
		return b, nil
	case 0:
		s.data = unsafe.Pointer(&empty)
		return b, nil
	}
	if s.cap < ln {
		*s = extendSlice(t, s, ln)
	}
	for i := 0; i < ln; i++ {
		b, err = decode(d, b, unsafe.Pointer(uintptr(s.data)+(uintptr(s.len)*size)))
		if err != nil {
			return b, err
		}
		s.len++
	}
	return b, nil
}
func (d decoder) decodeMap(b []byte, p unsafe.Pointer, t, kt, vt reflect.Type, kz, vz reflect.Value, decodeKey, decodeValue decodeFunc) ([]byte, error) {
	ln, b, err := ParseMapLen(b)
	if err != nil {
		return b, err
	}
	if ln == -1 {
		*(*unsafe.Pointer)(p) = nil
		return b, nil
	}
	m := reflect.NewAt(t, p).Elem()
	if m.IsNil() {
		m = reflect.MakeMap(t)
		*(*unsafe.Pointer)(p) = unsafe.Pointer(m.Pointer())
	}
	if ln == 0 {
		return b, nil
	}
	k := reflect.New(kt).Elem()
	v := reflect.New(vt).Elem()
	kptr := (*eface)(unsafe.Pointer(&k)).ptr
	vptr := (*eface)(unsafe.Pointer(&v)).ptr
	for i := 0; i < ln; i++ {
		k.Set(kz)
		v.Set(vz)
		if b, err = decodeKey(d, b, kptr); err != nil {
			return b, err
		}
		if b, err = decodeValue(d, b, vptr); err != nil {
			return b, err
		}
		m.SetMapIndex(k, v)
	}
	return b, nil
}
func (d decoder) decodeMapInterfaceInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	m, b, err := ParseMapAnyAny(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*map[any]any)(p) = m
	return b, nil
}
func (d decoder) decodeMapStringInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	m, b, err := ParseMapStringAny(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*map[string]any)(p) = m
	return b, nil
}
func (d decoder) decodeMapStringRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
func (d decoder) decodeMapStringString(b []byte, p unsafe.Pointer) ([]byte, error) {
	m, b, err := ParseMapStringString(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*map[string]string)(p) = m
	return b, nil
}
func (d decoder) decodeMapStringBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	m, b, err := ParseMapStringBool(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*map[string]bool)(p) = m
	return b, nil
}
func (d decoder) decodeStruct(b []byte, p unsafe.Pointer, st *structType) ([]byte, error) {
	n, isMap, b, err := parseStructLen(b)
	if err != nil {
		return b, err
	}
	if n == -1 {
		return b, nil
	}
	if !isMap {
		if n != len(st.fields) {
			return b, fmt.Errorf("msgp: got %d fields decoding struct, wanted %d", n, len(st.fields))
		}
		for i := range st.fields {
			f := &st.fields[i]
			if b, err = f.codec.decode(d, b, unsafe.Pointer(uintptr(p)+f.offset)); err != nil {
				return b, err
			}
		}
		return b, nil
	}
	for i := 0; i < n; i++ {
		var field *structField
		field, b, err = d.decodeStructField(b, st)
		if err != nil {
			return b, err
		}
		if field == nil {
			b, err = Skip(b)
			if err != nil {
				return b, err
			}
			continue
		}
		b, err = field.codec.decode(d, b, unsafe.Pointer(uintptr(p)+field.offset))
		if err != nil {
			return b, fmt.Errorf("msgp: decoding %q failed: %w", field.name, err)
		}
	}
	return b, nil
}
func (d decoder) decodeStructField(b []byte, st *structType) (*structField, []byte, error) {
	if fieldID, b, err := d.decodeStructFieldID(b); err == nil {
		field := st.fieldByID(fieldID)
		if field == nil && d.flags&AllowUnknownFields == 0 {
			return nil, nil, fmt.Errorf("msgp: unknown struct field id: %d", fieldID)
		}
		return field, b, nil
	} else if err != errInvalidFieldID {
		return nil, nil, errInvalidFieldID
	}
	fieldName, b, err := ParseBytes(b, ZeroCopyBytes)
	if err != nil {
		return nil, nil, err
	}
	field, ok := st.fieldMap[string(fieldName)]
	if !ok && d.flags&AllowUnknownFields == 0 {
		return nil, nil, fmt.Errorf("msgp: unknown struct field name: %q", string(fieldName))
	}
	return field, b, nil
}

var errInvalidFieldID = errors.New("invalid struct field id")

func (d decoder) decodeStructFieldID(b []byte) (int, []byte, error) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, nil, err
	}
	if msgpcode.IsFixedNum(c) {
		return int(int8(c)), b, nil
	}
	switch c {
	case msgpcode.Uint8:
		n, b, err := readByte(b)
		if err != nil {
			return 0, nil, err
		}
		return int(n), b, nil
	case msgpcode.Uint16:
		n, b, err := parseUint16(b)
		if err != nil {
			return 0, nil, err
		}
		return int(n), b, nil
	default:
		return 0, nil, errInvalidFieldID
	}
}
func (d decoder) decodeEmbeddedStructPointer(b []byte, p unsafe.Pointer, t reflect.Type, unexported bool, offset uintptr, decode decodeFunc) ([]byte, error) {
	v := *(*unsafe.Pointer)(p)
	if v == nil {
		if unexported {
			return nil, fmt.Errorf("msgp: cannot set embedded pointer to unexported struct: %s", t)
		}
		v = unsafe.Pointer(reflect.New(t).Pointer())
		*(*unsafe.Pointer)(p) = v
	}
	return decode(d, b, unsafe.Pointer(uintptr(v)+offset))
}
func (d decoder) decodePointer(b []byte, p unsafe.Pointer, t reflect.Type, decode decodeFunc) ([]byte, error) {
	if hasNilCode(b) {
		pp := *(*unsafe.Pointer)(p)
		if pp != nil && t.Kind() == reflect.Ptr {
			return decode(d, b, pp)
		}
		*(*unsafe.Pointer)(p) = nil
		return b[1:], nil
	}
	v := *(*unsafe.Pointer)(p)
	if v == nil {
		v = unsafe.Pointer(reflect.New(t).Pointer())
		*(*unsafe.Pointer)(p) = v
	}
	return decode(d, b, v)
}
func (d decoder) decodeInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	v, b, err := ParseAny(b, d.flags)
	if err != nil {
		return b, err
	}
	*(*any)(p) = v
	return b, nil
}
func (d decoder) decodeMaybeNilInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	return d.decodeInterface(b, p)
}
func (d decoder) decodeUnsupportedTypeError(b []byte, p unsafe.Pointer, t reflect.Type) ([]byte, error) {
	return b, &UnsupportedTypeError{Type: t}
}
func (d decoder) decodeRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	return d.decodeBytes(b, p)
}
func (d decoder) decodeParser(b []byte, p unsafe.Pointer, t reflect.Type, isPtr bool) ([]byte, error) {
	v := reflect.NewAt(t, p)
	if !isPtr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.IsNil() {
		v.Set(reflect.New(t))
	}
	return v.Interface().(Parser).ParseMsgpack(b, d.flags)
}
func (d decoder) decodeBinaryUnmarshaler(b []byte, p unsafe.Pointer, t reflect.Type, isPtr bool) ([]byte, error) {
	bs, b, err := ParseBytes(b, d.flags)
	if err != nil {
		return b, err
	}
	v := reflect.NewAt(t, p)
	if !isPtr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.IsNil() {
		v.Set(reflect.New(t))
	}
	return b, v.Interface().(encoding.BinaryUnmarshaler).UnmarshalBinary(bs)
}
func hasNilCode(b []byte) bool { return len(b) > 0 && b[0] == msgpcode.Nil }
