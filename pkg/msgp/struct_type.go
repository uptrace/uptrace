package msgp

import (
	"fmt"
	"github.com/uptrace/pkg/tagparser"
	"reflect"
	"strconv"
	"unsafe"
)

type structType struct {
	typ       reflect.Type
	fields    []structField
	fieldMap  map[string]*structField
	fieldList []*structField
	omitEmpty bool
}

func newStructType(t reflect.Type, seen map[reflect.Type]*structType, canAddr bool) *structType {
	if st, ok := seen[t]; ok {
		return st
	}
	st := &structType{typ: t, fields: make([]structField, 0, t.NumField()), fieldMap: make(map[string]*structField)}
	seen[t] = st
	st.processFields(t, 0, seen, canAddr)
	for i := range st.fields {
		f := &st.fields[i]
		if _, ok := st.fieldMap[f.name]; ok {
			panic(fmt.Errorf("field %q already exists", f.name))
		}
		st.fieldMap[f.name] = f
		if f.id != -1 {
			if st.fieldList == nil {
				st.fieldList = make([]*structField, len(st.fields))
			}
			if f.id >= len(st.fieldList) {
				fieldList := make([]*structField, f.id+1)
				copy(fieldList, st.fieldList)
				st.fieldList = fieldList
			}
			if other := st.fieldList[f.id]; other != nil {
				panic(fmt.Errorf("fields %s and %s have the same id=%d", f.name, other.name, f.id))
			}
			st.fieldList[f.id] = f
		}
	}
	if st.fieldList == nil {
		st.fieldList = make([]*structField, len(st.fields))
		for i := range st.fields {
			st.fieldList[i] = &st.fields[i]
		}
	}
	return st
}
func (st *structType) processFields(t reflect.Type, offset uintptr, seen map[reflect.Type]*structType, canAddr bool) {
	type embeddedField struct {
		offset     uintptr
		isPtr      bool
		unexported bool
		subtype    *structType
		subfield   *structField
	}
	names := make(map[string]struct{})
	embedded := make([]embeddedField, 0, 10)
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)
		unexported := f.PkgPath != ""
		tagstr := f.Tag.Get("msgpack")
		if tagstr == "-" {
			names[f.Name] = struct{}{}
			continue
		}
		tag := tagparser.Parse(tagstr)
		if f.Name == "msgpack" {
			if tag.Name != "" {
				panic(fmt.Errorf("got %q, but tag name must be empty", tag.Name))
			}
			st.omitEmpty = tag.HasOption("omitempty")
			continue
		}
		if unexported && !f.Anonymous {
			continue
		}
		if f.Anonymous && tagstr == "" {
			typ := f.Type
			isPtr := f.Type.Kind() == reflect.Ptr
			if isPtr {
				typ = f.Type.Elem()
			}
			if typ.Kind() != reflect.Struct {
				continue
			}
			subtype := newStructType(typ, seen, canAddr)
			for j := range subtype.fields {
				embedded = append(embedded, embeddedField{offset: offset + f.Offset, isPtr: isPtr, unexported: unexported, subtype: subtype, subfield: &subtype.fields[j]})
			}
			continue
		}
		name := f.Name
		if tag.Name != "" {
			name = tag.Name
		}
		id := -1
		if v, ok := tag.Option("id"); ok {
			n, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				panic(err)
			}
			id = int(n)
		}
		st.fields = append(st.fields, structField{typ: f.Type, zero: reflect.Zero(f.Type), id: id, name: name, offset: offset + f.Offset, tag: tag, omitempty: tag.HasOption("omitempty"), empty: emptyFuncOf(f.Type, canAddr), codec: createCodec(f.Type, seen, canAddr)})
		names[name] = struct{}{}
	}
	ambiguousNames := make(map[string]int)
	ambiguousTags := make(map[string]int)
	for name := range names {
		ambiguousNames[name]++
		ambiguousTags[name]++
	}
	for _, f := range embedded {
		ambiguousNames[f.subfield.name]++
		if !f.subfield.tag.IsZero() {
			ambiguousTags[f.subfield.name]++
		}
	}
	for _, embfield := range embedded {
		subfield := *embfield.subfield
		if ambiguousNames[subfield.name] > 1 && !(!subfield.tag.IsZero() && ambiguousTags[subfield.name] == 1) {
			continue
		}
		if embfield.isPtr {
			subfield.codec = createEmbeddedStructPointerCodec(embfield.subtype.typ, embfield.unexported, subfield.offset, subfield.codec)
			subfield.empty = createEmbeddedStructPointerEmptyFunc(embfield.subtype.typ, subfield.offset, subfield.empty)
			subfield.offset = embfield.offset
		} else {
			subfield.offset += embfield.offset
		}
		st.fields = append(st.fields, subfield)
	}
}
func (st *structType) fieldByID(id int) *structField {
	if id >= 0 && id < len(st.fieldList) {
		return st.fieldList[id]
	}
	return nil
}

