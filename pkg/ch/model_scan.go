package ch

import (
	"fmt"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type columnarModel struct {
	columnar bool
}

func (m *columnarModel) SetColumnar(on bool) {
	m.columnar = on
}

type scanModel struct {
	columnarModel
	values []reflect.Value
}

var _ Model = (*scanModel)(nil)

func (m *scanModel) UseQueryRow() bool {
	return true
}

func (m *scanModel) ScanBlock(block *chschema.Block) error {
	if block.NumRow == 0 {
		return nil
	}
	if block.NumColumn != len(m.values) {
		return fmt.Errorf("ch: got %d columns, but Scan has %d values",
			block.NumColumn, len(m.values))
	}

	if m.columnar {
		for i, col := range block.Columns {
			v := m.values[i]
			if v.Kind() == reflect.Interface {
				v.Set(reflect.ValueOf(col.Value()))
			} else {
				v.Set(reflect.AppendSlice(v, reflect.ValueOf(col.Value())))
			}
		}
		return nil
	}

	for i, col := range block.Columns {
		if err := col.ConvertAssign(0, m.values[i]); err != nil {
			return err
		}
	}
	return nil
}
