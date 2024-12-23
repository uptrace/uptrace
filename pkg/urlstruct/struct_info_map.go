package urlstruct

import (
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"reflect"
)

var globalMap = newStructInfoMap()

func DescribeStruct(typ reflect.Type) *StructInfo { return globalMap.DescribeStruct(typ) }

type structInfoMap struct {
	m *xsync.MapOf[reflect.Type, *StructInfo]
}

func newStructInfoMap() *structInfoMap {
	return &structInfoMap{m: xsync.NewMapOf[reflect.Type, *StructInfo]()}
}
func (m *structInfoMap) DescribeStruct(typ reflect.Type) *StructInfo {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("got %s, wanted %s", typ.Kind(), reflect.Struct))
	}
	if v, ok := m.m.Load(typ); ok {
		return v
	}
	sinfo := newStructInfo(typ)
	if v, loaded := m.m.LoadOrStore(typ, sinfo); loaded {
		return v
	}
	return sinfo
}
