package chschema

import (
	"bytes"
	"constraints"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"reflect"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chproto"
	"github.com/uptrace/go-clickhouse/ch/internal"
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
	ReadFrom(rd *chproto.Reader, numRow int) error
	WriteTo(wr *chproto.Writer) error

	Type() reflect.Type
	Set(v any)
	AppendValue(v reflect.Value)
	Value() any
	Len() int
	Index(idx int) any
	Slice(s, e int) any
	ConvertAssign(idx int, dest reflect.Value) error
}

func NewColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return ColumnFactory(typ, chType)(typ, chType, numRow)
}

func NewColumnFromCHType(chType string, numRow int) Columnar {
	typ := goType(chType)
	return NewColumn(typ, chType, numRow)
}

//------------------------------------------------------------------------------

type ColumnOf[T any] struct {
	Column []T
}

func NewColumnOf[T any](numRow int) ColumnOf[T] {
	return ColumnOf[T]{
		Column: make([]T, 0, numRow),
	}
}

func (c *ColumnOf[T]) Alloc(numRow int) {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:numRow]
	} else {
		c.Column = make([]T, numRow)
	}
}

func (c *ColumnOf[T]) Reset(numRow int) {
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
	dest.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

//------------------------------------------------------------------------------

type NumericColumnOf[T constraints.Integer | constraints.Float] struct {
	ColumnOf[T]
}

func NewNumericColumnOf[T constraints.Integer | constraints.Float](numRow int) NumericColumnOf[T] {
	col := NumericColumnOf[T]{}
	col.Column = make([]T, 0, numRow)
	return col
}

func (c NumericColumnOf[T]) ConvertAssign(idx int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		v.SetInt(int64(c.Column[idx]))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		v.SetUint(uint64(c.Column[idx]))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(c.Column[idx]))
	default:
		v.Set(reflect.ValueOf(c.Column[idx]))
	}
	return nil
}

//------------------------------------------------------------------------------

type BoolColumn struct {
	ColumnOf[bool]
}

var _ Columnar = (*BoolColumn)(nil)

func NewBoolColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &BoolColumn{
		ColumnOf: NewColumnOf[bool](numRow),
	}
}

func (c BoolColumn) Type() reflect.Type {
	return boolType
}

func (c BoolColumn) ConvertAssign(idx int, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(c.Column[idx])
	default:
		v.Set(reflect.ValueOf(c.Column[idx]))
	}
	return nil
}

func (c *BoolColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Bool())
}

func (c *BoolColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		flag, err := rd.Bool()
		if err != nil {
			return err
		}
		c.Column[i] = flag
	}

	return nil
}

func (c BoolColumn) WriteTo(wr *chproto.Writer) error {
	for _, flag := range c.Column {
		wr.Bool(flag)
	}
	return nil
}

//------------------------------------------------------------------------------

type Int8Column struct {
	NumericColumnOf[int8]
}

var _ Columnar = (*Int8Column)(nil)

func NewInt8Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Int8Column{
		NumericColumnOf: NewNumericColumnOf[int8](numRow),
	}
}

func (c Int8Column) Type() reflect.Type {
	return int8Type
}

func (c *Int8Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int8(v.Int()))
}

func (c *Int8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Int8()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Int8Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int8(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Int16Column struct {
	NumericColumnOf[int16]
}

var _ Columnar = (*Int16Column)(nil)

func NewInt16Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Int16Column{
		NumericColumnOf: NewNumericColumnOf[int16](numRow),
	}
}

func (c Int16Column) Type() reflect.Type {
	return int16Type
}

func (c *Int16Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int16(v.Int()))
}

func (c *Int16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Int16()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Int16Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int16(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Int32Column struct {
	NumericColumnOf[int32]
}

var _ Columnar = (*Int32Column)(nil)

func NewInt32Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Int32Column{
		NumericColumnOf: NewNumericColumnOf[int32](numRow),
	}
}

func (c Int32Column) Type() reflect.Type {
	return int32Type
}

func (c *Int32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int32(v.Int()))
}

