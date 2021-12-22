package ch

import (
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chschema"
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
	return &mapModel{
		m: v.Interface().(map[string]any),
	}
}

func (m *mapModel) SetColumnar(on bool) {
	m.columnar = on
}

func (m *mapModel) ScanBlock(block *chschema.Block) error {
	if m.columnar {
		for _, col := range block.Columns {
			set(m.m, col.Name, col.Value())
		}
		return nil
	}

	for _, col := range block.Columns {
		if col.Len() > 0 {
			set(m.m, col.Name, col.Index(0))
		} else {
			zero := reflect.Zero(col.Columnar.Type()).Interface()
			set(m.m, col.Name, zero)
		}
	}
	return nil
}
