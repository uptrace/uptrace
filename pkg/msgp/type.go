package msgp

import (
	"encoding"
	"reflect"
	"time"
	"unsafe"
)

var (
	nilType                   = reflect.TypeOf(nil)
	boolType                  = reflect.TypeOf(false)
	intType                   = reflect.TypeOf(int(0))
	int8Type                  = reflect.TypeOf(int8(0))
	int16Type                 = reflect.TypeOf(int16(0))
	int32Type                 = reflect.TypeOf(int32(0))
	int64Type                 = reflect.TypeOf(int64(0))
	uintType                  = reflect.TypeOf(uint(0))
	uint8Type                 = reflect.TypeOf(uint8(0))
	uint16Type                = reflect.TypeOf(uint16(0))
	uint32Type                = reflect.TypeOf(uint32(0))
	uint64Type                = reflect.TypeOf(uint64(0))
	uintptrType               = reflect.TypeOf(uintptr(0))
	float32Type               = reflect.TypeOf(float32(0))
	float64Type               = reflect.TypeOf(float64(0))
	stringType                = reflect.TypeOf("")
	stringsType               = reflect.TypeOf([]string(nil))
	bytesType                 = reflect.TypeOf(([]byte)(nil))
	durationType              = reflect.TypeOf(time.Duration(0))
	timeType                  = reflect.TypeOf(time.Time{})
	rawMessageType            = reflect.TypeOf(RawMessage(nil))
	durationPtrType           = reflect.PtrTo(durationType)
	timePtrType               = reflect.PtrTo(timeType)
	rawMessagePtrType         = reflect.PtrTo(rawMessageType)
	sliceInterfaceType        = reflect.TypeOf(([]any)(nil))
	sliceStringType           = reflect.TypeOf(([]any)(nil))
	mapInterfaceInterfaceType = reflect.TypeOf((map[any]any)(nil))
	mapStringInterfaceType    = reflect.TypeOf((map[string]any)(nil))
	mapStringRawMessageType   = reflect.TypeOf((map[string]RawMessage)(nil))
	mapStringStringType       = reflect.TypeOf((map[string]string)(nil))
	mapStringBoolType         = reflect.TypeOf((map[string]bool)(nil))
	interfaceType             = reflect.TypeOf((*any)(nil)).Elem()
	binaryMarshalerType       = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnmarshalerType     = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	appenderType              = reflect.TypeOf((*Appender)(nil)).Elem()
	parserType                = reflect.TypeOf((*Parser)(nil)).Elem()
	isZeroType                = reflect.TypeOf((*IsZeroer)(nil)).Elem()
)

type Sizer interface{ MsgpackSize() int }
type Appender interface {
	AppendMsgpack(b []byte, flags AppendFlags) ([]byte, error)
}
type Parser interface {
	ParseMsgpack(b []byte, flags ParseFlags) ([]byte, error)
}
type Marshaler interface{ MarshalMsgpack() ([]byte, error) }
type Unmarshaler interface{ UnmarshalMsgpack(b []byte) error }
type IsZeroer interface{ IsZero() bool }

func typeid(t reflect.Type) unsafe.Pointer { return (*eface)(unsafe.Pointer(&t)).ptr }

type eface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func unpackEface(x interface{}) *eface { return (*eface)(unsafe.Pointer(&x)) }
func packEface(t reflect.Type, ptr unsafe.Pointer) interface{} {
	var x interface{}
	v := (*eface)(unsafe.Pointer(&x))
	v.typ = (*eface)(unsafe.Pointer(&t)).ptr
	v.ptr = ptr
	return x
}

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}
type RawMessage []byte

func alignedSize(t reflect.Type) uintptr {
	a := t.Align()
	s := t.Size()
	return align(uintptr(a), uintptr(s))
}
func align(align, size uintptr) uintptr {
	if align != 0 && (size%align) != 0 {
		size = ((size / align) + 1) * align
	}
	return size
}
func inlined(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && inlined(t.Field(0).Type)
	default:
		return false
	}
}

//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
