package chschema

import (
	"fmt"
	"reflect"
	"unsafe"
)

type pointerOffset struct {
	typ    reflect.Type
	offset uintptr
}

func offsetForIndex(typ reflect.Type, index []int) []pointerOffset {
	offset := make([]pointerOffset, len(index))
	for i, idx := range index {
		offset[i].typ = typ
		switch typ.Kind() {
		case reflect.Ptr:
			typ = typ.Elem()
		case reflect.Interface:
			panic("not implemented")
		case reflect.Struct:
		default:
			panic(fmt.Errorf("got %s, wanted %s", typ, reflect.Struct))
		}
		field := typ.Field(idx)
		offset[i].offset = field.Offset
		typ = field.Type
	}
	return offset
}
func pointerAtOffset(ptr unsafe.Pointer, offset []pointerOffset) unsafe.Pointer {
	if len(offset) == 1 {
		return unsafe.Pointer(uintptr(ptr) + offset[0].offset)
	}
	for i, info := range offset {
		if i > 0 && info.typ.Kind() == reflect.Ptr {
			ptr = *(*unsafe.Pointer)(ptr)
		}
		ptr = unsafe.Pointer(uintptr(ptr) + info.offset)
	}
	return ptr
}
func indirectPointer(typ reflect.Type, ptr unsafe.Pointer) unsafe.Pointer {
	switch typ.Kind() {
	case reflect.Ptr:
		return *(*unsafe.Pointer)(ptr)
	case reflect.Interface:
		panic("not implemented")
	case reflect.Struct:
		return ptr
	default:
		panic(fmt.Errorf("unsupported kind: %s", typ.Kind()))
	}
}
