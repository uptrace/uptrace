package chschema

import (
	"reflect"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

type Int8Column struct {
	NumericColumnOf[int8]
}

var _ Columnar = (*Int8Column)(nil)

func NewInt8Column() Columnar {
	return new(Int8Column)
}

var _Int8Type = reflect.TypeOf((*int8)(nil)).Elem()

func (c *Int8Column) Type() reflect.Type {
	return _Int8Type
}

func (c *Int8Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int8(v.Int()))
}

//------------------------------------------------------------------------------

type ArrayInt8Column struct {
	ColumnOf[[]int8]
	elem Int8Column
}

var (
	_ Columnar      = (*ArrayInt8Column)(nil)
	_ ArrayColumnar = (*ArrayInt8Column)(nil)
)

func NewArrayInt8Column() Columnar {
	return new(ArrayInt8Column)
}

func (c *ArrayInt8Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayInt8Column) Type() reflect.Type {
	return reflect.TypeOf((*[]int8)(nil)).Elem()
}

func (c *ArrayInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayInt8Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayInt8Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayInt8Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayInt8Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayInt8Column struct {
	ColumnOf[[][]int8]
	elem ArrayInt8Column
}

var (
	_ Columnar      = (*ArrayArrayInt8Column)(nil)
	_ ArrayColumnar = (*ArrayArrayInt8Column)(nil)
)

func NewArrayArrayInt8Column() Columnar {
	return new(ArrayArrayInt8Column)
}

func (c *ArrayArrayInt8Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayInt8Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]int8)(nil)).Elem()
}

func (c *ArrayArrayInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayInt8Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayInt8Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayInt8Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayInt8Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type UInt8Column struct {
	NumericColumnOf[uint8]
}

var _ Columnar = (*UInt8Column)(nil)

func NewUInt8Column() Columnar {
	return new(UInt8Column)
}

var _UInt8Type = reflect.TypeOf((*uint8)(nil)).Elem()

func (c *UInt8Column) Type() reflect.Type {
	return _UInt8Type
}

func (c *UInt8Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint8(v.Uint()))
}

//------------------------------------------------------------------------------

type ArrayUInt8Column struct {
	ColumnOf[[]uint8]
	elem UInt8Column
}

var (
	_ Columnar      = (*ArrayUInt8Column)(nil)
	_ ArrayColumnar = (*ArrayUInt8Column)(nil)
)

func NewArrayUInt8Column() Columnar {
	return new(ArrayUInt8Column)
}

func (c *ArrayUInt8Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayUInt8Column) Type() reflect.Type {
	return reflect.TypeOf((*[]uint8)(nil)).Elem()
}

func (c *ArrayUInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayUInt8Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayUInt8Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayUInt8Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayUInt8Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayUInt8Column struct {
	ColumnOf[[][]uint8]
	elem ArrayUInt8Column
}

var (
	_ Columnar      = (*ArrayArrayUInt8Column)(nil)
	_ ArrayColumnar = (*ArrayArrayUInt8Column)(nil)
)

func NewArrayArrayUInt8Column() Columnar {
	return new(ArrayArrayUInt8Column)
}

func (c *ArrayArrayUInt8Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayUInt8Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]uint8)(nil)).Elem()
}

func (c *ArrayArrayUInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayUInt8Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayUInt8Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayUInt8Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayUInt8Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type Int16Column struct {
	NumericColumnOf[int16]
}

var _ Columnar = (*Int16Column)(nil)

func NewInt16Column() Columnar {
	return new(Int16Column)
}

var _Int16Type = reflect.TypeOf((*int16)(nil)).Elem()

func (c *Int16Column) Type() reflect.Type {
	return _Int16Type
}

func (c *Int16Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int16(v.Int()))
}

//------------------------------------------------------------------------------

type ArrayInt16Column struct {
	ColumnOf[[]int16]
	elem Int16Column
}

var (
	_ Columnar      = (*ArrayInt16Column)(nil)
	_ ArrayColumnar = (*ArrayInt16Column)(nil)
)

func NewArrayInt16Column() Columnar {
	return new(ArrayInt16Column)
}

func (c *ArrayInt16Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayInt16Column) Type() reflect.Type {
	return reflect.TypeOf((*[]int16)(nil)).Elem()
}

