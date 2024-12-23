package chschema

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"iter"
	"math"
	"net"
	"reflect"
	"slices"
	"time"
	"unsafe"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/pkg/clickhouse/bfloat16"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"github.com/uptrace/pkg/msgp"
	"github.com/uptrace/pkg/msgp/msgpcode"
	"github.com/uptrace/pkg/unixtime"
	"github.com/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/constraints"
)

type Column struct {
	Name  string
	Type  string
	Field *Field
	Columnar
}

func (c *Column) String() string { return fmt.Sprintf("column=%s", c.Name) }

type Columnar interface {
	Init(chType string, goType reflect.Type) error
	Type() reflect.Type
	Clear()
	Grow(numRow int)
	SetValue(ptr unsafe.Pointer)
	AddPointer(ptr unsafe.Pointer)
	Value() any
	Len() int
	Index(idx int) any
	Slice(s, e int) any
	Values() iter.Seq[any]
	ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error
	ReadPrefix(rd *chproto.Reader) error
	ReadData(rd *chproto.Reader, numRow int) error
	WritePrefix(wr *chproto.Writer) error
	WriteData(wr *chproto.Writer) error
}
type CustomEncoding interface{ customEncoding() }
type ColumnOf[T any] struct{ Column []T }

func (c *ColumnOf[T]) Init(chType string, goType reflect.Type) error { return nil }
func (c *ColumnOf[T]) Clear()                                        { clear(c.Column) }
func (c *ColumnOf[T]) Grow(numRow int)                               { c.Column = slices.Grow(c.Column[:0], numRow) }
func (c *ColumnOf[T]) SetValue(ptr unsafe.Pointer)                   { column := *(*[]T)(ptr); c.Column = column }
func (c *ColumnOf[T]) AddPointer(ptr unsafe.Pointer)                 { c.Column = append(c.Column, *(*T)(ptr)) }
func (c *ColumnOf[T]) Value() any                                    { return c.Column }
func (c *ColumnOf[T]) Len() int                                      { return len(c.Column) }
func (c *ColumnOf[T]) Index(idx int) any                             { return c.Column[idx] }
func (c *ColumnOf[T]) Slice(s, e int) any                            { return c.Column[s:e:e] }
func (c *ColumnOf[T]) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, el := range c.Column {
			if !yield(el) {
				return
			}
		}
	}
}
func (c *ColumnOf[T]) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	dest := reflect.NewAt(typ, ptr)
	if scanner, ok := dest.Interface().(sql.Scanner); ok {
		return scanner.Scan(c.Column[idx])
	}
	dest = dest.Elem()
	dest.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}
func (c *ColumnOf[T]) ReadPrefix(rd *chproto.Reader) error  { return nil }
func (c *ColumnOf[T]) WritePrefix(wr *chproto.Writer) error { return nil }

type NumericColumnOf[T constraints.Integer | constraints.Float] struct{ ColumnOf[T] }

