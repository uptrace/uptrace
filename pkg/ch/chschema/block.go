package chschema

import (
	"fmt"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

type Block struct {
	Table *Table

	NumColumn int // read-only
	NumRow    int // read-only

	Columns   []*Column
	columnMap map[string]*Column
}

func NewBlock(table *Table, numCol, numRow int) *Block {
	return &Block{
		Table:     table,
		NumColumn: numCol,
		NumRow:    numRow,
	}
}

func (b *Block) ColumnForField(field *Field) *Column {
	col := b.Column(field.CHName, field.CHType)
	col.Field = field
	return col
}

func (b *Block) Column(colName, colType string) *Column {
	if col, ok := b.columnMap[colName]; ok {
		return col
	}

	var col *Column
	if b.Table != nil {
		col = b.Table.NewColumn(colName, colType, b.NumRow)
	}
	if col == nil {
		col = &Column{
			Name:     colName,
			Type:     colType,
			Columnar: NewColumnFromCHType(colType, b.NumRow),
		}
	}

	if b.Columns == nil && b.columnMap == nil {
		b.Columns = make([]*Column, 0, b.NumColumn)
		b.columnMap = make(map[string]*Column, b.NumColumn)
	}
	b.Columns = append(b.Columns, col)
	b.columnMap[colName] = col

	return col
}

func (b *Block) WriteTo(wr *chproto.Writer) error {
	// Can't use b.NumRow for column oriented struct.
	var numRow int
	if len(b.Columns) > 0 {
		numRow = b.Columns[0].Len()
	}

	wr.Uvarint(uint64(len(b.Columns)))
	wr.Uvarint(uint64(numRow))

	for _, col := range b.Columns {
		if col.Len() != numRow {
			err := fmt.Errorf("%s does not have expected number of rows: got %d, wanted %d",
				col, col.Len(), numRow)
			panic(err)
		}
		wr.String(col.Name)
		wr.String(col.Type)
		if err := col.WriteTo(wr); err != nil {
			return err
		}
	}

	return nil
}
