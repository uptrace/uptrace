package ch

import (
	"context"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type structTableModel struct {
	db    *DB
	table *chschema.Table
	strct reflect.Value
}

var _ TableModel = (*structTableModel)(nil)

func newStructTableModel(db *DB, table *chschema.Table) *structTableModel {
	return &structTableModel{
		db:    db,
		table: table,
	}
}

func newStructTableModelValue(db *DB, v reflect.Value) *structTableModel {
	return &structTableModel{
		db:    db,
		table: chschema.TableForType(v.Type()),
		strct: v,
	}
}

func (m *structTableModel) UseQueryRow() bool {
	return !m.table.IsColumnar()
}

func (m *structTableModel) Table() *chschema.Table {
	return m.table
}

func (m *structTableModel) AppendNamedArg(
	fmter chschema.Formatter, b []byte, name string,
) ([]byte, bool) {
	field, ok := m.table.FieldMap[name]
	if ok {
		b = field.AppendValue(fmter, b, m.strct)
		return b, true
	}
	return b, false
}

func (m *structTableModel) ScanBlock(block *chschema.Block) error {
	if block.NumRow == 0 {
		return nil
	}

	if m.table.IsColumnar() {
		return scanColumns(m.db, m.table, m.strct, block)
	}
	return scanRow(m.db, m.table, m.strct, block, 0)
}

func scanRow(
	db *DB, table *chschema.Table, strct reflect.Value, block *chschema.Block, row int,
) error {
	for _, col := range block.Columns {
		field := table.FieldMap[col.Name]
		if field == nil {
			if !db.flags.Has(discardUnknownColumnsFlag) {
				return &chschema.UnknownColumnError{
					Table:  table,
					Column: col.Name,
				}
			}
			continue
		}

		fieldValue := field.Value(strct)
		if err := col.ConvertAssign(row, fieldValue); err != nil {
			return err
		}
	}
	return nil
}

func scanColumns(db *DB, table *chschema.Table, strct reflect.Value, block *chschema.Block) error {
	for _, col := range block.Columns {
		field := table.FieldMap[col.Name]
		if field == nil {
			if !db.flags.Has(discardUnknownColumnsFlag) {
				return &chschema.UnknownColumnError{
					Table:  table,
					Column: col.Name,
				}
			}
			continue
		}

		fieldValue := field.Value(strct)
		fieldValue.Set(reflect.AppendSlice(fieldValue, reflect.ValueOf(col.Value())))
	}
	return nil
}

func (m *structTableModel) Block(fields []*chschema.Field) *chschema.Block {
	block := chschema.NewBlock(m.table, len(fields), 1)

	for _, field := range fields {
		fieldValue := field.Value(m.strct)

		col := block.Column(field.CHName, field.CHType)
		if m.table.IsColumnar() {
			col.Set(fieldValue.Interface())
		} else {
			col.AppendValue(fieldValue)
		}
	}

	return block
}

var _ AfterScanRowHook = (*structTableModel)(nil)

func (m *structTableModel) AfterScanRow(ctx context.Context) error {
	if m.table.HasAfterScanRowHook() {
		return callAfterScanRowHook(ctx, m.strct.Addr())
	}
	return nil
}