func (c *ArrayInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayInt16Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayInt16Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayInt16Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayInt16Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayInt16Column struct {
	ColumnOf[[][]int16]
	elem ArrayInt16Column
}

var (
	_ Columnar      = (*ArrayArrayInt16Column)(nil)
	_ ArrayColumnar = (*ArrayArrayInt16Column)(nil)
)

func NewArrayArrayInt16Column() Columnar {
	return new(ArrayArrayInt16Column)
}

func (c *ArrayArrayInt16Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayInt16Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]int16)(nil)).Elem()
}

func (c *ArrayArrayInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayInt16Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayInt16Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayInt16Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayInt16Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type UInt16Column struct {
	NumericColumnOf[uint16]
}

var _ Columnar = (*UInt16Column)(nil)

func NewUInt16Column() Columnar {
	return new(UInt16Column)
}

var _UInt16Type = reflect.TypeOf((*uint16)(nil)).Elem()

func (c *UInt16Column) Type() reflect.Type {
	return _UInt16Type
}

func (c *UInt16Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint16(v.Uint()))
}

//------------------------------------------------------------------------------

type ArrayUInt16Column struct {
	ColumnOf[[]uint16]
	elem UInt16Column
}

var (
	_ Columnar      = (*ArrayUInt16Column)(nil)
	_ ArrayColumnar = (*ArrayUInt16Column)(nil)
)

func NewArrayUInt16Column() Columnar {
	return new(ArrayUInt16Column)
}

func (c *ArrayUInt16Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayUInt16Column) Type() reflect.Type {
	return reflect.TypeOf((*[]uint16)(nil)).Elem()
}

func (c *ArrayUInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayUInt16Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayUInt16Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayUInt16Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayUInt16Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayUInt16Column struct {
	ColumnOf[[][]uint16]
	elem ArrayUInt16Column
}

var (
	_ Columnar      = (*ArrayArrayUInt16Column)(nil)
	_ ArrayColumnar = (*ArrayArrayUInt16Column)(nil)
)

func NewArrayArrayUInt16Column() Columnar {
	return new(ArrayArrayUInt16Column)
}

func (c *ArrayArrayUInt16Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayUInt16Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]uint16)(nil)).Elem()
}

func (c *ArrayArrayUInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayUInt16Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayUInt16Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayUInt16Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayUInt16Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type Int32Column struct {
	NumericColumnOf[int32]
}

var _ Columnar = (*Int32Column)(nil)

func NewInt32Column() Columnar {
	return new(Int32Column)
}

var _Int32Type = reflect.TypeOf((*int32)(nil)).Elem()

func (c *Int32Column) Type() reflect.Type {
	return _Int32Type
}

func (c *Int32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int32(v.Int()))
}

//------------------------------------------------------------------------------

type ArrayInt32Column struct {
	ColumnOf[[]int32]
	elem Int32Column
}

var (
	_ Columnar      = (*ArrayInt32Column)(nil)
	_ ArrayColumnar = (*ArrayInt32Column)(nil)
)

func NewArrayInt32Column() Columnar {
	return new(ArrayInt32Column)
}

func (c *ArrayInt32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayInt32Column) Type() reflect.Type {
	return reflect.TypeOf((*[]int32)(nil)).Elem()
}

func (c *ArrayInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayInt32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayInt32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayInt32Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayInt32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayInt32Column struct {
	ColumnOf[[][]int32]
	elem ArrayInt32Column
}

var (
	_ Columnar      = (*ArrayArrayInt32Column)(nil)
	_ ArrayColumnar = (*ArrayArrayInt32Column)(nil)
)

func NewArrayArrayInt32Column() Columnar {
	return new(ArrayArrayInt32Column)
}

func (c *ArrayArrayInt32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayInt32Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]int32)(nil)).Elem()
}

func (c *ArrayArrayInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayInt32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayInt32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayInt32Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayInt32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type UInt32Column struct {
	NumericColumnOf[uint32]
}

var _ Columnar = (*UInt32Column)(nil)

func NewUInt32Column() Columnar {
	return new(UInt32Column)
}

var _UInt32Type = reflect.TypeOf((*uint32)(nil)).Elem()

func (c *UInt32Column) Type() reflect.Type {
	return _UInt32Type
}

func (c *UInt32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint32(v.Uint()))
}

