package msgp

import (
	"encoding"
	"fmt"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

func Append(b []byte, x any, flags AppendFlags) ([]byte, error) {
	if x == nil {
		return AppendNil(b), nil
	}
	t := reflect.TypeOf(x)
	p := (*eface)(unsafe.Pointer(&x)).ptr
	b, err := AppendPointer(b, t, p, flags)
	runtime.KeepAlive(x)
	return b, err
}
func AppendPointer(b []byte, t reflect.Type, p unsafe.Pointer, flags AppendFlags) ([]byte, error) {
	return CodecFor(t).Append(b, p, flags)
}
func Marshal(x any, flags AppendFlags) ([]byte, error) { return Append(nil, x, flags) }

type encoder struct{ flags AppendFlags }

func (e encoder) encodeNil(b []byte, p unsafe.Pointer) ([]byte, error) { return AppendNil(b), nil }
func (e encoder) encodeBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendBool(b, *(*bool)(p)), nil
}
func (e encoder) encodeInt(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, int64(*(*int)(p))), nil
}
func (e encoder) encodeInt8(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, int64(*(*int8)(p))), nil
}
func (e encoder) encodeInt16(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, int64(*(*int16)(p))), nil
}
func (e encoder) encodeInt32(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, int64(*(*int32)(p))), nil
}
func (e encoder) encodeInt64(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, *(*int64)(p)), nil
}
func (e encoder) encodeUint(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uint)(p))), nil
}
func (e encoder) encodeUintptr(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uintptr)(p))), nil
}
func (e encoder) encodeUint8(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uint8)(p))), nil
}
func (e encoder) encodeUint16(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uint16)(p))), nil
}
func (e encoder) encodeUint32(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uint32)(p))), nil
}
func (e encoder) encodeUint64(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendUvarint(b, uint64(*(*uint)(p))), nil
}
func (e encoder) encodeFloat32(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendFloat32(b, *(*float32)(p)), nil
}
func (e encoder) encodeFloat64(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendFloat64(b, *(*float64)(p)), nil
}
func (e encoder) encodeString(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendString(b, *(*string)(p)), nil
}
func (e encoder) encodeBytes(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendBytes(b, *(*[]byte)(p)), nil
}
func (e encoder) encodeDuration(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendVarint(b, int64(*(*time.Duration)(p))), nil
}
func (e encoder) encodeTime(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendTime(b, *(*time.Time)(p)), nil
}
func (e encoder) encodeStringSlice(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendStringSlice(b, *(*[]string)(p)), nil
}
func (e encoder) encodeSlice(b []byte, p unsafe.Pointer, size uintptr, t reflect.Type, encode encodeFunc) ([]byte, error) {
	s := (*slice)(p)
	if s.data == nil && s.len == 0 && s.cap == 0 {
		return AppendNil(b), nil
	}
	return e.encodeArray(b, s.data, s.len, size, t, encode)
}
func (e encoder) encodeArray(b []byte, p unsafe.Pointer, n int, size uintptr, t reflect.Type, encode encodeFunc) (_ []byte, err error) {
	b = AppendArrayLen(b, n)
	for i := 0; i < n; i++ {
		if b, err = encode(e, b, unsafe.Pointer(uintptr(p)+(uintptr(i)*size))); err != nil {
			return b, err
		}
	}
	return b, nil
}
func (e encoder) encodeMap(b []byte, p unsafe.Pointer, t reflect.Type, encodeKey, encodeValue encodeFunc, sortKeys sortFunc) (_ []byte, err error) {
	m := reflect.NewAt(t, p).Elem()
	if m.IsNil() {
		return AppendNil(b), nil
	}
	keys := m.MapKeys()
	if sortKeys != nil && (e.flags&SortedMapKeys) != 0 {
		sortKeys(keys)
	}
	b = AppendMapLen(b, len(keys))
	for _, k := range keys {
		v := m.MapIndex(k)
		if b, err = encodeKey(e, b, (*eface)(unsafe.Pointer(&k)).ptr); err != nil {
			return b, err
		}
		if b, err = encodeValue(e, b, (*eface)(unsafe.Pointer(&v)).ptr); err != nil {
			return b, err
		}
	}
	return b, nil
}
func (e encoder) encodeMapInterfaceInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendMapInterfaceInterface(b, *(*map[any]any)(p), e.flags)
}
func (e encoder) encodeMapStringInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendMapStringInterface(b, *(*map[string]any)(p), e.flags)
}
func (e encoder) encodeMapStringRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
func (e encoder) encodeMapStringString(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendMapStringString(b, *(*map[string]string)(p), e.flags)
}
func (e encoder) encodeMapStringBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendMapStringBool(b, *(*map[string]bool)(p), e.flags)
}
func (e encoder) encodeStruct(b []byte, p unsafe.Pointer, st *structType) (_ []byte, err error) {
	if e.flags&ArrayEncodedStructs != 0 {
		var numField int
		start := len(b)
		if len(st.fieldList) < 16 {
			b = AppendMapLen8(b, numField)
		} else {
			b = AppendMapLen16(b, numField)
		}
		for id, f := range st.fieldList {
			v := unsafe.Pointer(uintptr(p) + f.offset)
			if f.empty(v) {
				continue
			}
			numField++
			b = AppendUvarint(b, uint64(id))
			if b, err = f.codec.encode(e, b, v); err != nil {
				return b, err
			}
		}
		if len(st.fieldList) < 16 {
			AppendMapLen8(b[:start], numField)
		} else {
			AppendMapLen16(b[:start], numField)
		}
		return b, nil
	}
	var numField int
	start := len(b)
	if len(st.fields) < 16 {
		b = AppendMapLen8(b, numField)
	} else {
		b = AppendMapLen16(b, numField)
	}
	for i := range st.fields {
		f := &st.fields[i]
		v := unsafe.Pointer(uintptr(p) + f.offset)
		if f.empty(v) {
			continue
		}
		numField++
		if f.id != -1 {
			b = AppendUvarint(b, uint64(f.id))
		} else {
			b = AppendString(b, f.name)
		}
		if b, err = f.codec.encode(e, b, v); err != nil {
			return b, err
		}
	}
	if len(st.fields) < 16 {
		AppendMapLen8(b[:start], numField)
	} else {
		AppendMapLen16(b[:start], numField)
	}
	return b, nil
}
func (e encoder) encodeEmbeddedStructPointer(b []byte, p unsafe.Pointer, t reflect.Type, offset uintptr, encode encodeFunc) ([]byte, error) {
	p = *(*unsafe.Pointer)(p)
	if p == nil {
		return b, nil
	}
	return encode(e, b, unsafe.Pointer(uintptr(p)+offset))
}
func (e encoder) encodePointer(b []byte, p unsafe.Pointer, t reflect.Type, encode encodeFunc) ([]byte, error) {
	if p = *(*unsafe.Pointer)(p); p != nil {
		return encode(e, b, p)
	}
	return AppendNil(b), nil
}
func (e encoder) encodeInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	return Append(b, *(*any)(p), e.flags)
}
func (e encoder) encodeMaybeNilInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	v := *(*any)(p)
	if v == nil {
		return AppendNil(b), nil
	}
	return Append(b, *(*any)(p), e.flags)
}
func (e encoder) encodeUnsupportedTypeError(b []byte, p unsafe.Pointer, t reflect.Type) ([]byte, error) {
	return b, &UnsupportedTypeError{Type: t}
}
func (e encoder) encodeRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendBytes(b, *(*RawMessage)(p)), nil
}
func (e encoder) encodeAppender(b []byte, p unsafe.Pointer, t reflect.Type, isPtr bool) ([]byte, error) {
	v := reflect.NewAt(t, p)
	if !isPtr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return AppendNil(b), nil
		}
	}
	return v.Interface().(Appender).AppendMsgpack(b, e.flags)
}
func (e encoder) encodeBinaryMarshaler(b []byte, p unsafe.Pointer, t reflect.Type, isPtr bool) ([]byte, error) {
	v := reflect.NewAt(t, p)
	if !isPtr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return AppendNil(b), nil
		}
	}
	bs, err := v.Interface().(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return b, err
	}
	return AppendBytes(b, bs), nil
}
