package ch

import (
	"reflect"
)

type mapModel struct {
	m        map[string]any
	columnar bool
}

var _ Model = (*mapModel)(nil)

func newMapModel(v reflect.Value) *mapModel {
	if v.IsNil() {
		v.Set(reflect.MakeMap(mapType))
	}
	return &mapModel{m: v.Interface().(map[string]any)}
}
func (m *mapModel) SetColumnar(on bool) { m.columnar = on }
func (m *mapModel) ScanBlock(block *Block) error {
	if m.columnar {
		for _, col := range block.Columns {
			dest := getDestMap(m.m, col.Name)
			if slice, ok := dest[col.Name]; ok {
				dest[col.Name] = appendSlice(slice, col.Value())
			} else {
				dest[col.Name] = col.Slice(0, col.Len())
			}
		}
		return nil
	}
	for _, col := range block.Columns {
		dest := getDestMap(m.m, col.Name)
		if col.Len() > 0 {
			dest[col.Name] = col.Index(0)
		} else {
			zero := reflect.Zero(col.Columnar.Type()).Interface()
			dest[col.Name] = zero
		}
	}
	return nil
}
func appendSlice(destAny, srcAny any) any {
	dest := reflect.ValueOf(destAny)
	src := reflect.ValueOf(srcAny)
	return reflect.AppendSlice(dest, src).Interface()
}