func (c *NumericColumnOf[T]) Clear() {}
func (c *NumericColumnOf[T]) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ.Kind() {
	case reflect.Int:
		*(*int)(ptr) = int(c.Column[idx])
		return nil
	case reflect.Int64:
		*(*int64)(ptr) = int64(c.Column[idx])
		return nil
	case reflect.Uint64:
		*(*uint64)(ptr) = uint64(c.Column[idx])
		return nil
	case reflect.Float64:
		*(*float64)(ptr) = float64(c.Column[idx])
		return nil
	case reflect.Float32:
		*(*float32)(ptr) = float32(c.Column[idx])
		return nil
	case reflect.Int8:
		*(*int8)(ptr) = int8(int64(c.Column[idx]))
		return nil
	case reflect.Int16:
		*(*int16)(ptr) = int16(int64(c.Column[idx]))
		return nil
	case reflect.Int32:
		*(*int32)(ptr) = int32(int64(c.Column[idx]))
		return nil
	case reflect.Uint:
		*(*uint)(ptr) = uint(c.Column[idx])
		return nil
	case reflect.Uint8:
		*(*uint8)(ptr) = uint8(uint64(c.Column[idx]))
		return nil
	case reflect.Uint16:
		*(*uint16)(ptr) = uint16(uint64(c.Column[idx]))
		return nil
	case reflect.Uint32:
		*(*uint32)(ptr) = uint32(uint64(c.Column[idx]))
		return nil
	case reflect.String:
		*(*string)(ptr) = fmt.Sprint(c.Column[idx])
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *BoolColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ.Kind() {
	case reflect.Bool:
		*(*bool)(ptr) = c.Column[idx]
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *StringColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ.Kind() {
	case reflect.String:
		*(*string)(ptr) = c.Column[idx]
		return nil
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			*(*[]byte)(ptr) = unsafeconv.Bytes(c.Column[idx])
			return nil
		}
		return fmt.Errorf("ch: can't scan %s into %s", "string", typ)
	case reflect.Struct:
		b := unsafeconv.Bytes(c.Column[idx])
		b, err := msgp.CodecFor(typ).Parse(b, ptr, 0)
		if err != nil {
			return err
		}
		if len(b) != 0 {
			return fmt.Errorf("msgp: buffer has unread data: %.100x", b)
		}
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *BytesColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ.Kind() {
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			*(*[]byte)(ptr) = c.Column[idx]
			return nil
		}
		return fmt.Errorf("ch: can't scan %s into %s", "string", typ)
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *DateTimeColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ {
	case unixNanoType:
		*(*unixtime.Nano)(ptr) = c.Column[idx]
		return nil
	case timeType:
		*(*time.Time)(ptr) = chproto.TimeUnix(0, int64(c.Column[idx]))
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *GoDateTimeColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ {
	case unixNanoType:
		*(*int64)(ptr) = c.Column[idx].UnixNano()
		return nil
	case timeType:
		*(*time.Time)(ptr) = c.Column[idx]
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}

const ipSize = 16

var zeroIP = make([]byte, ipSize)

type IPColumn struct{ ColumnOf[net.IP] }

var _ Columnar = (*IPColumn)(nil)

func NewIPColumn() Columnar           { return new(IPColumn) }
func (c IPColumn) Type() reflect.Type { return ipType }
func (c IPColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if typ == ipType {
		*(*[]byte)(ptr) = c.Column[idx]
		return nil
	}
	return c.ColumnOf.ConvertAssign(idx, typ, ptr)
}
func (c *IPColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	data := make([]byte, ipSize*numRow)
	var offset int
	for i := range c.Column {
		b := data[offset : offset+ipSize]
		offset += ipSize
		if _, err := io.ReadFull(rd, b); err != nil {
			return err
		}
		c.Column[i] = b
	}
	return nil
}
func (c IPColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if len(el) == 0 {
			wr.Write(zeroIP)
			continue
		}
		if len(el) != ipSize {
			return fmt.Errorf("got %d bytes, wanted %d", len(el), ipSize)
		}
		wr.Write(el)
	}
	return nil
}

type GoDateTime64Column struct {
	ColumnOf[time.Time]
	Mul int64
}

var _ Columnar = (*GoDateTime64Column)(nil)

func NewGoDateTime64Column() Columnar { return new(GoDateTime64Column) }
func (c *GoDateTime64Column) Init(chType string, goType reflect.Type) error {
	prec := parseDateTime64Prec(chType)
	c.Mul = int64(math.Pow10(9 - prec))
	return nil
}
func (c *GoDateTime64Column) Type() reflect.Type { return timeType }
func (c *GoDateTime64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ {
	case unixNanoType:
		*(*unixtime.Nano)(ptr) = unixtime.Nano(c.Column[idx].UnixNano())
		return nil
	case timeType:
		*(*time.Time)(ptr) = c.Column[idx]
		return nil
	}
	switch typ.Kind() {
	case reflect.Int64:
		*(*int64)(ptr) = c.Column[idx].UnixNano()
		return nil
	default:
		return c.ColumnOf.ConvertAssign(idx, typ, ptr)
	}
}
func (c *GoDateTime64Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		n, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = chproto.TimeUnix(0, n*c.Mul)
	}
	return nil
}
func (c *GoDateTime64Column) WriteData(wr *chproto.Writer) error {
	for _, tm := range c.Column {
		wr.Int64(tm.UnixNano() / c.Mul)
	}
	return nil
}

const timePrecision = int64(time.Microsecond)

type TimeColumn struct{ DateTime64Column }

var _ Columnar = (*TimeColumn)(nil)

func NewTimeColumn() Columnar { col := new(TimeColumn); col.Mul = int64(timePrecision); return col }

type DateTime64Column struct {
	Int64Column
	Mul int64
}

var _ Columnar = (*DateTime64Column)(nil)

func NewDateTime64Column() Columnar { return new(DateTime64Column) }
func (c *DateTime64Column) Init(chType string, goType reflect.Type) error {
	prec := parseDateTime64Prec(chType)
	c.Mul = int64(math.Pow10(9 - prec))
	return nil
}
func (c *DateTime64Column) Index(idx int) any { return unixtime.Nano(c.Column[idx] * c.Mul) }
func (c *DateTime64Column) AddPointer(ptr unsafe.Pointer) {
	num := *(*int64)(ptr)
	c.Column = append(c.Column, num/c.Mul)
}
func (c *DateTime64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	switch typ {
	case unixNanoType:
		*(*unixtime.Nano)(ptr) = unixtime.Nano(c.Column[idx] * c.Mul)
		return nil
	case timeType:
		*(*time.Time)(ptr) = chproto.TimeUnix(0, c.Column[idx]*c.Mul)
		return nil
	}
	switch typ.Kind() {
	case reflect.Int64:
		*(*int64)(ptr) = c.Column[idx] * c.Mul
		return nil
	default:
		return fmt.Errorf("ch: can't scan DateTime64 into %q", typ)
	}
}
func (c *DateTime64Column) At(index int) unixtime.Nano { return unixtime.Nano(c.Column[index] * c.Mul) }

type EnumColumn struct {
	StringColumn
	enum *EnumInfo
}

var _ Columnar = (*EnumColumn)(nil)

func NewEnumColumn() Columnar { return new(EnumColumn) }
func (c *EnumColumn) Init(chType string, goType reflect.Type) error {
	enum, err := ParseEnum(chType)
	if err != nil {
		return err
	}
	c.enum = enum
	return nil
}
func (c *EnumColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		n, err := rd.Int8()
		if err != nil {
			return err
		}
		c.Column[i] = c.enum.Decode(int16(n))
	}
	return nil
}
func (c *EnumColumn) WriteData(wr *chproto.Writer) error {
	for _, s := range c.Column {
		num := c.enum.Encode(s)
		wr.Int8(int8(num))
	}
	return nil
}

