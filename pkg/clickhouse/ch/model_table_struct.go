package ch

import (
	"context"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"reflect"
	"unsafe"
)

type structTableModel struct {
	db       *DB
	table    *chschema.Table
	strct    reflect.Value
	columnar bool
}

func newStructTableModel(db *DB, table *chschema.Table) *structTableModel {
	return &structTableModel{db: db, table: table}
}
func newStructTableModelValue(db *DB, v reflect.Value) *structTableModel {
	return &structTableModel{db: db, table: chschema.TableForType(v.Type()), strct: v}
}
func (m *structTableModel) Table() *chschema.Table { return m.table }
func (m *structTableModel) AppendNamedArg(fmter chschema.Formatter, b []byte, name string) ([]byte, bool) {
	field, ok := m.table.FieldMap[name]
	if ok {
		b = field.AppendValue(fmter, b, m.strct)
		return b, true
	}
	return b, false
}
func (m *structTableModel) SetColumnar(on bool) { m.columnar = on }
func (m *structTableModel) UseQueryRow() bool   { return !m.isColumnar() }
func (m *structTableModel) isColumnar() bool    { return m.columnar || m.table.IsColumnar() }
func (m *structTableModel) ScanBlock(block *Block) error {
	if block.NumRow == 0 {
		return nil
	}
	if m.isColumnar() {
		return scanColumns(m.db, m.table, m.strct, block)
	}
	return scanRow(m.db, m.table, m.strct.Addr().UnsafePointer(), block, 0)
}
func scanRow(db *DB, table *chschema.Table, strct unsafe.Pointer, block *Block, row int) error {
	for _, col := range block.Columns {
		field := table.FieldMap[col.Name]
		if field == nil {
			if !db.flags.Has(discardUnknownColumnsFlag) {
				return &chschema.UnknownColumnError{Table: table, Column: col.Name}
			}
			continue
		}
		if err := col.ConvertAssign(row, field.Type, field.Pointer(strct)); err != nil {
			return err
		}
	}
	return nil
}
func scanColumns(db *DB, table *chschema.Table, strct reflect.Value, block *Block) error {
	for _, col := range block.Columns {
		field := table.FieldMap[col.Name]
		if field == nil {
			if !db.flags.Has(discardUnknownColumnsFlag) {
				return &chschema.UnknownColumnError{Table: table, Column: col.Name}
			}
			continue
		}
		fieldValue := field.Value(strct)
		fieldValue.Set(reflect.AppendSlice(fieldValue, reflect.ValueOf(col.Value())))
	}
	return nil
}
func (m *structTableModel) WriteToBlock(block *Block, fields []*chschema.Field) {
	block.initTable(m.table, len(fields), 1)
	structPtr := unsafe.Pointer(m.strct.Addr().Pointer())
	for _, field := range fields {
		col := block.ColumnForField(field)
		col.Grow(1)
		if m.isColumnar() {
			col.SetValue(col.Field.Pointer(structPtr))
		} else {
			col.AddPointer(col.Field.Pointer(structPtr))
		}
	}
}

var _ AfterScanRowHook = (*structTableModel)(nil)

func (m *structTableModel) AfterScanRow(ctx context.Context) error {
	if m.table.HasAfterScanRowHook() {
		return callAfterScanRowHook(ctx, m.strct.Addr())
	}
	return nil
}
