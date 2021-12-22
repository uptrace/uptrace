package urlstruct

import (
	"context"
	"net/url"
	"reflect"

	"github.com/codemodus/kace"
	"github.com/vmihailenco/tagparser"
)

type ValuesUnmarshaler interface {
	UnmarshalValues(ctx context.Context, values url.Values) error
}

type ParamUnmarshaler interface {
	UnmarshalParam(ctx context.Context, name string, values []string) error
}

//------------------------------------------------------------------------------

type StructInfo struct {
	fields   []*Field
	fieldMap map[string]*Field

	isValuesUnmarshaler bool
	isParamUnmarshaler  bool
}

func newStructInfo(typ reflect.Type) *StructInfo {
	sinfo := &StructInfo{
		fields:   make([]*Field, 0, typ.NumField()),
		fieldMap: make(map[string]*Field),

		isValuesUnmarshaler: isValuesUnmarshaler(reflect.PtrTo(typ)),
		isParamUnmarshaler:  isParamUnmarshaler(reflect.PtrTo(typ)),
	}
	sinfo.addFields(typ, nil, "")
	return sinfo
}

func (sinfo *StructInfo) Field(name string) *Field {
	return sinfo.fieldMap[name]
}

func (sinfo *StructInfo) addFields(typ reflect.Type, baseIndex []int, baseName string) {
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}

		if !sf.Anonymous {
			sinfo.addField(sf, baseIndex, baseName)
		}

		tag := sf.Tag.Get("urlstruct")
		if tag == "-" {
			continue
		}

		sfType := sf.Type
		if sfType.Kind() == reflect.Ptr {
			sfType = sfType.Elem()
		}
		if sfType.Kind() != reflect.Struct {
			continue
		}

		sinfo.addFields(sfType, joinIndex(baseIndex, sf.Index), baseName)
	}
}

func (sinfo *StructInfo) addField(sf reflect.StructField, baseIndex []int, baseName string) {
	tag := tagparser.Parse(sf.Tag.Get("urlstruct"))
	if tag.Name == "-" {
		return
	}

	index := joinIndex(baseIndex, sf.Index)

	name := tag.Name
	if name == "" {
		name = kace.Snake(sf.Name)
	}
	name = baseName + name

	if sf.Type.Kind() == reflect.Struct {
		prefix := name + "."
		if s, ok := tag.Options["prefix"]; ok {
			prefix = s
		}
		sinfo.addFields(sf.Type, index, prefix)
	}

	f := &Field{
		Type:  sf.Type,
		Name:  name,
		Index: index,
		Tag:   tag,
	}
	f.init()

	if f.scanValue != nil {
		sinfo.fields = append(sinfo.fields, f)
		sinfo.fieldMap[f.Name] = f
	}
}

func joinIndex(base, idx []int) []int {
	if len(base) == 0 {
		return idx
	}
	r := make([]int, 0, len(base)+len(idx))
	r = append(r, base...)
	r = append(r, idx...)
	return r
}

//------------------------------------------------------------------------------

var (
	valuesUnmarshalerType = reflect.TypeOf((*ValuesUnmarshaler)(nil)).Elem()
	paramUnmarshalerType  = reflect.TypeOf((*ParamUnmarshaler)(nil)).Elem()
)

func isValuesUnmarshaler(typ reflect.Type) bool {
	return typ.Implements(valuesUnmarshalerType)
}

func isParamUnmarshaler(typ reflect.Type) bool {
	return typ.Implements(paramUnmarshalerType)
}