//------------------------------------------------------------------------------

type ArrayUInt32Column struct {
	ColumnOf[[]uint32]
	elem UInt32Column
}

var (
	_ Columnar      = (*ArrayUInt32Column)(nil)
	_ ArrayColumnar = (*ArrayUInt32Column)(nil)
)

func NewArrayUInt32Column() Columnar {
	return new(ArrayUInt32Column)
}

func (c *ArrayUInt32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayUInt32Column) Type() reflect.Type {
	return reflect.TypeOf((*[]uint32)(nil)).Elem()
}

func (c *ArrayUInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayUInt32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayUInt32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayUInt32Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayUInt32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayUInt32Column struct {
	ColumnOf[[][]uint32]
	elem ArrayUInt32Column
}

var (
	_ Columnar      = (*ArrayArrayUInt32Column)(nil)
	_ ArrayColumnar = (*ArrayArrayUInt32Column)(nil)
)

func NewArrayArrayUInt32Column() Columnar {
	return new(ArrayArrayUInt32Column)
}

func (c *ArrayArrayUInt32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayUInt32Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]uint32)(nil)).Elem()
}

func (c *ArrayArrayUInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayUInt32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayUInt32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayUInt32Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayUInt32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type Int64Column struct {
	NumericColumnOf[int64]
}

var _ Columnar = (*Int64Column)(nil)

func NewInt64Column() Columnar {
	return new(Int64Column)
}

var _Int64Type = reflect.TypeOf((*int64)(nil)).Elem()

func (c *Int64Column) Type() reflect.Type {
	return _Int64Type
}

func (c *Int64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, int64(v.Int()))
}

//------------------------------------------------------------------------------

type ArrayInt64Column struct {
	ColumnOf[[]int64]
	elem Int64Column
}

var (
	_ Columnar      = (*ArrayInt64Column)(nil)
	_ ArrayColumnar = (*ArrayInt64Column)(nil)
)

func NewArrayInt64Column() Columnar {
	return new(ArrayInt64Column)
}

func (c *ArrayInt64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayInt64Column) Type() reflect.Type {
	return reflect.TypeOf((*[]int64)(nil)).Elem()
}

func (c *ArrayInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayInt64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayInt64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayInt64Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayInt64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayInt64Column struct {
	ColumnOf[[][]int64]
	elem ArrayInt64Column
}

var (
	_ Columnar      = (*ArrayArrayInt64Column)(nil)
	_ ArrayColumnar = (*ArrayArrayInt64Column)(nil)
)

func NewArrayArrayInt64Column() Columnar {
	return new(ArrayArrayInt64Column)
}

func (c *ArrayArrayInt64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayInt64Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]int64)(nil)).Elem()
}

func (c *ArrayArrayInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayInt64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayInt64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayInt64Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayInt64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type UInt64Column struct {
	NumericColumnOf[uint64]
}

var _ Columnar = (*UInt64Column)(nil)

func NewUInt64Column() Columnar {
	return new(UInt64Column)
}

var _UInt64Type = reflect.TypeOf((*uint64)(nil)).Elem()

func (c *UInt64Column) Type() reflect.Type {
	return _UInt64Type
}

func (c *UInt64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, uint64(v.Uint()))
}

//------------------------------------------------------------------------------

type ArrayUInt64Column struct {
	ColumnOf[[]uint64]
	elem UInt64Column
}

var (
	_ Columnar      = (*ArrayUInt64Column)(nil)
	_ ArrayColumnar = (*ArrayUInt64Column)(nil)
)

func NewArrayUInt64Column() Columnar {
	return new(ArrayUInt64Column)
}

func (c *ArrayUInt64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayUInt64Column) Type() reflect.Type {
	return reflect.TypeOf((*[]uint64)(nil)).Elem()
}

