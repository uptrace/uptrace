package chschema

import (
	"fmt"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

type ArrayColumnar interface {
	WriteOffset(wr *chproto.Writer, offset int) int
	WriteData(wr *chproto.Writer) error
}

type ArrayLCStringColumn struct {
	*LCStringColumn
}

func (c ArrayLCStringColumn) Type() reflect.Type {
	return stringSliceType
}

func (c *ArrayLCStringColumn) WriteTo(wr *chproto.Writer) error {
	c.writeData(wr)
	return nil
}

func (c *ArrayLCStringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}
	return c.readData(rd, numRow)
}

//------------------------------------------------------------------------------

type ArrayColumn struct {
	Column reflect.Value

	typ       reflect.Type
	elem      Columnar
	arrayElem ArrayColumnar
}

var _ Columnar = (*ArrayColumn)(nil)

func NewArrayColumn(typ reflect.Type, chType string, numRow int) Columnar {
	elemType := chArrayElemType(chType)
	if elemType == "" {
		panic(fmt.Errorf("invalid array type: %q (Go type is %s)",
			chType, typ.String()))
	}

	elem := NewColumn(typ.Elem(), elemType, 0)
	var arrayElem ArrayColumnar

	if _, ok := elem.(*LCStringColumn); ok {
		panic("not reached")
	}
	arrayElem, _ = elem.(ArrayColumnar)

	c := &ArrayColumn{
		typ:       reflect.SliceOf(typ),
		elem:      elem,
		arrayElem: arrayElem,
	}

	c.Column = reflect.MakeSlice(c.typ, 0, numRow)

	return c
}

func (c ArrayColumn) Type() reflect.Type {
	return c.typ.Elem()
}

func (c *ArrayColumn) Reset(numRow int) {
	if c.Column.Cap() >= numRow {
		c.Column = c.Column.Slice(0, 0)
	} else {
		c.Column = reflect.MakeSlice(c.typ, 0, numRow)
	}
}

func (c *ArrayColumn) Set(v any) {
	c.Column = reflect.ValueOf(v)
}

func (c *ArrayColumn) Value() any {
	return c.Column.Interface()
}

func (c *ArrayColumn) Len() int {
	return c.Column.Len()
}

func (c *ArrayColumn) Index(idx int) any {
	return c.Column.Index(idx).Interface()
}

func (c ArrayColumn) Slice(s, e int) any {
	return c.Column.Slice(s, e).Interface()
}

func (c *ArrayColumn) ConvertAssign(idx int, v reflect.Value) error {
	v.Set(c.Column.Index(idx))
	return nil
}

func (c *ArrayColumn) AppendValue(v reflect.Value) {
	c.Column = reflect.Append(c.Column, v)
}

func (c *ArrayColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if c.Column.Cap() >= numRow {
		c.Column = c.Column.Slice(0, numRow)
	} else {
		c.Column = reflect.MakeSlice(c.typ, numRow, numRow)
	}

	if numRow == 0 {
		return nil
	}

	offsets := make([]int, numRow)
	for i := 0; i < len(offsets); i++ {
		offset, err := rd.Uint64()
		if err != nil {
			return err
		}
		offsets[i] = int(offset)
	}

	if err := c.elem.ReadFrom(rd, offsets[len(offsets)-1]); err != nil {
		return err
	}

	var prev int
	for i, offset := range offsets {
		c.Column.Index(i).Set(reflect.ValueOf(c.elem.Slice(prev, offset)))
		prev = offset
	}

	return nil
}

func (c *ArrayColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)

	colLen := c.Column.Len()
	for i := 0; i < colLen; i++ {
		// TODO: add SetValue or SetPointer
		c.elem.Set(c.Column.Index(i).Interface())

		var err error
		if c.arrayElem != nil {
			err = c.arrayElem.WriteData(wr)
		} else {
			err = c.elem.WriteTo(wr)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ArrayColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	colLen := c.Column.Len()

	for i := 0; i < colLen; i++ {
		el := c.Column.Index(i)
		offset += el.Len()
		wr.Uint64(uint64(offset))
	}

	if c.arrayElem == nil {
		return offset
	}

	offset = 0
	for i := 0; i < colLen; i++ {
		el := c.Column.Index(i)
		c.elem.Set(el.Interface()) // Use SetValue or SetPointer
		offset = c.arrayElem.WriteOffset(wr, offset)
	}

	return offset
}

//------------------------------------------------------------------------------

type StringArrayColumn struct {
	Column     [][]string
	elem       Columnar
	stringElem *StringColumn
	lcElem     *LCStringColumn
}

var _ Columnar = (*StringArrayColumn)(nil)

func NewStringArrayColumn(typ reflect.Type, chType string, numRow int) Columnar {
	if _, funcType := aggFuncNameAndType(chType); funcType != "" {
		chType = funcType
	}
	elemType := chArrayElemType(chType)
	if elemType == "" {
		panic(fmt.Errorf("invalid array type: %q (Go type is %s)",
			chType, typ.String()))
	}

	columnar := NewColumn(typ.Elem(), elemType, 0)
	var stringElem *StringColumn
	var lcElem *LCStringColumn

	switch v := columnar.(type) {
	case *StringColumn:
		stringElem = v
	case *LCStringColumn:
		stringElem = &v.StringColumn
		lcElem = v
		columnar = &ArrayLCStringColumn{v}
	case *EnumColumn:
		stringElem = &v.StringColumn
	default:
		panic(fmt.Errorf("unsupported column: %T", v))
	}

	return &StringArrayColumn{
		Column:     make([][]string, 0, numRow),
		elem:       columnar,
		stringElem: stringElem,
		lcElem:     lcElem,
	}
}

func (c *StringArrayColumn) Reset(numRow int) {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:0]
	} else {
		c.Column = make([][]string, 0, numRow)
	}
}

func (c *StringArrayColumn) Type() reflect.Type {
	return stringSliceType
}

func (c *StringArrayColumn) Set(v any) {
	c.Column = v.([][]string)
}

func (c *StringArrayColumn) Value() any {
	return c.Column
}

func (c *StringArrayColumn) Len() int {
	return len(c.Column)
}

func (c *StringArrayColumn) Index(idx int) any {
	return c.Column[idx]
}

func (c StringArrayColumn) Slice(s, e int) any {
	return c.Column[s:e]
}

func (c *StringArrayColumn) ConvertAssign(idx int, v reflect.Value) error {
	v.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

func (c *StringArrayColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Interface().([]string))
}

func (c *StringArrayColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}

	if cap(c.Column) >= numRow {
		c.Column = c.Column[:numRow]
	} else {
		c.Column = make([][]string, numRow)
	}

	if c.lcElem != nil {
		if err := c.lcElem.readPrefix(rd, numRow); err != nil {
			return err
		}
	}

	offsets := make([]int, numRow)

	for i := 0; i < len(offsets); i++ {
		offset, err := rd.Uint64()
		if err != nil {
			return err
		}
		offsets[i] = int(offset)
	}

	if err := c.elem.ReadFrom(rd, offsets[len(offsets)-1]); err != nil {
		return err
	}

	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.stringElem.Column[prev:offset]
		prev = offset
	}

	return nil
}

func (c *StringArrayColumn) WriteTo(wr *chproto.Writer) error {
	if c.lcElem != nil {
		c.lcElem.writePrefix(wr)
	}

	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

var _ ArrayColumnar = (*StringArrayColumn)(nil)

func (c *StringArrayColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.Uint64(uint64(offset))
	}
	return offset
}

func (c *StringArrayColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.stringElem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}
