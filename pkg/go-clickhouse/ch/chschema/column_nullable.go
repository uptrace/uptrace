package chschema

import (
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

type NullableColumn struct {
	Nulls    UInt8Column
	Values   Columnar
	nullable reflect.Value // reflect.Slice
}

func NewNullableColumnFunc(fn NewColumnFunc) NewColumnFunc {
	return func() Columnar {
		return &NullableColumn{
			Values: fn(),
		}
	}
}

var _ Columnar = (*NullableColumn)(nil)

func (c *NullableColumn) Init(chType string) error {
	return nil
}

func (c *NullableColumn) AllocForReading(numRow int) {
	c.Nulls.AllocForReading(numRow)
	c.Values.AllocForReading(numRow)
}

func (c *NullableColumn) ResetForWriting(numRow int) {
	c.Nulls.ResetForWriting(numRow)
	c.Values.ResetForWriting(numRow)
}

func (c *NullableColumn) Type() reflect.Type {
	return reflect.PtrTo(c.Values.Type())
}

func (c *NullableColumn) Set(v any) {
	panic("not reached")
}

func (c *NullableColumn) AppendValue(v reflect.Value) {
	if v.IsNil() {
		c.Nulls.Column = append(c.Nulls.Column, 1)
		c.Values.AppendValue(reflect.New(c.Values.Type()).Elem())
	} else {
		c.Nulls.Column = append(c.Nulls.Column, 0)
		c.Values.AppendValue(v.Elem())
	}
}

func (c *NullableColumn) Value() any {
	return c.nullable.Interface()
}

func (c *NullableColumn) Nullable(nulls UInt8Column) any {
	panic("not implemented")
}

func (c *NullableColumn) Len() int {
	return c.Values.Len()
}

func (c *NullableColumn) Index(idx int) any {
	elem := c.nullable.Index(idx)
	if elem.IsNil() {
		return nil
	}
	return elem.Elem().Interface()
}

func (c *NullableColumn) Slice(s, e int) any {
	panic("not implemented")
}

func (c *NullableColumn) ConvertAssign(idx int, dest reflect.Value) error {
	if idx < len(c.Nulls.Column) && c.Nulls.Column[idx] == 1 {
		return nil
	}
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	return c.Values.ConvertAssign(idx, dest.Elem())
}

func (c *NullableColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}
	if err := c.Nulls.ReadFrom(rd, numRow); err != nil {
		return err
	}
	if err := c.Values.ReadFrom(rd, numRow); err != nil {
		return err
	}
	c.nullable = reflect.ValueOf(c.Values.Nullable(c.Nulls))
	return nil
}

func (c *NullableColumn) WriteTo(wr *chproto.Writer) error {
	if err := c.Nulls.WriteTo(wr); err != nil {
		return err
	}
	return c.Values.WriteTo(wr)
}

func isNilValue(v reflect.Value) bool {
	return false
}
