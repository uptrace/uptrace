package chschema

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/go-clickhouse/ch/chproto"
	"github.com/uptrace/go-clickhouse/ch/internal"

	"golang.org/x/exp/constraints"
)

type Column struct {
	Name  string
	Type  string
	Field *Field

	Columnar
}

func (c *Column) String() string {
	return fmt.Sprintf("column=%s", c.Name)
}

type Columnar interface {
	Init(chType string) error
	AllocForReading(numRow int)
	ResetForWriting(numRow int)

	Type() reflect.Type
	Set(v any)
	AppendValue(v reflect.Value)
	Value() any
	Nullable(nulls UInt8Column) any
	Len() int
	Index(idx int) any
	Slice(s, e int) any
	ConvertAssign(idx int, dest reflect.Value) error

	ReadFrom(rd *chproto.Reader, numRow int) error
	WriteTo(wr *chproto.Writer) error
}

type ArrayColumnar interface {
	WriteOffset(wr *chproto.Writer, offset int) int
	WriteData(wr *chproto.Writer) error
}

//------------------------------------------------------------------------------

type ColumnOf[T any] struct {
	Column []T
}

func (c *ColumnOf[T]) Init(chType string) error {
	return nil
}

func (c *ColumnOf[T]) AllocForReading(numRow int) {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:numRow]
	} else {
		c.Column = make([]T, numRow)
	}
}

func (c *ColumnOf[T]) ResetForWriting(numRow int) {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:0]
	} else {
		c.Column = make([]T, 0, numRow)
	}
}

func (c *ColumnOf[T]) Set(v any) {
	c.Column = v.([]T)
}

func (c ColumnOf[T]) Value() any {
	return c.Column
}

func (c ColumnOf[T]) Nullable(nulls UInt8Column) any {
	nullable := make([]*T, len(c.Column))
	for i := range c.Column {
		if nulls.Column[i] == 0 {
			nullable[i] = &c.Column[i]
		}
	}
	return nullable
}

func (c ColumnOf[T]) Len() int {
	return len(c.Column)
}

func (c ColumnOf[T]) Index(idx int) any {
	return c.Column[idx]
}

func (c ColumnOf[T]) Slice(s, e int) any {
	return c.Column[s:e]
}

func (c *ColumnOf[T]) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Interface().(T))
}

func (c *ColumnOf[T]) ConvertAssign(idx int, dest reflect.Value) error {
	if scanner, ok := dest.Interface().(sql.Scanner); ok {
		return scanner.Scan(c.Column[idx])
	}
	dest.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

//------------------------------------------------------------------------------

type NumericColumnOf[T constraints.Integer | constraints.Float] struct {
	ColumnOf[T]
}

func (c NumericColumnOf[T]) ConvertAssign(idx int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		v.SetInt(int64(c.Column[idx]))
		return nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		v.SetUint(uint64(c.Column[idx]))
		return nil
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(c.Column[idx]))
		return nil
	case reflect.String:
		v.SetString(fmt.Sprint(c.Column[idx]))
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, v)
	}
}

func (c BoolColumn) ConvertAssign(idx int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(c.Column[idx])
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, v)
	}
}

func (c StringColumn) ConvertAssign(idx int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(c.Column[idx])
		return nil
	case reflect.Slice:
		if v.Type() == bytesType {
			v.SetBytes(internal.Bytes(c.Column[idx]))
			return nil
		}
	case reflect.Map:
		dec := json.NewDecoder(strings.NewReader(c.Column[idx]))
		dec.UseNumber()
		return dec.Decode(v.Addr().Interface())
	default:
		v.Set(reflect.ValueOf(c.Column[idx]))
		return nil
	}
	return fmt.Errorf("ch: can't scan %s into %s", "string", v.Type())
}

//------------------------------------------------------------------------------

type UUID [16]byte

// TODO: rework to use []byte
type UUIDColumn struct {
	ColumnOf[UUID]
}

var _ Columnar = (*UUIDColumn)(nil)

func NewUUIDColumn() Columnar {
	return new(UUIDColumn)
}

func (c UUIDColumn) Type() reflect.Type {
	return uuidType
}

func (c UUIDColumn) ConvertAssign(idx int, v reflect.Value) error {
	b := v.Slice(0, v.Len()).Bytes()
	copy(b, c.Column[idx][:])
	return nil
}

func (c *UUIDColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Convert(uuidType).Interface().(UUID))
}

func (c *UUIDColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		err := rd.UUID(c.Column[i][:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c UUIDColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		wr.UUID(c.Column[i][:])
	}
	return nil
}

//------------------------------------------------------------------------------

const ipSize = 16

var zeroIP = make([]byte, ipSize)

type IPColumn struct {
	ColumnOf[net.IP]
}

var _ Columnar = (*IPColumn)(nil)

func NewIPColumn() Columnar {
	return new(IPColumn)
}

func (c IPColumn) Type() reflect.Type {
	return ipType
}

func (c IPColumn) ConvertAssign(idx int, v reflect.Value) error {
	v.SetBytes(c.Column[idx])
	return nil
}

func (c *IPColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Bytes())
}

func (c *IPColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	mem := make([]byte, ipSize*numRow)
	var idx int
	for i := range c.Column {
		b := mem[idx : idx+ipSize]
		idx += ipSize

		if _, err := io.ReadFull(rd, b); err != nil {
			return err
		}
		c.Column[i] = b
	}

	return nil
}