func (c *ArrayUInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayUInt64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayUInt64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayUInt64Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayUInt64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayUInt64Column struct {
	ColumnOf[[][]uint64]
	elem ArrayUInt64Column
}

var (
	_ Columnar      = (*ArrayArrayUInt64Column)(nil)
	_ ArrayColumnar = (*ArrayArrayUInt64Column)(nil)
)

func NewArrayArrayUInt64Column() Columnar {
	return new(ArrayArrayUInt64Column)
}

func (c *ArrayArrayUInt64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayUInt64Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]uint64)(nil)).Elem()
}

func (c *ArrayArrayUInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayUInt64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayUInt64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayUInt64Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayUInt64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type Float32Column struct {
	NumericColumnOf[float32]
}

var _ Columnar = (*Float32Column)(nil)

func NewFloat32Column() Columnar {
	return new(Float32Column)
}

var _Float32Type = reflect.TypeOf((*float32)(nil)).Elem()

func (c *Float32Column) Type() reflect.Type {
	return _Float32Type
}

func (c *Float32Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, float32(v.Float()))
}

//------------------------------------------------------------------------------

type ArrayFloat32Column struct {
	ColumnOf[[]float32]
	elem Float32Column
}

var (
	_ Columnar      = (*ArrayFloat32Column)(nil)
	_ ArrayColumnar = (*ArrayFloat32Column)(nil)
)

func NewArrayFloat32Column() Columnar {
	return new(ArrayFloat32Column)
}

func (c *ArrayFloat32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayFloat32Column) Type() reflect.Type {
	return reflect.TypeOf((*[]float32)(nil)).Elem()
}

func (c *ArrayFloat32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayFloat32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayFloat32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayFloat32Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayFloat32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayFloat32Column struct {
	ColumnOf[[][]float32]
	elem ArrayFloat32Column
}

var (
	_ Columnar      = (*ArrayArrayFloat32Column)(nil)
	_ ArrayColumnar = (*ArrayArrayFloat32Column)(nil)
)

func NewArrayArrayFloat32Column() Columnar {
	return new(ArrayArrayFloat32Column)
}

func (c *ArrayArrayFloat32Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayFloat32Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]float32)(nil)).Elem()
}

func (c *ArrayArrayFloat32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayFloat32Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayFloat32Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayFloat32Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayFloat32Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type Float64Column struct {
	NumericColumnOf[float64]
}

var _ Columnar = (*Float64Column)(nil)

func NewFloat64Column() Columnar {
	return new(Float64Column)
}

var _Float64Type = reflect.TypeOf((*float64)(nil)).Elem()

func (c *Float64Column) Type() reflect.Type {
	return _Float64Type
}

func (c *Float64Column) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, float64(v.Float()))
}

//------------------------------------------------------------------------------

type ArrayFloat64Column struct {
	ColumnOf[[]float64]
	elem Float64Column
}

var (
	_ Columnar      = (*ArrayFloat64Column)(nil)
	_ ArrayColumnar = (*ArrayFloat64Column)(nil)
)

func NewArrayFloat64Column() Columnar {
	return new(ArrayFloat64Column)
}

func (c *ArrayFloat64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayFloat64Column) Type() reflect.Type {
	return reflect.TypeOf((*[]float64)(nil)).Elem()
}

func (c *ArrayFloat64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayFloat64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayFloat64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayFloat64Column) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayFloat64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayFloat64Column struct {
	ColumnOf[[][]float64]
	elem ArrayFloat64Column
}

var (
	_ Columnar      = (*ArrayArrayFloat64Column)(nil)
	_ ArrayColumnar = (*ArrayArrayFloat64Column)(nil)
)

func NewArrayArrayFloat64Column() Columnar {
	return new(ArrayArrayFloat64Column)
}

func (c *ArrayArrayFloat64Column) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayFloat64Column) Type() reflect.Type {
	return reflect.TypeOf((*[][]float64)(nil)).Elem()
}

func (c *ArrayArrayFloat64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayFloat64Column) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayFloat64Column) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayFloat64Column) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayFloat64Column) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type BoolColumn struct {
	ColumnOf[bool]
}

var _ Columnar = (*BoolColumn)(nil)

func NewBoolColumn() Columnar {
	return new(BoolColumn)
}

var _BoolType = reflect.TypeOf((*bool)(nil)).Elem()

