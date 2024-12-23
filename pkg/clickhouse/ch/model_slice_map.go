package ch

import (
	"reflect"
	"strings"
)

type sliceMapModel struct {
	v     reflect.Value
	slice []map[string]any
}

var _ Model = (*sliceMapModel)(nil)

func newSliceMapModel(v reflect.Value) *sliceMapModel {
	return &sliceMapModel{v: v, slice: v.Interface().([]map[string]any)}
}
func (m *sliceMapModel) ScanBlock(block *Block) error {
	for i := 0; i < block.NumRow; i++ {
		row := make(map[string]any, block.NumColumn)
		for _, col := range block.Columns {
			dest := getDestMap(row, col.Name)
			dest[col.Name] = col.Index(i)
		}
		m.slice = append(m.slice, row)
	}
	m.v.Set(reflect.ValueOf(m.slice))
	return nil
}
func getDestMap(m map[string]any, key string) map[string]any {
	const sep = "__"
	for {
		idx := strings.Index(key, sep)
		if idx == -1 {
			break
		}
		subKey := key[:idx]
		key = key[idx+len(sep):]
		if subMap, ok := m[subKey].(map[string]any); ok {
			m = subMap
			continue
		}
		subMap := make(map[string]any)
		m[subKey] = subMap
		m = subMap
	}
	return m
}