func (c *Int32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Int32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Int32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int32(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Int64Column struct {
	NumericColumnOf[int64]
}

var _ Columnar = (*Int64Column)(nil)

func NewInt64Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Int64Column{
		NumericColumnOf: NewNumericColumnOf[int64](numRow),
	}
}

func (c Int64Column) Type() reflect.Type {
	return int64Type
}

func (c *Int64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Int())
}

func (c *Int64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Int64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int64(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Uint8Column struct {
	NumericColumnOf[uint8]
}

var _ Columnar = (*Uint8Column)(nil)

func NewUint8Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Uint8Column{
		NumericColumnOf: NewNumericColumnOf[uint8](numRow),
	}
}

func (c Uint8Column) Type() reflect.Type {
	return uint8Type
}

func (c *Uint8Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint8(v.Uint()))
}

func (c *Uint8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Uint8()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Uint8Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Uint8(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Uint16Column struct {
	NumericColumnOf[uint16]
}

var _ Columnar = (*Uint16Column)(nil)

func NewUint16Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Uint16Column{
		NumericColumnOf: NewNumericColumnOf[uint16](numRow),
	}
}

func (c Uint16Column) Type() reflect.Type {
	return uint16Type
}

func (c *Uint16Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint16(v.Uint()))
}

func (c *Uint16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Uint16()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Uint16Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Uint16(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Uint32Column struct {
	NumericColumnOf[uint32]
}

var _ Columnar = (*Uint32Column)(nil)

func NewUint32Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Uint32Column{
		NumericColumnOf: NewNumericColumnOf[uint32](numRow),
	}
}

func (c Uint32Column) Type() reflect.Type {
	return uint32Type
}

func (c *Uint32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint32(v.Uint()))
}

func (c *Uint32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Uint32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Uint32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Uint32(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Uint64Column struct {
	NumericColumnOf[uint64]
}

var _ Columnar = (*Uint64Column)(nil)

func NewUint64Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Uint64Column{
		NumericColumnOf: NewNumericColumnOf[uint64](numRow),
	}
}

func (c Uint64Column) Type() reflect.Type {
	return uint64Type
}

func (c *Uint64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Uint())
}

func (c *Uint64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Uint64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Uint64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Uint64(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Float32Column struct {
	NumericColumnOf[float32]
}

var _ Columnar = (*Float32Column)(nil)

func NewFloat32Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Float32Column{
		NumericColumnOf: NewNumericColumnOf[float32](numRow),
	}
}

func (c Float32Column) Type() reflect.Type {
	return float32Type
}

func (c *Float32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, float32(v.Float()))
}

func (c *Float32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Float32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Float32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Float32(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type Float64Column struct {
	NumericColumnOf[float64]
}

var _ Columnar = (*Float64Column)(nil)

func NewFloat64Column(typ reflect.Type, chType string, numRow int) Columnar {
	return &Float64Column{
		NumericColumnOf: NewNumericColumnOf[float64](numRow),
	}
}

func (c Float64Column) Type() reflect.Type {
	return float64Type
}

func (c *Float64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Float())
}

func (c *Float64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Float64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c Float64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Float64(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type StringColumn struct {
	ColumnOf[string]
}

var _ Columnar = (*StringColumn)(nil)

func NewStringColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &StringColumn{
		ColumnOf: NewColumnOf[string](numRow),
	}
}

func (c StringColumn) Type() reflect.Type {
	return stringType
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
	default:
		v.Set(reflect.ValueOf(c.Column[idx]))
		return nil
	}
	return fmt.Errorf("ch: can't scan %s into %s", "string", v.Type())
}

func (c *StringColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.String())
}

func (c *StringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.String()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c StringColumn) WriteTo(wr *chproto.Writer) error {
	for _, s := range c.Column {
		wr.String(s)
	}
	return nil
}

//------------------------------------------------------------------------------

type UUID [16]byte

// TODO: rework to use []byte
type UUIDColumn struct {
	ColumnOf[UUID]
}

var _ Columnar = (*UUIDColumn)(nil)

func NewUUIDColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &UUIDColumn{
		ColumnOf: NewColumnOf[UUID](numRow),
	}
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
	c.Alloc(numRow)

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

func NewIPColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &IPColumn{
		ColumnOf: NewColumnOf[net.IP](numRow),
	}
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
	c.Alloc(numRow)

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

type DateTimeColumn struct {
	ColumnOf[time.Time]
}

var _ Columnar = (*DateTimeColumn)(nil)

func NewDateTimeColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &DateTimeColumn{
		ColumnOf: NewColumnOf[time.Time](numRow),
	}
}

func (c DateTimeColumn) Type() reflect.Type {
	return timeType
}

func (c DateTimeColumn) ConvertAssign(idx int, v reflect.Value) error {
	v.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

func (c *DateTimeColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Interface().(time.Time))
}

func (c *DateTimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.DateTime()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c DateTimeColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		wr.DateTime(c.Column[i])
	}
	return nil
}

//------------------------------------------------------------------------------

type Int64TimeColumn struct {
	Int64Column
}

var _ Columnar = (*Int64TimeColumn)(nil)

func NewInt64TimeColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &Int64TimeColumn{
		Int64Column: Int64Column{
			NumericColumnOf: NewNumericColumnOf[int64](numRow),
		},
	}
}

func (c *Int64TimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

	for i := range c.Column {
		n, err := rd.Uint32()
		if err != nil {
			return err
		}
		c.Column[i] = int64(n) * int64(time.Second)
	}

	return nil
}

func (c Int64TimeColumn) WriteTo(wr *chproto.Writer) error {
	for i := range c.Column {
		wr.Uint32(uint32(c.Column[i] / int64(time.Second)))
	}
	return nil
}

//------------------------------------------------------------------------------

type DateColumn struct {
	DateTimeColumn
}

var _ Columnar = (*DateColumn)(nil)

func NewDateColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &DateColumn{
		DateTimeColumn: DateTimeColumn{
			ColumnOf: NewColumnOf[time.Time](numRow),
		},
	}
}

func (c *DateColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

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

func NewTimeColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &TimeColumn{
		DateTimeColumn: DateTimeColumn{
			ColumnOf: NewColumnOf[time.Time](numRow),
		},
	}
}

func (c *TimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.Alloc(numRow)

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

type BytesColumn struct {
	ColumnOf[[]byte]
}

var _ Columnar = (*BytesColumn)(nil)

func NewBytesColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &BytesColumn{
		ColumnOf: NewColumnOf[[]byte](numRow),
	}
}

func (c *BytesColumn) Reset(numRow int) {
	if cap(c.Column) >= numRow {
		for i := range c.Column {
			c.Column[i] = nil
		}
		c.Column = c.Column[:0]
	} else {
		c.Column = make([][]byte, 0, numRow)
	}
}

func (c BytesColumn) Type() reflect.Type {
	return bytesType
}

func (c BytesColumn) ConvertAssign(idx int, v reflect.Value) error {
	v.SetBytes(c.Column[idx])
	return nil
}

func (c *BytesColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Bytes())
}

func (c *BytesColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:numRow]
	} else {
		c.Column = make([][]byte, numRow)
	}

	for i := 0; i < len(c.Column); i++ {
		b, err := rd.Bytes()
		if err != nil {
			return err
		}
		c.Column[i] = b
	}

	return nil
}

func (c BytesColumn) WriteTo(wr *chproto.Writer) error {
	for _, b := range c.Column {
		wr.Bytes(b)
	}
	return nil
}

//------------------------------------------------------------------------------

type JSONColumn struct {
	BytesColumn
	Values []reflect.Value
}

var _ Columnar = (*JSONColumn)(nil)

func NewJSONColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return new(JSONColumn)
}