const nullVariant = 255

type DynamicColumn struct {
	CustomEncoding
	ColumnOf[any]
	types   []string
	layout  []uint8
	offsets []int
	counts  []int
	columns []Columnar
}

var _ Columnar = (*DynamicColumn)(nil)

func NewDynamicColumn() Columnar            { return new(DynamicColumn) }
func (c *DynamicColumn) Type() reflect.Type { return anyType }
func (c *DynamicColumn) ReadData(rd *chproto.Reader, numRow int) error {
	version, err := rd.UInt64()
	if err != nil {
		return fmt.Errorf("failed to read Dynamic version: %w", err)
	}
	if version != 1 {
		return fmt.Errorf("unsupported Dynamic version: %d", version)
	}
	numType, err := rd.UInt8()
	if err != nil {
		return err
	}
	numType2, err := rd.UInt8()
	if err != nil {
		return err
	}
	if numType2 != numType {
		return fmt.Errorf("Dynamic number of types don't match: %d != %d", numType2, numType)
	}
	c.types = slices.Grow(c.types, int(numType))
	c.types = c.types[:numType]
	for i := 0; i < int(numType); i++ {
		typeName, err := rd.String()
		if err != nil {
			return err
		}
		c.types[i] = typeName
	}
	version2, err := rd.UInt64()
	if err != nil {
		return fmt.Errorf("failed to read Dynamic version2: %w", err)
	}
	if version2 != 0 {
		return fmt.Errorf("unsupported Dynamic version2: %d", version)
	}
	c.layout = slices.Grow(c.layout, numRow)
	c.layout = c.layout[:numRow]
	c.offsets = slices.Grow(c.offsets, numRow)
	c.offsets = c.offsets[:numRow]
	c.counts = slices.Grow(c.counts, int(numType))
	c.counts = c.counts[:numType]
	defer clear(c.counts)
	for i := 0; i < numRow; i++ {
		typeIndex, err := rd.UInt8()
		if err != nil {
			return fmt.Errorf("failed to read Dynamic type at index %d: %w", i, err)
		}
		if int(typeIndex) <= len(c.types)/2 {
			typeIndex++
		}
		if typeIndex > 0 {
			typeIndex--
		}
		c.layout[i] = typeIndex
		if typeIndex == nullVariant {
			continue
		}
		count := c.counts[typeIndex]
		c.counts[typeIndex] = count + 1
		c.offsets[i] = count
	}
	c.columns = slices.Grow(c.columns, int(numType))
	c.columns = c.columns[:numType]
	defer clear(c.columns)
	for typeIndex, numRow := range c.counts {
		if numRow == 0 {
			continue
		}
		col := c.columns[typeIndex]
		if col == nil {
			col = NewColumn(c.types[typeIndex], nil)
			c.columns[typeIndex] = col
		}
		col.Grow(numRow)
		if err := col.ReadData(rd, numRow); err != nil {
			return err
		}
	}
	c.Column = c.Column[:len(c.layout)]
	for i, typeIndex := range c.layout {
		if typeIndex == nullVariant {
			c.Column[i] = nil
			continue
		}
		col := c.columns[typeIndex]
		offset := c.offsets[i]
		c.Column[i] = col.Index(offset)
	}
	return nil
}
func (c *DynamicColumn) WriteData(wr *chproto.Writer) error { panic("not implemented") }