func (c *BoolColumn) Type() reflect.Type {
	return _BoolType
}

func (c *BoolColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, bool(v.Bool()))
}

func (c *BoolColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Bool()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *BoolColumn) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Bool(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayBoolColumn struct {
	ColumnOf[[]bool]
	elem BoolColumn
}

var (
	_ Columnar      = (*ArrayBoolColumn)(nil)
	_ ArrayColumnar = (*ArrayBoolColumn)(nil)
)

func NewArrayBoolColumn() Columnar {
	return new(ArrayBoolColumn)
}

func (c *ArrayBoolColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayBoolColumn) Type() reflect.Type {
	return reflect.TypeOf((*[]bool)(nil)).Elem()
}

func (c *ArrayBoolColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayBoolColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayBoolColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayBoolColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayBoolColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayBoolColumn struct {
	ColumnOf[[][]bool]
	elem ArrayBoolColumn
}

var (
	_ Columnar      = (*ArrayArrayBoolColumn)(nil)
	_ ArrayColumnar = (*ArrayArrayBoolColumn)(nil)
)

func NewArrayArrayBoolColumn() Columnar {
	return new(ArrayArrayBoolColumn)
}

func (c *ArrayArrayBoolColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayBoolColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][]bool)(nil)).Elem()
}

func (c *ArrayArrayBoolColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayBoolColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayBoolColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayBoolColumn) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayBoolColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type StringColumn struct {
	ColumnOf[string]
}

var _ Columnar = (*StringColumn)(nil)

func NewStringColumn() Columnar {
	return new(StringColumn)
}

var _StringType = reflect.TypeOf((*string)(nil)).Elem()

func (c *StringColumn) Type() reflect.Type {
	return _StringType
}

func (c *StringColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, string(v.String()))
}

func (c *StringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.String()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *StringColumn) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.String(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayStringColumn struct {
	ColumnOf[[]string]
	elem StringColumn
}

var (
	_ Columnar      = (*ArrayStringColumn)(nil)
	_ ArrayColumnar = (*ArrayStringColumn)(nil)
)

func NewArrayStringColumn() Columnar {
	return new(ArrayStringColumn)
}

func (c *ArrayStringColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayStringColumn) Type() reflect.Type {
	return reflect.TypeOf((*[]string)(nil)).Elem()
}

func (c *ArrayStringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayStringColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayStringColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayStringColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayStringColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayStringColumn struct {
	ColumnOf[[][]string]
	elem ArrayStringColumn
}

var (
	_ Columnar      = (*ArrayArrayStringColumn)(nil)
	_ ArrayColumnar = (*ArrayArrayStringColumn)(nil)
)

func NewArrayArrayStringColumn() Columnar {
	return new(ArrayArrayStringColumn)
}

func (c *ArrayArrayStringColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayStringColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][]string)(nil)).Elem()
}

func (c *ArrayArrayStringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayStringColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayStringColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayStringColumn) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayStringColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type BytesColumn struct {
	ColumnOf[[]byte]
}

var _ Columnar = (*BytesColumn)(nil)

func NewBytesColumn() Columnar {
	return new(BytesColumn)
}

var _BytesType = reflect.TypeOf((*[]byte)(nil)).Elem()

func (c *BytesColumn) Type() reflect.Type {
	return _BytesType
}

func (c *BytesColumn) AppendValue(v reflect.Value) {
	c.Column = append(c.Column, []byte(v.Bytes()))
}

func (c *BytesColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Bytes()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *BytesColumn) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Bytes(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayBytesColumn struct {
	ColumnOf[[][]byte]
	elem BytesColumn
}

var (
	_ Columnar      = (*ArrayBytesColumn)(nil)
	_ ArrayColumnar = (*ArrayBytesColumn)(nil)
)

func NewArrayBytesColumn() Columnar {
	return new(ArrayBytesColumn)
}

func (c *ArrayBytesColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayBytesColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][]byte)(nil)).Elem()
}

