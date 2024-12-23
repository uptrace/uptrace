package msgp

import (
	"bytes"
	"cmp"
	"encoding"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"unsafe"
)

func CodecFor(t reflect.Type) Codec {
	cache := cacheLoad()
	c, found := cache[typeid(t)]
	if !found {
		c = createCodec(t, map[reflect.Type]*structType{}, t.Kind() == reflect.Ptr)
		if inlined(t) {
			c.encode = createInlineValueEncodeFunc(c.encode)
		}
		cacheStore(t, c, cache)
	}
	return c
}

type Codec struct {
	encode encodeFunc
	decode decodeFunc
}

func (c Codec) Append(b []byte, p unsafe.Pointer, flags AppendFlags) ([]byte, error) {
	return c.encode(encoder{flags: flags}, b, p)
}
func (c Codec) Parse(b []byte, p unsafe.Pointer, flags ParseFlags) ([]byte, error) {
	return c.decode(decoder{flags: flags}, b, p)
}

type (
	encodeFunc func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error)
	decodeFunc func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error)
	emptyFunc  func(p unsafe.Pointer) bool
	sortFunc   func([]reflect.Value)
)

var cache unsafe.Pointer

func cacheLoad() map[unsafe.Pointer]Codec {
	p := atomic.LoadPointer(&cache)
	return *(*map[unsafe.Pointer]Codec)(unsafe.Pointer(&p))
}
func cacheStore(typ reflect.Type, cod Codec, oldCodecs map[unsafe.Pointer]Codec) {
	newCodecs := make(map[unsafe.Pointer]Codec, len(oldCodecs)+1)
	newCodecs[typeid(typ)] = cod
	for t, c := range oldCodecs {
		newCodecs[t] = c
	}
	atomic.StorePointer(&cache, *(*unsafe.Pointer)(unsafe.Pointer(&newCodecs)))
}
func createCodec(t reflect.Type, seen map[reflect.Type]*structType, canAddr bool) Codec {
	switch t {
	case nilType:
		return Codec{encode: encoder.encodeNil, decode: decoder.decodeNil}
	case bytesType:
		return Codec{encode: encoder.encodeBytes, decode: decoder.decodeBytes}
	case durationType:
		return Codec{encode: encoder.encodeDuration, decode: decoder.decodeDuration}
	case timeType:
		return Codec{encode: encoder.encodeTime, decode: decoder.decodeTime}
	case interfaceType:
		return Codec{encode: encoder.encodeInterface, decode: decoder.decodeInterface}
	case rawMessageType:
		return Codec{encode: encoder.encodeRawMessage, decode: decoder.decodeRawMessage}
	case durationPtrType:
		return createPointerCodec(durationPtrType, nil)
	case timePtrType:
		return createPointerCodec(t, seen)
	case rawMessagePtrType:
		return createPointerCodec(rawMessagePtrType, nil)
	}
	var c Codec
	switch t.Kind() {
	case reflect.Bool:
		c = Codec{encode: encoder.encodeBool, decode: decoder.decodeBool}
	case reflect.Int:
		c = Codec{encode: encoder.encodeInt, decode: decoder.decodeInt}
	case reflect.Int8:
		c = Codec{encode: encoder.encodeInt8, decode: decoder.decodeInt8}
	case reflect.Int16:
		c = Codec{encode: encoder.encodeInt16, decode: decoder.decodeInt16}
	case reflect.Int32:
		c = Codec{encode: encoder.encodeInt32, decode: decoder.decodeInt32}
	case reflect.Int64:
		c = Codec{encode: encoder.encodeInt64, decode: decoder.decodeInt64}
	case reflect.Uint:
		c = Codec{encode: encoder.encodeUint, decode: decoder.decodeUint}
	case reflect.Uintptr:
		c = Codec{encode: encoder.encodeUintptr, decode: decoder.decodeUintptr}
	case reflect.Uint8:
		c = Codec{encode: encoder.encodeUint8, decode: decoder.decodeUint8}
	case reflect.Uint16:
		c = Codec{encode: encoder.encodeUint16, decode: decoder.decodeUint16}
	case reflect.Uint32:
		c = Codec{encode: encoder.encodeUint32, decode: decoder.decodeUint32}
	case reflect.Uint64:
		c = Codec{encode: encoder.encodeUint64, decode: decoder.decodeUint64}
	case reflect.Float32:
		c = Codec{encode: encoder.encodeFloat32, decode: decoder.decodeFloat32}
	case reflect.Float64:
		c = Codec{encode: encoder.encodeFloat64, decode: decoder.decodeFloat64}
	case reflect.String:
		c = Codec{encode: encoder.encodeString, decode: decoder.decodeString}
	case reflect.Interface:
		c = createInterfaceCodec()
	case reflect.Array:
		c = createArrayCodec(t, seen, canAddr)
	case reflect.Slice:
		c = createSliceCodec(t, seen)
	case reflect.Map:
		c = createMapCodec(t, seen)
	case reflect.Struct:
		c = createStructCodec(t, seen, canAddr)
	case reflect.Ptr:
		c = createPointerCodec(t, seen)
	default:
		c = createUnsupportedTypeCodec(t)
	}
	p := reflect.PtrTo(t)
	if canAddr {
		switch {
		case p.Implements(appenderType):
			c.encode = createAppenderEncodeFunc(t, true)
		case p.Implements(binaryMarshalerType):
			c.encode = createBinaryMarshalerEncodeFunc(t, true)
		}
	}
	switch {
	case t.Implements(appenderType):
		c.encode = createAppenderEncodeFunc(t, false)
	case t.Implements(binaryMarshalerType):
		c.encode = createBinaryMarshalerEncodeFunc(t, false)
	}
	switch {
	case p.Implements(parserType):
		c.decode = createParserDecodeFunc(t, true)
	case p.Implements(binaryUnmarshalerType):
		c.decode = createBinaryUnmarshalerDecodeFunc(t, true)
	}
	return c
}
func createArrayCodec(t reflect.Type, seen map[reflect.Type]*structType, canAddr bool) Codec {
	elem := t.Elem()
	c := createCodec(elem, seen, canAddr)
	size := alignedSize(elem)
	return Codec{encode: createArrayEncodeFunc(size, t, c.encode), decode: createArrayDecodeFunc(size, t, c.decode)}
}
func createArrayEncodeFunc(size uintptr, t reflect.Type, encode encodeFunc) encodeFunc {
	n := t.Len()
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeArray(b, p, n, size, t, encode)
	}
}
func createArrayDecodeFunc(size uintptr, t reflect.Type, decode decodeFunc) decodeFunc {
	n := t.Len()
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeArray(b, p, n, size, t, decode)
	}
}
func createSliceCodec(t reflect.Type, seen map[reflect.Type]*structType) Codec {
	elem := t.Elem()
	if elem == stringType {
		return Codec{encode: func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) { return e.encodeStringSlice(b, p) }, decode: func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) { return d.decodeStringSlice(b, p) }}
	}
	size := alignedSize(elem)
	if elem.Kind() == reflect.Uint8 {
		ptr := reflect.PtrTo(elem)
		var c Codec
		switch {
		case elem.Implements(appenderType):
			c.encode = createAppenderEncodeFunc(elem, false)
		case ptr.Implements(appenderType):
			c.encode = createAppenderEncodeFunc(elem, true)
		case elem.Implements(binaryMarshalerType):
			c.encode = createBinaryMarshalerEncodeFunc(elem, false)
		case ptr.Implements(binaryMarshalerType):
			c.encode = createBinaryMarshalerEncodeFunc(elem, true)
		}
		switch {
		case elem.Implements(parserType):
			c.decode = createParserDecodeFunc(elem, false)
		case ptr.Implements(parserType):
			c.decode = createParserDecodeFunc(elem, true)
		case elem.Implements(binaryUnmarshalerType):
			c.decode = createBinaryUnmarshalerDecodeFunc(elem, false)
		case ptr.Implements(binaryUnmarshalerType):
			c.decode = createBinaryUnmarshalerDecodeFunc(elem, true)
		}
		if c.encode != nil {
			c.encode = createSliceEncodeFunc(size, t, c.encode)
		} else {
			c.encode = encoder.encodeBytes
		}
		if c.decode != nil {
			c.decode = createSliceDecodeFunc(size, t, c.decode)
		} else {
			c.decode = decoder.decodeBytes
		}
		return c
	}
	c := createCodec(elem, seen, true)
	return Codec{encode: createSliceEncodeFunc(size, t, c.encode), decode: createSliceDecodeFunc(size, t, c.decode)}
}
func createSliceEncodeFunc(size uintptr, t reflect.Type, encode encodeFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeSlice(b, p, size, t, encode)
	}
}
func createSliceDecodeFunc(size uintptr, t reflect.Type, decode decodeFunc) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeSlice(b, p, size, t, decode)
	}
}
func createMapCodec(t reflect.Type, seen map[reflect.Type]*structType) Codec {
	k := t.Key()
	v := t.Elem()
	switch {
	case k == interfaceType && v == interfaceType:
		return Codec{encode: encoder.encodeMapInterfaceInterface, decode: decoder.decodeMapInterfaceInterface}
	case k == stringType && v == interfaceType:
		return Codec{encode: encoder.encodeMapStringInterface, decode: decoder.decodeMapStringInterface}
	case k == stringType && v == rawMessageType:
		return Codec{encode: encoder.encodeMapStringRawMessage, decode: decoder.decodeMapStringRawMessage}
	case k == stringType && v == stringType:
		return Codec{encode: encoder.encodeMapStringString, decode: decoder.decodeMapStringString}
	case k == stringType && v == boolType:
		return Codec{encode: encoder.encodeMapStringBool, decode: decoder.decodeMapStringBool}
	}
	kc := Codec{}
	vc := createCodec(v, seen, false)
	var sortKeys sortFunc
	if k.Implements(binaryMarshalerType) || reflect.PtrTo(k).Implements(binaryUnmarshalerType) {
		kc.encode = createBinaryMarshalerEncodeFunc(k, false)
		kc.decode = createBinaryUnmarshalerDecodeFunc(k, true)
		sortKeys = func(keys []reflect.Value) {
			slices.SortFunc(keys, func(a, b reflect.Value) int {
				k1, _ := a.Interface().(encoding.BinaryMarshaler).MarshalBinary()
				k2, _ := b.Interface().(encoding.BinaryMarshaler).MarshalBinary()
				return bytes.Compare(k1, k2)
			})
		}
	} else {
		switch k.Kind() {
		case reflect.String:
			kc.encode = encoder.encodeString
			kc.decode = decoder.decodeString
			sortKeys = func(keys []reflect.Value) {
				slices.SortFunc(keys, func(a, b reflect.Value) int { return strings.Compare(a.String(), b.String()) })
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			kc = createCodec(k, seen, false)
			sortKeys = func(keys []reflect.Value) {
				slices.SortFunc(keys, func(a, b reflect.Value) int { return cmp.Compare(a.Int(), b.Int()) })
			}
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			kc = createCodec(k, seen, false)
			sortKeys = func(keys []reflect.Value) {
				slices.SortFunc(keys, func(a, b reflect.Value) int { return cmp.Compare(a.Uint(), b.Uint()) })
			}
		default:
			return createUnsupportedTypeCodec(t)
		}
	}
	if inlined(v) {
		vc.encode = createInlineValueEncodeFunc(vc.encode)
	}
	return Codec{encode: createMapEncodeFunc(t, kc.encode, vc.encode, sortKeys), decode: createMapDecodeFunc(t, kc.decode, vc.decode)}
}
func createMapEncodeFunc(t reflect.Type, encodeKey, encodeValue encodeFunc, sortKeys sortFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeMap(b, p, t, encodeKey, encodeValue, sortKeys)
	}
}
func createMapDecodeFunc(t reflect.Type, decodeKey, decodeValue decodeFunc) decodeFunc {
	kt := t.Key()
	vt := t.Elem()
	kz := reflect.Zero(kt)
	vz := reflect.Zero(vt)
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeMap(b, p, t, kt, vt, kz, vz, decodeKey, decodeValue)
	}
}
func createStructCodec(t reflect.Type, seen map[reflect.Type]*structType, canAddr bool) Codec {
	st := newStructType(t, seen, canAddr)
	return Codec{encode: createStructEncodeFunc(st), decode: createStructDecodeFunc(st)}
}
func createStructEncodeFunc(st *structType) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) { return e.encodeStruct(b, p, st) }
}
func createStructDecodeFunc(st *structType) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) { return d.decodeStruct(b, p, st) }
}
func createEmbeddedStructPointerCodec(t reflect.Type, unexported bool, offset uintptr, field Codec) Codec {
	return Codec{encode: createEmbeddedStructPointerEncodeFunc(t, offset, field.encode), decode: createEmbeddedStructPointerDecodeFunc(t, unexported, offset, field.decode)}
}
func createEmbeddedStructPointerEncodeFunc(t reflect.Type, offset uintptr, encode encodeFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeEmbeddedStructPointer(b, p, t, offset, encode)
	}
}
func createEmbeddedStructPointerDecodeFunc(t reflect.Type, unexported bool, offset uintptr, decode decodeFunc) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeEmbeddedStructPointer(b, p, t, unexported, offset, decode)
	}
}
func createPointerCodec(t reflect.Type, seen map[reflect.Type]*structType) Codec {
	e := t.Elem()
	c := createCodec(e, seen, true)
	return Codec{encode: createPointerEncodeFunc(e, c.encode), decode: createPointerDecodeFunc(e, c.decode)}
}
func createPointerEncodeFunc(t reflect.Type, encode encodeFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) { return e.encodePointer(b, p, t, encode) }
}
func createPointerDecodeFunc(t reflect.Type, decode decodeFunc) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) { return d.decodePointer(b, p, t, decode) }
}
func createInterfaceCodec() Codec {
	return Codec{encode: createMaybeNilInterfaceencoderFunc(), decode: createMaybeNilInterfacedecoderFunc()}
}
func createMaybeNilInterfaceencoderFunc() encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) { return e.encodeMaybeNilInterface(b, p) }
}
func createMaybeNilInterfacedecoderFunc() decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) { return d.decodeMaybeNilInterface(b, p) }
}
func createUnsupportedTypeCodec(t reflect.Type) Codec {
	return Codec{encode: createUnsupportedTypeEncodeFunc(t), decode: createUnsupportedTypeDecodeFunc(t)}
}
func createUnsupportedTypeEncodeFunc(t reflect.Type) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeUnsupportedTypeError(b, p, t)
	}
}
func createUnsupportedTypeDecodeFunc(t reflect.Type) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeUnsupportedTypeError(b, p, t)
	}
}
func createAppenderEncodeFunc(t reflect.Type, isPtr bool) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) { return e.encodeAppender(b, p, t, isPtr) }
}
func createParserDecodeFunc(t reflect.Type, isPtr bool) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) { return d.decodeParser(b, p, t, isPtr) }
}
func createBinaryMarshalerEncodeFunc(t reflect.Type, isPtr bool) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return e.encodeBinaryMarshaler(b, p, t, isPtr)
	}
}
func createBinaryUnmarshalerDecodeFunc(t reflect.Type, isPtr bool) decodeFunc {
	return func(d decoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return d.decodeBinaryUnmarshaler(b, p, t, isPtr)
	}
}
func createInlineValueEncodeFunc(encode encodeFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return encode(e, b, noescape(unsafe.Pointer(&p)))
	}
}