type baseJSONColumn struct {
	typ    reflect.Type
	values []unsafe.Pointer
}

func (c *baseJSONColumn) Init(chType string, goType reflect.Type) error { c.typ = goType; return nil }
func (c *baseJSONColumn) ResetForWriting(numRow int) {
	if cap(c.values) >= numRow {
		c.values = c.values[:0]
	} else {
		c.values = make([]unsafe.Pointer, 0, numRow)
	}
}
func (c *baseJSONColumn) AddPointer(p unsafe.Pointer) { c.values = append(c.values, p) }
func (c *baseJSONColumn) WriteData(wr *chproto.Writer) error {
	codec := json.CodecFor(c.typ)
	var buf []byte
	for _, ptr := range c.values {
		var err error
		buf, err = codec.Append(buf[:0], ptr, json.SortMapKeys)
		if err != nil {
			return err
		}
		if isZeroJSON(buf) {
			wr.Bytes(nil)
		} else {
			wr.Bytes(buf)
		}
	}
	return nil
}
func isZeroJSON(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	return bytes.Equal(b, []byte("null"))
}

type JSONColumn struct {
	baseJSONColumn
	DynamicColumn
}

var _ Columnar = (*JSONColumn)(nil)

func NewJSONColumn() Columnar                    { return new(JSONColumn) }
func (c *JSONColumn) ResetForWriting(numRow int) { c.baseJSONColumn.ResetForWriting(numRow) }
func (c *JSONColumn) Len() int {
	if len(c.values) > 0 {
		return len(c.values)
	}
	return len(c.Column)
}
func (c *JSONColumn) WriteData(wr *chproto.Writer) error {
	wr.UInt64(1)
	return c.baseJSONColumn.WriteData(wr)
}

type JSONBytesColumn struct {
	baseJSONColumn
	BytesColumn
}

var _ Columnar = (*JSONBytesColumn)(nil)

func NewJSONBytesColumn() Columnar { return new(JSONBytesColumn) }
func (c *JSONBytesColumn) Clear()  { clear(c.values); c.BytesColumn.Clear() }
func (c *JSONBytesColumn) Grow(numRow int) {
	c.values = slices.Grow(c.values[:0], numRow)
	c.BytesColumn.Grow(numRow)
}
func (c *JSONBytesColumn) Len() int {
	if len(c.values) > 0 {
		return len(c.values)
	}
	return len(c.Column)
}
func (c *JSONBytesColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	dec := json.NewDecoder(bytes.NewReader(c.Column[idx]))
	dec.UseNumber()
	dest := reflect.NewAt(typ, ptr).Elem()
	return dec.Decode(dest.Interface())
}
func (c *JSONBytesColumn) AddPointer(p unsafe.Pointer)        { c.baseJSONColumn.AddPointer(p) }
func (c *JSONBytesColumn) WriteData(wr *chproto.Writer) error { return c.baseJSONColumn.WriteData(wr) }

type MsgpackColumn struct {
	BytesColumn
	typ    reflect.Type
	values []unsafe.Pointer
}

var _ Columnar = (*MsgpackColumn)(nil)