type structField struct {
	typ       reflect.Type
	zero      reflect.Value
	id        int
	name      string
	offset    uintptr
	tag       tagparser.Tag
	omitempty bool
	empty     emptyFunc
	codec     Codec
}

func emptyFuncOf(t reflect.Type, canAddr bool) emptyFunc {
	switch t {
	case bytesType, rawMessageType:
		return func(p unsafe.Pointer) bool { return (*slice)(p).len == 0 }
	}
	p := reflect.PtrTo(t)
	if canAddr {
		if p.Implements(isZeroType) {
			return isZeroChecker(t, canAddr)
		}
	}
	if t.Implements(isZeroType) {
		return isZeroChecker(t, canAddr)
	}
	switch t.Kind() {
	case reflect.Array:
		if t.Len() == 0 {
			return func(unsafe.Pointer) bool { return true }
		}
		size := t.Size()
		return func(p unsafe.Pointer) bool {
			b := unsafe.Slice((*uint8)(p), size)
			for _, c := range b {
				if c != 0 {
					return false
				}
			}
			return true
		}
	case reflect.Map:
		return func(p unsafe.Pointer) bool { return *(*unsafe.Pointer)(p) == nil }
	case reflect.Slice:
		return func(p unsafe.Pointer) bool { return (*slice)(p).data == nil }
	case reflect.String:
		return func(p unsafe.Pointer) bool { return len(*(*string)(p)) == 0 }
	case reflect.Bool:
		return func(p unsafe.Pointer) bool { return !*(*bool)(p) }
	case reflect.Int, reflect.Uint:
		return func(p unsafe.Pointer) bool { return *(*uint)(p) == 0 }
	case reflect.Uintptr:
		return func(p unsafe.Pointer) bool { return *(*uintptr)(p) == 0 }
	case reflect.Int8, reflect.Uint8:
		return func(p unsafe.Pointer) bool { return *(*uint8)(p) == 0 }
	case reflect.Int16, reflect.Uint16:
		return func(p unsafe.Pointer) bool { return *(*uint16)(p) == 0 }
	case reflect.Int32, reflect.Uint32:
		return func(p unsafe.Pointer) bool { return *(*uint32)(p) == 0 }
	case reflect.Int64, reflect.Uint64:
		return func(p unsafe.Pointer) bool { return *(*uint64)(p) == 0 }
	case reflect.Float32:
		return func(p unsafe.Pointer) bool { return *(*float32)(p) == 0 }
	case reflect.Float64:
		return func(p unsafe.Pointer) bool { return *(*float64)(p) == 0 }
	case reflect.Ptr:
		return func(p unsafe.Pointer) bool { return *(*unsafe.Pointer)(p) == nil }
	case reflect.Interface:
		return func(p unsafe.Pointer) bool { return (*eface)(p).ptr == nil }
	}
	return func(unsafe.Pointer) bool { return false }
}
func createEmbeddedStructPointerEmptyFunc(t reflect.Type, offset uintptr, empty emptyFunc) emptyFunc {
	return func(p unsafe.Pointer) bool {
		p = *(*unsafe.Pointer)(p)
		if p == nil {
			return true
		}
		return empty(unsafe.Pointer(uintptr(p) + offset))
	}
}
func isZeroChecker(t reflect.Type, isPtr bool) func(p unsafe.Pointer) bool {
	return func(p unsafe.Pointer) bool {
		v := reflect.NewAt(t, p)
		if v.IsNil() {
			return true
		}
		if !isPtr {
			v = v.Elem()
		}
		return v.Interface().(IsZeroer).IsZero()
	}
}