func (c *ArrayBytesColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayBytesColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayBytesColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayBytesColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayBytesColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayBytesColumn struct {
	ColumnOf[[][][]byte]
	elem ArrayBytesColumn
}

var (
	_ Columnar      = (*ArrayArrayBytesColumn)(nil)
	_ ArrayColumnar = (*ArrayArrayBytesColumn)(nil)
)

func NewArrayArrayBytesColumn() Columnar {
	return new(ArrayArrayBytesColumn)
}

func (c *ArrayArrayBytesColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayBytesColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][][]byte)(nil)).Elem()
}

func (c *ArrayArrayBytesColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayBytesColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayBytesColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayBytesColumn) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayBytesColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayEnumColumn struct {
	ColumnOf[[]string]
	elem EnumColumn
}

var (
	_ Columnar      = (*ArrayEnumColumn)(nil)
	_ ArrayColumnar = (*ArrayEnumColumn)(nil)
)

func NewArrayEnumColumn() Columnar {
	return new(ArrayEnumColumn)
}

func (c *ArrayEnumColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayEnumColumn) Type() reflect.Type {
	return reflect.TypeOf((*[]string)(nil)).Elem()
}

func (c *ArrayEnumColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayEnumColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayEnumColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayEnumColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayEnumColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayEnumColumn struct {
	ColumnOf[[][]string]
	elem ArrayEnumColumn
}

var (
	_ Columnar      = (*ArrayArrayEnumColumn)(nil)
	_ ArrayColumnar = (*ArrayArrayEnumColumn)(nil)
)

func NewArrayArrayEnumColumn() Columnar {
	return new(ArrayArrayEnumColumn)
}

func (c *ArrayArrayEnumColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayEnumColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][]string)(nil)).Elem()
}

func (c *ArrayArrayEnumColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayEnumColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayEnumColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayEnumColumn) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayEnumColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}

type DateTimeColumn struct {
	ColumnOf[time.Time]
}

var _ Columnar = (*DateTimeColumn)(nil)

func NewDateTimeColumn() Columnar {
	return new(DateTimeColumn)
}

var _DateTimeType = reflect.TypeOf((*time.Time)(nil)).Elem()

func (c *DateTimeColumn) Type() reflect.Type {
	return _DateTimeType
}

func (c *DateTimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.DateTime()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *DateTimeColumn) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.DateTime(n)
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayDateTimeColumn struct {
	ColumnOf[[]time.Time]
	elem DateTimeColumn
}

var (
	_ Columnar      = (*ArrayDateTimeColumn)(nil)
	_ ArrayColumnar = (*ArrayDateTimeColumn)(nil)
)

func NewArrayDateTimeColumn() Columnar {
	return new(ArrayDateTimeColumn)
}

func (c *ArrayDateTimeColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chType))
}

func (c *ArrayDateTimeColumn) Type() reflect.Type {
	return reflect.TypeOf((*[]time.Time)(nil)).Elem()
}

func (c *ArrayDateTimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayDateTimeColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayDateTimeColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayDateTimeColumn) WriteOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}

	return offset
}

func (c *ArrayDateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteTo(wr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type ArrayArrayDateTimeColumn struct {
	ColumnOf[[][]time.Time]
	elem ArrayDateTimeColumn
}

var (
	_ Columnar      = (*ArrayArrayDateTimeColumn)(nil)
	_ ArrayColumnar = (*ArrayArrayDateTimeColumn)(nil)
)

func NewArrayArrayDateTimeColumn() Columnar {
	return new(ArrayArrayDateTimeColumn)
}

func (c *ArrayArrayDateTimeColumn) Init(chType string) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)))
}

func (c *ArrayArrayDateTimeColumn) Type() reflect.Type {
	return reflect.TypeOf((*[][]time.Time)(nil)).Elem()
}

func (c *ArrayArrayDateTimeColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
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

func (c *ArrayArrayDateTimeColumn) readOffsets(rd *chproto.Reader, numRow int) ([]int, error) {
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

func (c *ArrayArrayDateTimeColumn) WriteTo(wr *chproto.Writer) error {
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayArrayDateTimeColumn) WriteOffset(wr *chproto.Writer, offset int) int {
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

func (c *ArrayArrayDateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.elem.Column = ss
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