func (c *JSONColumn) Reset(numRow int) {
	c.Values = c.Values[:0]
	c.BytesColumn.Reset(numRow)
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

type EnumColumn struct {
	StringColumn
	enum *enumInfo
}

var _ Columnar = (*EnumColumn)(nil)

func NewEnumColumn(typ reflect.Type, chType string, numRow int) Columnar {
	return &EnumColumn{
		StringColumn: StringColumn{
			ColumnOf: NewColumnOf[string](numRow),
		},
		enum: parseEnum(chType),
	}
}

func (c *EnumColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if cap(c.Column) >= numRow {
		c.Column = c.Column[:numRow]
	} else {
		c.Column = make([]string, numRow)
	}

	for i := 0; i < len(c.Column); i++ {
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

type LCStringColumn struct {
	StringColumn
}

var _ Columnar = (*LCStringColumn)(nil)

func NewLCStringColumn(typ reflect.Type, chType string, numRow int) Columnar {
	col := new(LCStringColumn)
	col.Column = make([]string, 0, numRow)
	return col
}

func (c *LCStringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}
	if err := c.readPrefix(rd, numRow); err != nil {
		return err
	}
	return c.readData(rd, numRow)
}

func (c *LCStringColumn) readPrefix(rd *chproto.Reader, numRow int) error {
	version, err := rd.Int64()
	if err != nil {
		return err
	}
	if version != 1 {
		return fmt.Errorf("ch: got version=%d, wanted 1", version)
	}
	return nil
}

func (c *LCStringColumn) readData(rd *chproto.Reader, numRow int) error {
	flags, err := rd.Int64()
	if err != nil {
		return err
	}
	lcKey := newLCKeyType(flags & 0xf)

	dictSize, err := rd.Uint64()
	if err != nil {
		return err
	}

	dict := make([]string, dictSize)

	for i := range dict {
		s, err := rd.String()
		if err != nil {
			return err
		}
		dict[i] = s
	}

	numKey, err := rd.Uint64()
	if err != nil {
		return err
	}
	if int(numKey) != numRow {
		return fmt.Errorf("%d != %d", numKey, numRow)
	}

	if cap(c.Column) >= int(numKey) {
		c.Column = c.Column[:numKey]
	} else {
		c.Column = make([]string, numKey)
	}

	for i := 0; i < int(numKey); i++ {
		key, err := lcKey.read(rd)
		if err != nil {
			return err
		}
		c.Column[i] = dict[key]
	}

	return nil
}

func (c *LCStringColumn) WriteTo(wr *chproto.Writer) error {
	c.writePrefix(wr)
	c.writeData(wr)
	return nil
}

func (c *LCStringColumn) writePrefix(wr *chproto.Writer) {
	wr.Int64(1)
}

func (c *LCStringColumn) writeData(wr *chproto.Writer) {
	if len(c.Column) == 0 {
		return
	}

	keys := make([]int, len(c.Column))
	var lc lowCard

	for i, s := range c.Column {
		keys[i] = lc.Add(s)
	}

	const hasAdditionalKeys = 1 << 9
	const needUpdateDict = 1 << 10

	dict := lc.Dict()
	lcKey := newLCKey(len(dict))

	wr.Int64(int64(lcKey.typ) | hasAdditionalKeys | needUpdateDict)

	wr.Int64(int64(len(dict)))
	for _, s := range dict {
		wr.String(s)
	}

	wr.Int64(int64(len(keys)))
	for _, key := range keys {
		lcKey.write(wr, key)
	}
}

//------------------------------------------------------------------------------

type lcKey struct {
	typ   int8
	read  func(*chproto.Reader) (int, error)
	write func(*chproto.Writer, int)
}

func newLCKey(numKey int) lcKey {
	if numKey <= math.MaxUint8 {
		return newLCKeyType(0)
	}
	if numKey <= math.MaxUint16 {
		return newLCKeyType(1)
	}
	if numKey <= math.MaxUint32 {
		return newLCKeyType(2)
	}
	return newLCKeyType(3)
}

func newLCKeyType(typ int64) lcKey {
	switch typ {
	case 0:
		return lcKey{
			typ: 0,
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.Uint8()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.Uint8(uint8(n))
			},
		}
	case 1:
		return lcKey{
			typ: int8(1),
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.Uint16()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.Uint16(uint16(n))
			},
		}
	case 2:
		return lcKey{
			typ: 2,
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.Uint32()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.Uint32(uint32(n))
			},
		}
	case 3:
		return lcKey{
			typ: 3,
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.Uint64()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.Uint64(uint64(n))
			},
		}
	default:
		panic("not reached")
	}
}
