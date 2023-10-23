package ch

import (
	"context"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type sliceTableModel struct {
	db       *DB
	table    *chschema.Table
	slice    reflect.Value
	nextElem func() reflect.Value
}

var _ TableModel = (*sliceTableModel)(nil)

func newSliceTableModel(db *DB, slice reflect.Value, elemType reflect.Type) TableModel {
	return &sliceTableModel{
		db:       db,
		table:    chschema.TableForType(elemType),
		slice:    slice,
		nextElem: internal.MakeSliceNextElemFunc(slice),
	}
}

func (m *sliceTableModel) Table() *chschema.Table {
	return m.table
}

func (m *sliceTableModel) AppendParam(
	fmter chschema.Formatter, b []byte, name string,
) ([]byte, bool) {
	return b, false
}

func (m *sliceTableModel) ScanBlock(block *chschema.Block) error {
	for row := 0; row < block.NumRow; row++ {
		elem := m.nextElem()
		if err := scanRow(m.db, m.table, elem, block, row); err != nil {
			return err
		}
	}
	return nil
}

func (m *sliceTableModel) Block(fields []*chschema.Field) *chschema.Block {
	sliceLen := m.slice.Len()
	block := chschema.NewBlock(m.table, len(fields), sliceLen)

	if sliceLen == 0 {
		return block
	}

	for _, field := range fields {
		_ = block.ColumnForField(field)
	}

	for i := 0; i < sliceLen; i++ {
		elem := indirect(m.slice.Index(i))
		for _, col := range block.Columns {
			col.AppendValue(col.Field.Value(elem))
		}
	}

	return block
}

var _ AfterScanRowHook = (*sliceTableModel)(nil)

func (m *sliceTableModel) AfterScanRow(ctx context.Context) error {
	if m.table.HasAfterScanRowHook() {
		return callAfterScanRowHookSlice(ctx, m.slice)
	}
	return nil
}