func (c IPColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		b := c.Column[i]
		if len(b) == 0 {
			wr.Write(zeroIP)
			continue
		}

		if len(b) != ipSize {
			return fmt.Errorf("got %d bytes, wanted %d", len(b), ipSize)
		}
		wr.Write(b)
	}
	return nil
}

//------------------------------------------------------------------------------

type DateTime64Column struct {
	ColumnOf[time.Time]
	prec int
}

var _ Columnar = (*DateTime64Column)(nil)

func NewDateTime64Column() Columnar {
	return new(DateTime64Column)
}

func (c *DateTime64Column) Init(chType string) error {
	c.prec = parseDateTime64Prec(chType)
	return nil
}

func (c *DateTime64Column) Type() reflect.Type {
	return timeType
}

func (c *DateTime64Column) ConvertAssign(idx int, v reflect.Value) error {
	v.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

func (c *DateTime64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	mul := int64(math.Pow10(9 - c.prec))
	for i := range c.Column {
		n, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = time.Unix(0, n*mul)
	}

	return nil
}

func (c *DateTime64Column) WriteTo(wr *chproto.Writer) error {
	div := int64(math.Pow10(9 - c.prec))
	for i := range c.Column {
		wr.Int64(c.Column[i].UnixNano() / div)
	}
	return nil
}

//------------------------------------------------------------------------------

type DateColumn struct {
	DateTimeColumn
}

var _ Columnar = (*DateColumn)(nil)

func NewDateColumn() Columnar {
	return new(DateColumn)
}

func (c *DateColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Date()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c DateColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		wr.Date(c.Column[i])
	}
	return nil
}

//------------------------------------------------------------------------------

const timePrecision = int64(time.Microsecond)

type TimeColumn struct {
	DateTimeColumn
}

var _ Columnar = (*TimeColumn)(nil)

func NewTimeColumn() Columnar {
	return new(TimeColumn)
}

func (c *TimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = time.Unix(0, n*timePrecision)
	}

	return nil
}

func (c TimeColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		wr.Int64(c.Column[i].UnixNano() / timePrecision)
	}
	return nil
}

//------------------------------------------------------------------------------

type EnumColumn struct {
	StringColumn
	enum *enumInfo
}

var _ Columnar = (*EnumColumn)(nil)

func NewEnumColumn() Columnar {
	return new(EnumColumn)
}

func (c *EnumColumn) Init(chType string) error {
	c.enum = parseEnum(chType)
	return nil
}

func (c *EnumColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int8()
		if err != nil {
			return err
		}
		c.Column[i] = c.enum.Decode(int16(n))
	}

	return nil
}

func (c *EnumColumn) WriteTo(wr *chproto.Writer) error {
	for _, s := range c.Column {
		n, ok := c.enum.Encode(s)
		if !ok {
			log.Printf("unknown enum value in %s: %s", c.enum.chType, s)
		}
		wr.Int8(int8(n))
	}
	return nil
}

//------------------------------------------------------------------------------

type JSONColumn struct {
	BytesColumn
	Values []reflect.Value
}

var _ Columnar = (*JSONColumn)(nil)

func NewJSONColumn() Columnar {
	return new(JSONColumn)
}

func (c *JSONColumn) ResetForWriting(numRow int) {
	c.Values = c.Values[:0]
	c.BytesColumn.ResetForWriting(numRow)
}

func (c *JSONColumn) Len() int {
	if len(c.Values) > 0 {
		return len(c.Values)
	}
	return len(c.Column)
}

func (c *JSONColumn) ConvertAssign(idx int, v reflect.Value) error {
	dec := json.NewDecoder(bytes.NewReader(c.Column[idx]))
	dec.UseNumber()
	return dec.Decode(v.Addr().Interface())
}

func (c *JSONColumn) AppendValue(v reflect.Value) {
	if c.Values == nil {
		c.Values = make([]reflect.Value, 0, len(c.Column))
	}
	c.Values = append(c.Values, v)
}

func (c *JSONColumn) WriteTo(wr *chproto.Writer) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, v := range c.Values {
		buf.Reset()
		if err := enc.Encode(v.Interface()); err != nil {
			return err
		}
		wr.Bytes(buf.Bytes())
	}
	return nil
}

//------------------------------------------------------------------------------

type BFloat16HistColumn struct {
	ColumnOf[map[bfloat16.T]uint64]
}

var _ Columnar = (*BFloat16HistColumn)(nil)

func NewBFloat16HistColumn() Columnar {
	return new(BFloat16HistColumn)
}

func (c BFloat16HistColumn) Type() reflect.Type {
	return bfloat16MapType
}

func (c *BFloat16HistColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Uvarint()
		if err != nil {
			return err
		}

		data := make(map[bfloat16.T]uint64, n)

		for j := 0; j < int(n); j++ {
			value, err := rd.UInt16()
			if err != nil {
				return err
			}

			count, err := rd.UInt64()
			if err != nil {
				return err
			}

			data[bfloat16.T(value)] = count
		}

		c.Column[i] = data
	}

	return nil
}

func (c BFloat16HistColumn) WriteTo(wr *chproto.Writer) error {
	for _, m := range c.Column {
		wr.Uvarint(uint64(len(m)))

		for k, v := range m {
			wr.UInt16(uint16(k))
			wr.UInt64(v)
		}
	}
	return nil
}