func NewMsgpackColumn() Columnar                                       { return new(MsgpackColumn) }
func (c *MsgpackColumn) Init(chType string, goType reflect.Type) error { c.typ = goType; return nil }
func (c *MsgpackColumn) Clear()                                        { clear(c.values); c.BytesColumn.Clear() }
func (c *MsgpackColumn) Grow(numRow int) {
	c.values = slices.Grow(c.values[:0], numRow)
	c.BytesColumn.Grow(numRow)
}
func (c *MsgpackColumn) Len() int {
	if len(c.values) > 0 {
		return len(c.values)
	}
	return len(c.Column)
}
func (c *MsgpackColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	b := c.Column[idx]
	if len(b) == 0 {
		return nil
	}
	b, err := msgp.CodecFor(typ).Parse(b, ptr, 0)
	if err != nil {
		return err
	}
	if len(b) != 0 {
		return fmt.Errorf("msgp: buffer has unread data: %.100x", b)
	}
	return nil
}
func (c *MsgpackColumn) AddPointer(p unsafe.Pointer) { c.values = append(c.values, p) }
func (c *MsgpackColumn) WriteData(wr *chproto.Writer) error {
	if len(c.values) == 0 {
		return nil
	}
	codec := msgp.CodecFor(c.typ)
	var buf []byte
	for _, ptr := range c.values {
		var err error
		buf, err = codec.Append(buf[:0], ptr, msgp.SortedMapKeys)
		if err != nil {
			return err
		}
		if isZeroMsgpack(buf) {
			wr.Bytes(nil)
		} else {
			wr.Bytes(buf)
		}
	}
	return nil
}
func isZeroMsgpack(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	if len(b) > 1 {
		return false
	}
	switch c := b[0]; {
	case c == msgpcode.Nil:
		return true
	case c >= msgpcode.FixedMapLow && c <= msgpcode.FixedMapHigh:
		return int(c&msgpcode.FixedMapMask) == 0
	case c >= msgpcode.FixedArrayLow && c <= msgpcode.FixedArrayHigh:
		return int(c&msgpcode.FixedArrayMask) == 0
	}
	return false
}

type TDigestColumn struct{ ColumnOf[[]float32] }

var _ Columnar = (*TDigestColumn)(nil)

func NewTDigestColumn() Columnar           { return new(TDigestColumn) }
func (c TDigestColumn) Type() reflect.Type { return sliceFloat32Type }
func (c *TDigestColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Interface().([]float32))
}
func (c *TDigestColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		n, err := rd.Uvarint()
		if err != nil {
			return err
		}
		data := make([]float32, 0, 2*n)
		for j := 0; j < int(n); j++ {
			value, err := rd.Float32()
			if err != nil {
				return err
			}
			count, err := rd.Float32()
			if err != nil {
				return err
			}
			data = append(data, value, count)
		}
		c.Column[i] = data
	}
	return nil
}
func (c TDigestColumn) WriteData(wr *chproto.Writer) error {
	for _, data := range c.Column {
		n := len(data) / 2
		wr.Uvarint(uint64(n))
		for _, num := range data {
			wr.Float32(num)
		}
	}
	return nil
}

type TDigestMapColumn struct {
	ColumnOf[map[bfloat16.T]uint64]
}

var _ Columnar = (*TDigestMapColumn)(nil)

func NewTDigestMapColumn() Columnar           { return new(TDigestMapColumn) }
func (c TDigestMapColumn) Type() reflect.Type { return sliceFloat32Type }
func (c *TDigestMapColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, v.Interface().(map[bfloat16.T]uint64))
}
func (c *TDigestMapColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		n, err := rd.Uvarint()
		if err != nil {
			return err
		}
		data := make(map[bfloat16.T]uint64, n)
		for j := 0; j < int(n); j++ {
			value, err := rd.Float32()
			if err != nil {
				return err
			}
			count, err := rd.Float32()
			if err != nil {
				return err
			}
			data[bfloat16.From32(value)] = uint64(count)
		}
		c.Column[i] = data
	}
	return nil
}
func (c TDigestMapColumn) WriteData(wr *chproto.Writer) error {
	for _, data := range c.Column {
		n := len(data)
		wr.Uvarint(uint64(n))
		for value, count := range data {
			wr.Float32(value.Float32())
			wr.Float32(float32(count))
		}
	}
	return nil
}
func readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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
