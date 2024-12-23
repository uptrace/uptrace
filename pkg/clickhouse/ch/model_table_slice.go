package ch

import (
	"context"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"reflect"
	"unsafe"
)

type sliceTableModel struct {
	db       *DB
	table    *chschema.Table
	slice    reflect.Value
	nextElem func() reflect.Value
}

func newSliceTableModel(db *DB, slice reflect.Value, elemType reflect.Type) TableModel {
	return &sliceTableModel{db: db, table: chschema.TableForType(elemType), slice: slice, nextElem: internal.MakeSliceNextElemFunc(slice)}
}
func (m *sliceTableModel) Table() *chschema.Table { return m.table }
func (m *sliceTableModel) AppendParam(fmter chschema.Formatter, b []byte, name string) ([]byte, bool) {
	return b, false
}
func (m *sliceTableModel) ScanBlock(block *Block) error {
	for row := 0; row < block.NumRow; row++ {
		elem := m.nextElem().Addr().UnsafePointer()
		if err := scanRow(m.db, m.table, elem, block, row); err != nil {
			return err
		}
	}
	return nil
}
func (m *sliceTableModel) WriteToBlock(block *Block, fields []*chschema.Field) {
	sliceLen := m.slice.Len()
	block.initTable(m.table, len(fields), sliceLen)
	if sliceLen == 0 {
		return
	}
	for _, field := range fields {
		col := block.ColumnForField(field)
		col.Grow(sliceLen)
	}
	slicePtr := unsafe.Pointer(m.slice.Addr().Pointer())
	sh := (*reflect.SliceHeader)(slicePtr)
	ptr := unsafe.Pointer(sh.Data)
	elemType := m.slice.Type().Elem()
	elemSize := elemType.Size()
	isPtr := elemType.Kind() == reflect.Ptr
	for i := 0; i < sliceLen; i++ {
		elemPtr := unsafe.Pointer(uintptr(ptr) + uintptr(i)*elemSize)
		if isPtr {
			elemPtr = *(*unsafe.Pointer)(elemPtr)
		}
		for _, col := range block.Columns {
			col.AddPointer(col.Field.Pointer(elemPtr))
		}
	}
}

var _ AfterScanRowHook = (*sliceTableModel)(nil)

func (m *sliceTableModel) AfterScanRow(ctx context.Context) error {
	if m.table.HasAfterScanRowHook() {
		return callAfterScanRowHookSlice(ctx, m.slice)
	}
	return nil
}
