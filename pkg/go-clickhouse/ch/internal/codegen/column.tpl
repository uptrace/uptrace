package chschema

import (
	"time"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

{{- range . }}

{{ if not .IsCustom }}

type {{ .Name }}Column struct {
	{{ if gt .Size 0 }}Numeric{{ end }}ColumnOf[{{ .GoType }}]
}

var _ Columnar = (*{{ .Name }}Column)(nil)

func New{{ .Name }}Column() Columnar {
	return new({{ .Name }}Column)
}

var _{{ .Name }}Type = reflect.TypeOf((*{{ .GoType }})(nil)).Elem()

func (c *{{ .Name }}Column) Type() reflect.Type {
   return _{{ .Name }}Type
}

{{ if .GoReflect }}
func (c *{{ .Name }}Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, {{ .GoType }}(v.{{ .GoReflect }}()))
}
{{ end }}

{{ if eq .Size 0 }}

func (c *{{ .Name }}Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.{{ .Name }}()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *{{ .Name }}Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.{{ .Name }}(n)
	}
	return nil
}

{{ end }}
{{ end }}

//------------------------------------------------------------------------------

type Array{{ .Name }}Column struct {
	ColumnOf[[]{{ .GoType }}]
	elem {{ .Name }}Column
}

var (
	_ Columnar		 = (*Array{{ .Name }}Column)(nil)
	_ ArrayColumnar = (*Array{{ .Name }}Column)(nil)
)

func NewArray{{ .Name }}Column() Columnar {
	return new(Array{{ .Name }}Column)
}

func (c *Array{{ .Name }}Column) Init(chType string) error {
    return c.elem.Init(chArrayElemType(chType))
}

func (c *Array{{ .Name }}Column) Type() reflect.Type {
   return reflect.TypeOf((*[]{{ .GoType }})(nil)).Elem()
}

func (c *Array{{ .Name }}Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	offsets, err := c.readOffsets(rd, numRow)
	if err != nil {
		 return err
	}

	if err := c.elem.ReadFrom(rd, offsets[len(offsets)-1]); err != nil {
		return err
	}

	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset]
		prev = offset
	}

	return nil
}

func (c *Array{{ .Name }}Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
	offsets := make([]int, numRow)
	for i := range offsets {
		offset, err := rd.UInt64()
		if err != nil {
			return nil, err
		}
		offsets[i] = int(offset)
	}
	return offsets, nil
}

func (c *Array{{ .Name }}Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *Array{{ .Name }}Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *Array{{ .Name }}Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArray{{ .Name }}Column struct {
	ColumnOf[[][]{{ .GoType }}]
	elem Array{{ .Name }}Column
}

var (
	_ Columnar		 = (*ArrayArray{{ .Name }}Column)(nil)
	_ ArrayColumnar = (*ArrayArray{{ .Name }}Column)(nil)
)

func NewArrayArray{{ .Name }}Column() Columnar {
	return new(ArrayArray{{ .Name }}Column)
}

func (c *ArrayArray{{ .Name }}Column) Init(chType string) error {
    return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArray{{ .Name }}Column) Type() reflect.Type {
   return reflect.TypeOf((*[][]{{ .GoType }})(nil)).Elem()
}

func (c *ArrayArray{{ .Name }}Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	offsets, err := c.readOffsets(rd, numRow)
	if err != nil {
		 return err
	}

	if err := c.elem.ReadFrom(rd, offsets[len(offsets)-1]); err != nil {
		return err
	}

	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset]
		prev = offset
	}

	return nil
}

func (c *ArrayArray{{ .Name }}Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
	offsets := make([]int, numRow)
	for i := range offsets {
		offset, err := rd.UInt64()
		if err != nil {
			return nil, err
		}
		offsets[i] = int(offset)
	}
	return offsets, nil
}

func (c *ArrayArray{{ .Name }}Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArray{{ .Name }}Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	offset = 0
	for _, elem := range c.Column {
		c.elem.Column = elem
		offset = c.elem.WriteOffset(wr, offset)
	}

	return offset
}

func (c *ArrayArray{{ .Name }}Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

{{- end }}
