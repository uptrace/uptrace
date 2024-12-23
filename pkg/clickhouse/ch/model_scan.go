package ch

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

type columnarModel struct{ columnar bool }

func (m *columnarModel) SetColumnar(on bool) { m.columnar = on }

type scanModel struct {
	columnarModel
	values []any
}

var _ Model = (*scanModel)(nil)

func (m *scanModel) UseQueryRow() bool { return true }
func (m *scanModel) ScanBlock(block *Block) error {
	if block.NumRow == 0 {
		return nil
	}
	if block.NumColumn != len(m.values) {
		return fmt.Errorf("ch: got %d columns, but Scan has %d values", block.NumColumn, len(m.values))
	}
	if m.columnar {
		for i, col := range block.Columns {
			v := reflect.ValueOf(m.values[i]).Elem()
			if v.Kind() == reflect.Interface {
				v.Set(reflect.ValueOf(col.Value()))
			} else {
				v.Set(reflect.AppendSlice(v, reflect.ValueOf(col.Value())))
			}
		}
		return nil
	}
	for i, col := range block.Columns {
		x := m.values[i]
		typ := reflect.TypeOf(x).Elem()
		ptr := (*eface)(unsafe.Pointer(&x)).ptr
		if err := col.ConvertAssign(0, typ, ptr); err != nil {
			return err
		}
		runtime.KeepAlive(x)
	}
	return nil
}
