package chschema

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"github.com/uptrace/pkg/unixtime"
)

type Int8Column struct{ NumericColumnOf[int8] }

var _ Columnar = (*Int8Column)(nil)

func NewInt8Column() Columnar { return new(Int8Column) }

var _Int8Type = reflect.TypeFor[int8]()

func (c *Int8Column) Type() reflect.Type { return _Int8Type }

type NullableInt8Column struct {
	ColumnOf[*int8]
	Nulls UInt8Column
	Data  Int8Column
}

var _ Columnar = (*NullableInt8Column)(nil)

func NewNullableInt8Column() Columnar { return &NullableInt8Column{} }
func (c *NullableInt8Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableInt8Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableInt8Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableInt8Type = reflect.TypeFor[*int8]()

func (c *NullableInt8Column) Type() reflect.Type { return nullableInt8Type }
func (c *NullableInt8Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableInt8Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero int8
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayInt8Column struct {
	ColumnOf[[]int8]
	elem Int8Column
}

var _ Columnar = (*ArrayInt8Column)(nil)

func NewArrayInt8Column() Columnar { return new(ArrayInt8Column) }
func (c *ArrayInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayInt8Type = reflect.TypeFor[[]int8]()

func (c *ArrayInt8Column) Type() reflect.Type                  { return arrayInt8Type }
func (c *ArrayInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayInt8Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayInt8Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayInt8Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableInt8Column struct {
	ColumnOf[[]*int8]
	elem NullableInt8Column
}

var _ Columnar = (*ArrayNullableInt8Column)(nil)

func NewArrayNullableInt8Column() Columnar { return new(ArrayNullableInt8Column) }
func (c *ArrayNullableInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableInt8Type = reflect.TypeFor[[]*int8]()

func (c *ArrayNullableInt8Column) Type() reflect.Type                  { return arrayNullableInt8Type }
func (c *ArrayNullableInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableInt8Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableInt8Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayInt8Column struct {
	ColumnOf[[][]int8]
	elem ArrayInt8Column
}

var _ Columnar = (*ArrayArrayInt8Column)(nil)

func NewArrayArrayInt8Column() Columnar { return new(ArrayArrayInt8Column) }
func (c *ArrayArrayInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayInt8Type = reflect.TypeFor[[][]int8]()

func (c *ArrayArrayInt8Column) Type() reflect.Type                  { return arrayArrayInt8Type }
func (c *ArrayArrayInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayInt8Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayInt8Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type UInt8Column struct{ NumericColumnOf[uint8] }

var _ Columnar = (*UInt8Column)(nil)

func NewUInt8Column() Columnar { return new(UInt8Column) }

var _UInt8Type = reflect.TypeFor[uint8]()

func (c *UInt8Column) Type() reflect.Type { return _UInt8Type }

type NullableUInt8Column struct {
	ColumnOf[*uint8]
	Nulls UInt8Column
	Data  UInt8Column
}

var _ Columnar = (*NullableUInt8Column)(nil)

func NewNullableUInt8Column() Columnar { return &NullableUInt8Column{} }
func (c *NullableUInt8Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableUInt8Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableUInt8Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableUInt8Type = reflect.TypeFor[*uint8]()

func (c *NullableUInt8Column) Type() reflect.Type { return nullableUInt8Type }
func (c *NullableUInt8Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableUInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableUInt8Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero uint8
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayUInt8Column struct {
	ColumnOf[[]uint8]
	elem UInt8Column
}

var _ Columnar = (*ArrayUInt8Column)(nil)

func NewArrayUInt8Column() Columnar { return new(ArrayUInt8Column) }
func (c *ArrayUInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayUInt8Type = reflect.TypeFor[[]uint8]()

func (c *ArrayUInt8Column) Type() reflect.Type                  { return arrayUInt8Type }
func (c *ArrayUInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayUInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayUInt8Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayUInt8Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayUInt8Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableUInt8Column struct {
	ColumnOf[[]*uint8]
	elem NullableUInt8Column
}

var _ Columnar = (*ArrayNullableUInt8Column)(nil)

func NewArrayNullableUInt8Column() Columnar { return new(ArrayNullableUInt8Column) }
func (c *ArrayNullableUInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableUInt8Type = reflect.TypeFor[[]*uint8]()

func (c *ArrayNullableUInt8Column) Type() reflect.Type                  { return arrayNullableUInt8Type }
func (c *ArrayNullableUInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableUInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableUInt8Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableUInt8Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayUInt8Column struct {
	ColumnOf[[][]uint8]
	elem ArrayUInt8Column
}

var _ Columnar = (*ArrayArrayUInt8Column)(nil)

func NewArrayArrayUInt8Column() Columnar { return new(ArrayArrayUInt8Column) }
func (c *ArrayArrayUInt8Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayUInt8Type = reflect.TypeFor[[][]uint8]()

func (c *ArrayArrayUInt8Column) Type() reflect.Type                  { return arrayArrayUInt8Type }
func (c *ArrayArrayUInt8Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayUInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayUInt8Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayUInt8Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type Int16Column struct{ NumericColumnOf[int16] }

var _ Columnar = (*Int16Column)(nil)

func NewInt16Column() Columnar { return new(Int16Column) }

var _Int16Type = reflect.TypeFor[int16]()

func (c *Int16Column) Type() reflect.Type { return _Int16Type }

type NullableInt16Column struct {
	ColumnOf[*int16]
	Nulls UInt8Column
	Data  Int16Column
}

var _ Columnar = (*NullableInt16Column)(nil)

func NewNullableInt16Column() Columnar { return &NullableInt16Column{} }
func (c *NullableInt16Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableInt16Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableInt16Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableInt16Type = reflect.TypeFor[*int16]()

func (c *NullableInt16Column) Type() reflect.Type { return nullableInt16Type }
func (c *NullableInt16Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableInt16Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero int16
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayInt16Column struct {
	ColumnOf[[]int16]
	elem Int16Column
}

var _ Columnar = (*ArrayInt16Column)(nil)

func NewArrayInt16Column() Columnar { return new(ArrayInt16Column) }
func (c *ArrayInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayInt16Type = reflect.TypeFor[[]int16]()

func (c *ArrayInt16Column) Type() reflect.Type                  { return arrayInt16Type }
func (c *ArrayInt16Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayInt16Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayInt16Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayInt16Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableInt16Column struct {
	ColumnOf[[]*int16]
	elem NullableInt16Column
}

var _ Columnar = (*ArrayNullableInt16Column)(nil)

func NewArrayNullableInt16Column() Columnar { return new(ArrayNullableInt16Column) }
func (c *ArrayNullableInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableInt16Type = reflect.TypeFor[[]*int16]()

func (c *ArrayNullableInt16Column) Type() reflect.Type                  { return arrayNullableInt16Type }
func (c *ArrayNullableInt16Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableInt16Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableInt16Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayInt16Column struct {
	ColumnOf[[][]int16]
	elem ArrayInt16Column
}

var _ Columnar = (*ArrayArrayInt16Column)(nil)

func NewArrayArrayInt16Column() Columnar { return new(ArrayArrayInt16Column) }
func (c *ArrayArrayInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayInt16Type = reflect.TypeFor[[][]int16]()

func (c *ArrayArrayInt16Column) Type() reflect.Type                  { return arrayArrayInt16Type }
func (c *ArrayArrayInt16Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayInt16Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayInt16Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type UInt16Column struct{ NumericColumnOf[uint16] }

var _ Columnar = (*UInt16Column)(nil)

func NewUInt16Column() Columnar { return new(UInt16Column) }

var _UInt16Type = reflect.TypeFor[uint16]()

func (c *UInt16Column) Type() reflect.Type { return _UInt16Type }

type NullableUInt16Column struct {
	ColumnOf[*uint16]
	Nulls UInt8Column
	Data  UInt16Column
}

var _ Columnar = (*NullableUInt16Column)(nil)

func NewNullableUInt16Column() Columnar { return &NullableUInt16Column{} }
func (c *NullableUInt16Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableUInt16Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableUInt16Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableUInt16Type = reflect.TypeFor[*uint16]()

func (c *NullableUInt16Column) Type() reflect.Type { return nullableUInt16Type }
func (c *NullableUInt16Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableUInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableUInt16Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero uint16
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayUInt16Column struct {
	ColumnOf[[]uint16]
	elem UInt16Column
}

var _ Columnar = (*ArrayUInt16Column)(nil)

func NewArrayUInt16Column() Columnar { return new(ArrayUInt16Column) }
func (c *ArrayUInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayUInt16Type = reflect.TypeFor[[]uint16]()

func (c *ArrayUInt16Column) Type() reflect.Type                  { return arrayUInt16Type }
func (c *ArrayUInt16Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayUInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayUInt16Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayUInt16Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayUInt16Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableUInt16Column struct {
	ColumnOf[[]*uint16]
	elem NullableUInt16Column
}

var _ Columnar = (*ArrayNullableUInt16Column)(nil)

func NewArrayNullableUInt16Column() Columnar { return new(ArrayNullableUInt16Column) }
func (c *ArrayNullableUInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableUInt16Type = reflect.TypeFor[[]*uint16]()

func (c *ArrayNullableUInt16Column) Type() reflect.Type { return arrayNullableUInt16Type }
func (c *ArrayNullableUInt16Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableUInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableUInt16Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableUInt16Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayUInt16Column struct {
	ColumnOf[[][]uint16]
	elem ArrayUInt16Column
}

var _ Columnar = (*ArrayArrayUInt16Column)(nil)

func NewArrayArrayUInt16Column() Columnar { return new(ArrayArrayUInt16Column) }
func (c *ArrayArrayUInt16Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayUInt16Type = reflect.TypeFor[[][]uint16]()

func (c *ArrayArrayUInt16Column) Type() reflect.Type                  { return arrayArrayUInt16Type }
func (c *ArrayArrayUInt16Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayUInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayUInt16Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayUInt16Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type Int32Column struct{ NumericColumnOf[int32] }

var _ Columnar = (*Int32Column)(nil)

func NewInt32Column() Columnar { return new(Int32Column) }

var _Int32Type = reflect.TypeFor[int32]()

func (c *Int32Column) Type() reflect.Type { return _Int32Type }

type NullableInt32Column struct {
	ColumnOf[*int32]
	Nulls UInt8Column
	Data  Int32Column
}

var _ Columnar = (*NullableInt32Column)(nil)

func NewNullableInt32Column() Columnar { return &NullableInt32Column{} }
func (c *NullableInt32Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableInt32Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableInt32Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableInt32Type = reflect.TypeFor[*int32]()

func (c *NullableInt32Column) Type() reflect.Type { return nullableInt32Type }
func (c *NullableInt32Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableInt32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero int32
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayInt32Column struct {
	ColumnOf[[]int32]
	elem Int32Column
}

var _ Columnar = (*ArrayInt32Column)(nil)

func NewArrayInt32Column() Columnar { return new(ArrayInt32Column) }
func (c *ArrayInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayInt32Type = reflect.TypeFor[[]int32]()

func (c *ArrayInt32Column) Type() reflect.Type                  { return arrayInt32Type }
func (c *ArrayInt32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayInt32Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayInt32Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayInt32Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableInt32Column struct {
	ColumnOf[[]*int32]
	elem NullableInt32Column
}

var _ Columnar = (*ArrayNullableInt32Column)(nil)

func NewArrayNullableInt32Column() Columnar { return new(ArrayNullableInt32Column) }
func (c *ArrayNullableInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableInt32Type = reflect.TypeFor[[]*int32]()

func (c *ArrayNullableInt32Column) Type() reflect.Type                  { return arrayNullableInt32Type }
func (c *ArrayNullableInt32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableInt32Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableInt32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayInt32Column struct {
	ColumnOf[[][]int32]
	elem ArrayInt32Column
}

var _ Columnar = (*ArrayArrayInt32Column)(nil)

func NewArrayArrayInt32Column() Columnar { return new(ArrayArrayInt32Column) }
func (c *ArrayArrayInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayInt32Type = reflect.TypeFor[[][]int32]()

func (c *ArrayArrayInt32Column) Type() reflect.Type                  { return arrayArrayInt32Type }
func (c *ArrayArrayInt32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayInt32Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayInt32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type UInt32Column struct{ NumericColumnOf[uint32] }

var _ Columnar = (*UInt32Column)(nil)

func NewUInt32Column() Columnar { return new(UInt32Column) }

var _UInt32Type = reflect.TypeFor[uint32]()

func (c *UInt32Column) Type() reflect.Type { return _UInt32Type }

type NullableUInt32Column struct {
	ColumnOf[*uint32]
	Nulls UInt8Column
	Data  UInt32Column
}

var _ Columnar = (*NullableUInt32Column)(nil)

func NewNullableUInt32Column() Columnar { return &NullableUInt32Column{} }
func (c *NullableUInt32Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableUInt32Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableUInt32Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableUInt32Type = reflect.TypeFor[*uint32]()

func (c *NullableUInt32Column) Type() reflect.Type { return nullableUInt32Type }
func (c *NullableUInt32Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableUInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableUInt32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero uint32
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayUInt32Column struct {
	ColumnOf[[]uint32]
	elem UInt32Column
}

var _ Columnar = (*ArrayUInt32Column)(nil)

func NewArrayUInt32Column() Columnar { return new(ArrayUInt32Column) }
func (c *ArrayUInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayUInt32Type = reflect.TypeFor[[]uint32]()

func (c *ArrayUInt32Column) Type() reflect.Type                  { return arrayUInt32Type }
func (c *ArrayUInt32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayUInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayUInt32Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayUInt32Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayUInt32Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableUInt32Column struct {
	ColumnOf[[]*uint32]
	elem NullableUInt32Column
}

var _ Columnar = (*ArrayNullableUInt32Column)(nil)

func NewArrayNullableUInt32Column() Columnar { return new(ArrayNullableUInt32Column) }
func (c *ArrayNullableUInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableUInt32Type = reflect.TypeFor[[]*uint32]()

func (c *ArrayNullableUInt32Column) Type() reflect.Type { return arrayNullableUInt32Type }
func (c *ArrayNullableUInt32Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableUInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableUInt32Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableUInt32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayUInt32Column struct {
	ColumnOf[[][]uint32]
	elem ArrayUInt32Column
}

var _ Columnar = (*ArrayArrayUInt32Column)(nil)

func NewArrayArrayUInt32Column() Columnar { return new(ArrayArrayUInt32Column) }
func (c *ArrayArrayUInt32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayUInt32Type = reflect.TypeFor[[][]uint32]()

func (c *ArrayArrayUInt32Column) Type() reflect.Type                  { return arrayArrayUInt32Type }
func (c *ArrayArrayUInt32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayUInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayUInt32Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayUInt32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type Int64Column struct{ NumericColumnOf[int64] }

var _ Columnar = (*Int64Column)(nil)

func NewInt64Column() Columnar { return new(Int64Column) }

var _Int64Type = reflect.TypeFor[int64]()

func (c *Int64Column) Type() reflect.Type { return _Int64Type }

type NullableInt64Column struct {
	ColumnOf[*int64]
	Nulls UInt8Column
	Data  Int64Column
}

var _ Columnar = (*NullableInt64Column)(nil)

func NewNullableInt64Column() Columnar { return &NullableInt64Column{} }
func (c *NullableInt64Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableInt64Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableInt64Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableInt64Type = reflect.TypeFor[*int64]()

func (c *NullableInt64Column) Type() reflect.Type { return nullableInt64Type }
func (c *NullableInt64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableInt64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero int64
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayInt64Column struct {
	ColumnOf[[]int64]
	elem Int64Column
}

var _ Columnar = (*ArrayInt64Column)(nil)

func NewArrayInt64Column() Columnar { return new(ArrayInt64Column) }
func (c *ArrayInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayInt64Type = reflect.TypeFor[[]int64]()

func (c *ArrayInt64Column) Type() reflect.Type                  { return arrayInt64Type }
func (c *ArrayInt64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayInt64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayInt64Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayInt64Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableInt64Column struct {
	ColumnOf[[]*int64]
	elem NullableInt64Column
}

var _ Columnar = (*ArrayNullableInt64Column)(nil)

func NewArrayNullableInt64Column() Columnar { return new(ArrayNullableInt64Column) }
func (c *ArrayNullableInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableInt64Type = reflect.TypeFor[[]*int64]()

func (c *ArrayNullableInt64Column) Type() reflect.Type                  { return arrayNullableInt64Type }
func (c *ArrayNullableInt64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableInt64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableInt64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayInt64Column struct {
	ColumnOf[[][]int64]
	elem ArrayInt64Column
}

var _ Columnar = (*ArrayArrayInt64Column)(nil)

func NewArrayArrayInt64Column() Columnar { return new(ArrayArrayInt64Column) }
func (c *ArrayArrayInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayInt64Type = reflect.TypeFor[[][]int64]()

func (c *ArrayArrayInt64Column) Type() reflect.Type                  { return arrayArrayInt64Type }
func (c *ArrayArrayInt64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayInt64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayInt64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type UInt64Column struct{ NumericColumnOf[uint64] }

var _ Columnar = (*UInt64Column)(nil)

func NewUInt64Column() Columnar { return new(UInt64Column) }

var _UInt64Type = reflect.TypeFor[uint64]()

func (c *UInt64Column) Type() reflect.Type { return _UInt64Type }

type NullableUInt64Column struct {
	ColumnOf[*uint64]
	Nulls UInt8Column
	Data  UInt64Column
}

var _ Columnar = (*NullableUInt64Column)(nil)

func NewNullableUInt64Column() Columnar { return &NullableUInt64Column{} }
func (c *NullableUInt64Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableUInt64Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableUInt64Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableUInt64Type = reflect.TypeFor[*uint64]()

func (c *NullableUInt64Column) Type() reflect.Type { return nullableUInt64Type }
func (c *NullableUInt64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableUInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableUInt64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero uint64
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayUInt64Column struct {
	ColumnOf[[]uint64]
	elem UInt64Column
}

var _ Columnar = (*ArrayUInt64Column)(nil)

func NewArrayUInt64Column() Columnar { return new(ArrayUInt64Column) }
func (c *ArrayUInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayUInt64Type = reflect.TypeFor[[]uint64]()

func (c *ArrayUInt64Column) Type() reflect.Type                  { return arrayUInt64Type }
func (c *ArrayUInt64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayUInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayUInt64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayUInt64Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayUInt64Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableUInt64Column struct {
	ColumnOf[[]*uint64]
	elem NullableUInt64Column
}

var _ Columnar = (*ArrayNullableUInt64Column)(nil)

func NewArrayNullableUInt64Column() Columnar { return new(ArrayNullableUInt64Column) }
func (c *ArrayNullableUInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableUInt64Type = reflect.TypeFor[[]*uint64]()

func (c *ArrayNullableUInt64Column) Type() reflect.Type { return arrayNullableUInt64Type }
func (c *ArrayNullableUInt64Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableUInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableUInt64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableUInt64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayUInt64Column struct {
	ColumnOf[[][]uint64]
	elem ArrayUInt64Column
}

var _ Columnar = (*ArrayArrayUInt64Column)(nil)

func NewArrayArrayUInt64Column() Columnar { return new(ArrayArrayUInt64Column) }
func (c *ArrayArrayUInt64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayUInt64Type = reflect.TypeFor[[][]uint64]()

func (c *ArrayArrayUInt64Column) Type() reflect.Type                  { return arrayArrayUInt64Type }
func (c *ArrayArrayUInt64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayUInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayUInt64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayUInt64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type Float32Column struct{ NumericColumnOf[float32] }

var _ Columnar = (*Float32Column)(nil)

func NewFloat32Column() Columnar { return new(Float32Column) }

var _Float32Type = reflect.TypeFor[float32]()

func (c *Float32Column) Type() reflect.Type { return _Float32Type }

type NullableFloat32Column struct {
	ColumnOf[*float32]
	Nulls UInt8Column
	Data  Float32Column
}

var _ Columnar = (*NullableFloat32Column)(nil)

func NewNullableFloat32Column() Columnar { return &NullableFloat32Column{} }
func (c *NullableFloat32Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableFloat32Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableFloat32Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableFloat32Type = reflect.TypeFor[*float32]()

func (c *NullableFloat32Column) Type() reflect.Type { return nullableFloat32Type }
func (c *NullableFloat32Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableFloat32Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableFloat32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero float32
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayFloat32Column struct {
	ColumnOf[[]float32]
	elem Float32Column
}

var _ Columnar = (*ArrayFloat32Column)(nil)

func NewArrayFloat32Column() Columnar { return new(ArrayFloat32Column) }
func (c *ArrayFloat32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayFloat32Type = reflect.TypeFor[[]float32]()

func (c *ArrayFloat32Column) Type() reflect.Type                  { return arrayFloat32Type }
func (c *ArrayFloat32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayFloat32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayFloat32Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayFloat32Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayFloat32Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableFloat32Column struct {
	ColumnOf[[]*float32]
	elem NullableFloat32Column
}

var _ Columnar = (*ArrayNullableFloat32Column)(nil)

func NewArrayNullableFloat32Column() Columnar { return new(ArrayNullableFloat32Column) }
func (c *ArrayNullableFloat32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableFloat32Type = reflect.TypeFor[[]*float32]()

func (c *ArrayNullableFloat32Column) Type() reflect.Type { return arrayNullableFloat32Type }
func (c *ArrayNullableFloat32Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableFloat32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableFloat32Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableFloat32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayFloat32Column struct {
	ColumnOf[[][]float32]
	elem ArrayFloat32Column
}

var _ Columnar = (*ArrayArrayFloat32Column)(nil)

func NewArrayArrayFloat32Column() Columnar { return new(ArrayArrayFloat32Column) }
func (c *ArrayArrayFloat32Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayFloat32Type = reflect.TypeFor[[][]float32]()

func (c *ArrayArrayFloat32Column) Type() reflect.Type                  { return arrayArrayFloat32Type }
func (c *ArrayArrayFloat32Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayFloat32Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayFloat32Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayFloat32Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type Float64Column struct{ NumericColumnOf[float64] }

var _ Columnar = (*Float64Column)(nil)

func NewFloat64Column() Columnar { return new(Float64Column) }

var _Float64Type = reflect.TypeFor[float64]()

func (c *Float64Column) Type() reflect.Type { return _Float64Type }

type NullableFloat64Column struct {
	ColumnOf[*float64]
	Nulls UInt8Column
	Data  Float64Column
}

var _ Columnar = (*NullableFloat64Column)(nil)

func NewNullableFloat64Column() Columnar { return &NullableFloat64Column{} }
func (c *NullableFloat64Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableFloat64Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableFloat64Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableFloat64Type = reflect.TypeFor[*float64]()

func (c *NullableFloat64Column) Type() reflect.Type { return nullableFloat64Type }
func (c *NullableFloat64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableFloat64Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableFloat64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero float64
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayFloat64Column struct {
	ColumnOf[[]float64]
	elem Float64Column
}

var _ Columnar = (*ArrayFloat64Column)(nil)

func NewArrayFloat64Column() Columnar { return new(ArrayFloat64Column) }
func (c *ArrayFloat64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayFloat64Type = reflect.TypeFor[[]float64]()

func (c *ArrayFloat64Column) Type() reflect.Type                  { return arrayFloat64Type }
func (c *ArrayFloat64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayFloat64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayFloat64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayFloat64Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayFloat64Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableFloat64Column struct {
	ColumnOf[[]*float64]
	elem NullableFloat64Column
}

var _ Columnar = (*ArrayNullableFloat64Column)(nil)

func NewArrayNullableFloat64Column() Columnar { return new(ArrayNullableFloat64Column) }
func (c *ArrayNullableFloat64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableFloat64Type = reflect.TypeFor[[]*float64]()

func (c *ArrayNullableFloat64Column) Type() reflect.Type { return arrayNullableFloat64Type }
func (c *ArrayNullableFloat64Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableFloat64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableFloat64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableFloat64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayFloat64Column struct {
	ColumnOf[[][]float64]
	elem ArrayFloat64Column
}

var _ Columnar = (*ArrayArrayFloat64Column)(nil)

func NewArrayArrayFloat64Column() Columnar { return new(ArrayArrayFloat64Column) }
func (c *ArrayArrayFloat64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayFloat64Type = reflect.TypeFor[[][]float64]()

func (c *ArrayArrayFloat64Column) Type() reflect.Type                  { return arrayArrayFloat64Type }
func (c *ArrayArrayFloat64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayFloat64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayFloat64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayFloat64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type BoolColumn struct{ ColumnOf[bool] }

var _ Columnar = (*BoolColumn)(nil)

func NewBoolColumn() Columnar { return new(BoolColumn) }

var _BoolType = reflect.TypeFor[bool]()

func (c *BoolColumn) Type() reflect.Type { return _BoolType }
func (c *BoolColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.Bool()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *BoolColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Bool(n)
	}
	return nil
}

type NullableBoolColumn struct {
	ColumnOf[*bool]
	Nulls UInt8Column
	Data  BoolColumn
}

var _ Columnar = (*NullableBoolColumn)(nil)

func NewNullableBoolColumn() Columnar { return &NullableBoolColumn{} }
func (c *NullableBoolColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableBoolColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableBoolColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableBoolType = reflect.TypeFor[*bool]()

func (c *NullableBoolColumn) Type() reflect.Type { return nullableBoolType }
func (c *NullableBoolColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableBoolColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableBoolColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero bool
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayBoolColumn struct {
	ColumnOf[[]bool]
	elem BoolColumn
}

var _ Columnar = (*ArrayBoolColumn)(nil)

func NewArrayBoolColumn() Columnar { return new(ArrayBoolColumn) }
func (c *ArrayBoolColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayBoolType = reflect.TypeFor[[]bool]()

func (c *ArrayBoolColumn) Type() reflect.Type                  { return arrayBoolType }
func (c *ArrayBoolColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayBoolColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayBoolColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayBoolColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayBoolColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableBoolColumn struct {
	ColumnOf[[]*bool]
	elem NullableBoolColumn
}

var _ Columnar = (*ArrayNullableBoolColumn)(nil)

func NewArrayNullableBoolColumn() Columnar { return new(ArrayNullableBoolColumn) }
func (c *ArrayNullableBoolColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableBoolType = reflect.TypeFor[[]*bool]()

func (c *ArrayNullableBoolColumn) Type() reflect.Type                  { return arrayNullableBoolType }
func (c *ArrayNullableBoolColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableBoolColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableBoolColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableBoolColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayBoolColumn struct {
	ColumnOf[[][]bool]
	elem ArrayBoolColumn
}

var _ Columnar = (*ArrayArrayBoolColumn)(nil)

func NewArrayArrayBoolColumn() Columnar { return new(ArrayArrayBoolColumn) }
func (c *ArrayArrayBoolColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayBoolType = reflect.TypeFor[[][]bool]()

func (c *ArrayArrayBoolColumn) Type() reflect.Type                  { return arrayArrayBoolType }
func (c *ArrayArrayBoolColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayBoolColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayBoolColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayBoolColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type StringColumn struct{ ColumnOf[string] }

var _ Columnar = (*StringColumn)(nil)

func NewStringColumn() Columnar { return new(StringColumn) }

var _StringType = reflect.TypeFor[string]()

func (c *StringColumn) Type() reflect.Type { return _StringType }
func (c *StringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.String()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *StringColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.String(n)
	}
	return nil
}

type NullableStringColumn struct {
	ColumnOf[*string]
	Nulls UInt8Column
	Data  StringColumn
}

var _ Columnar = (*NullableStringColumn)(nil)

func NewNullableStringColumn() Columnar { return &NullableStringColumn{} }
func (c *NullableStringColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableStringColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableStringColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableStringType = reflect.TypeFor[*string]()

func (c *NullableStringColumn) Type() reflect.Type { return nullableStringType }
func (c *NullableStringColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableStringColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero string
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayStringColumn struct {
	ColumnOf[[]string]
	elem StringColumn
}

var _ Columnar = (*ArrayStringColumn)(nil)

func NewArrayStringColumn() Columnar { return new(ArrayStringColumn) }
func (c *ArrayStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayStringType = reflect.TypeFor[[]string]()

func (c *ArrayStringColumn) Type() reflect.Type                  { return arrayStringType }
func (c *ArrayStringColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayStringColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayStringColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayStringColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableStringColumn struct {
	ColumnOf[[]*string]
	elem NullableStringColumn
}

var _ Columnar = (*ArrayNullableStringColumn)(nil)

func NewArrayNullableStringColumn() Columnar { return new(ArrayNullableStringColumn) }
func (c *ArrayNullableStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableStringType = reflect.TypeFor[[]*string]()

func (c *ArrayNullableStringColumn) Type() reflect.Type { return arrayNullableStringType }
func (c *ArrayNullableStringColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableStringColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableStringColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayStringColumn struct {
	ColumnOf[[][]string]
	elem ArrayStringColumn
}

var _ Columnar = (*ArrayArrayStringColumn)(nil)

func NewArrayArrayStringColumn() Columnar { return new(ArrayArrayStringColumn) }
func (c *ArrayArrayStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayStringType = reflect.TypeFor[[][]string]()

func (c *ArrayArrayStringColumn) Type() reflect.Type                  { return arrayArrayStringType }
func (c *ArrayArrayStringColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayStringColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayStringColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type BytesColumn struct{ ColumnOf[[]byte] }

var _ Columnar = (*BytesColumn)(nil)

func NewBytesColumn() Columnar { return new(BytesColumn) }

var _BytesType = reflect.TypeFor[[]byte]()

func (c *BytesColumn) Type() reflect.Type { return _BytesType }
func (c *BytesColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.Bytes()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *BytesColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Bytes(n)
	}
	return nil
}

type NullableBytesColumn struct {
	ColumnOf[*[]byte]
	Nulls UInt8Column
	Data  BytesColumn
}

var _ Columnar = (*NullableBytesColumn)(nil)

func NewNullableBytesColumn() Columnar { return &NullableBytesColumn{} }
func (c *NullableBytesColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableBytesColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableBytesColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableBytesType = reflect.TypeFor[*[]byte]()

func (c *NullableBytesColumn) Type() reflect.Type { return nullableBytesType }
func (c *NullableBytesColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableBytesColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableBytesColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero []byte
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayBytesColumn struct {
	ColumnOf[[][]byte]
	elem BytesColumn
}

var _ Columnar = (*ArrayBytesColumn)(nil)

func NewArrayBytesColumn() Columnar { return new(ArrayBytesColumn) }
func (c *ArrayBytesColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayBytesType = reflect.TypeFor[[][]byte]()

func (c *ArrayBytesColumn) Type() reflect.Type                  { return arrayBytesType }
func (c *ArrayBytesColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayBytesColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayBytesColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayBytesColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayBytesColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableBytesColumn struct {
	ColumnOf[[]*[]byte]
	elem NullableBytesColumn
}

var _ Columnar = (*ArrayNullableBytesColumn)(nil)

func NewArrayNullableBytesColumn() Columnar { return new(ArrayNullableBytesColumn) }
func (c *ArrayNullableBytesColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableBytesType = reflect.TypeFor[[]*[]byte]()

func (c *ArrayNullableBytesColumn) Type() reflect.Type                  { return arrayNullableBytesType }
func (c *ArrayNullableBytesColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableBytesColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableBytesColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableBytesColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayBytesColumn struct {
	ColumnOf[[][][]byte]
	elem ArrayBytesColumn
}

var _ Columnar = (*ArrayArrayBytesColumn)(nil)

func NewArrayArrayBytesColumn() Columnar { return new(ArrayArrayBytesColumn) }
func (c *ArrayArrayBytesColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayBytesType = reflect.TypeFor[[][][]byte]()

func (c *ArrayArrayBytesColumn) Type() reflect.Type                  { return arrayArrayBytesType }
func (c *ArrayArrayBytesColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayBytesColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayBytesColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayBytesColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type NullableEnumColumn struct {
	ColumnOf[*string]
	Nulls UInt8Column
	Data  EnumColumn
}

var _ Columnar = (*NullableEnumColumn)(nil)

func NewNullableEnumColumn() Columnar { return &NullableEnumColumn{} }
func (c *NullableEnumColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableEnumColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableEnumColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableEnumType = reflect.TypeFor[*string]()

func (c *NullableEnumColumn) Type() reflect.Type { return nullableEnumType }
func (c *NullableEnumColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableEnumColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableEnumColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero string
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayEnumColumn struct {
	ColumnOf[[]string]
	elem EnumColumn
}

var _ Columnar = (*ArrayEnumColumn)(nil)

func NewArrayEnumColumn() Columnar { return new(ArrayEnumColumn) }
func (c *ArrayEnumColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayEnumType = reflect.TypeFor[[]string]()

func (c *ArrayEnumColumn) Type() reflect.Type                  { return arrayEnumType }
func (c *ArrayEnumColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayEnumColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayEnumColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayEnumColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayEnumColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableEnumColumn struct {
	ColumnOf[[]*string]
	elem NullableEnumColumn
}

var _ Columnar = (*ArrayNullableEnumColumn)(nil)

func NewArrayNullableEnumColumn() Columnar { return new(ArrayNullableEnumColumn) }
func (c *ArrayNullableEnumColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableEnumType = reflect.TypeFor[[]*string]()

func (c *ArrayNullableEnumColumn) Type() reflect.Type                  { return arrayNullableEnumType }
func (c *ArrayNullableEnumColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableEnumColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableEnumColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableEnumColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayEnumColumn struct {
	ColumnOf[[][]string]
	elem ArrayEnumColumn
}

var _ Columnar = (*ArrayArrayEnumColumn)(nil)

func NewArrayArrayEnumColumn() Columnar { return new(ArrayArrayEnumColumn) }
func (c *ArrayArrayEnumColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayEnumType = reflect.TypeFor[[][]string]()

func (c *ArrayArrayEnumColumn) Type() reflect.Type                  { return arrayArrayEnumType }
func (c *ArrayArrayEnumColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayEnumColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayEnumColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayEnumColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type NullableLCStringColumn struct {
	ColumnOf[*string]
	Nulls UInt8Column
	Data  LCStringColumn
}

var _ Columnar = (*NullableLCStringColumn)(nil)

func NewNullableLCStringColumn() Columnar { return &NullableLCStringColumn{} }
func (c *NullableLCStringColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableLCStringColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableLCStringColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableLCStringType = reflect.TypeFor[*string]()

func (c *NullableLCStringColumn) Type() reflect.Type { return nullableLCStringType }
func (c *NullableLCStringColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableLCStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableLCStringColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero string
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayLCStringColumn struct {
	ColumnOf[[]string]
	elem LCStringColumn
}

var _ Columnar = (*ArrayLCStringColumn)(nil)

func NewArrayLCStringColumn() Columnar { return new(ArrayLCStringColumn) }
func (c *ArrayLCStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayLCStringType = reflect.TypeFor[[]string]()

func (c *ArrayLCStringColumn) Type() reflect.Type                  { return arrayLCStringType }
func (c *ArrayLCStringColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayLCStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayLCStringColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayLCStringColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayLCStringColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableLCStringColumn struct {
	ColumnOf[[]*string]
	elem NullableLCStringColumn
}

var _ Columnar = (*ArrayNullableLCStringColumn)(nil)

func NewArrayNullableLCStringColumn() Columnar { return new(ArrayNullableLCStringColumn) }
func (c *ArrayNullableLCStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableLCStringType = reflect.TypeFor[[]*string]()

func (c *ArrayNullableLCStringColumn) Type() reflect.Type { return arrayNullableLCStringType }
func (c *ArrayNullableLCStringColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableLCStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableLCStringColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableLCStringColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayLCStringColumn struct {
	ColumnOf[[][]string]
	elem ArrayLCStringColumn
}

var _ Columnar = (*ArrayArrayLCStringColumn)(nil)

func NewArrayArrayLCStringColumn() Columnar { return new(ArrayArrayLCStringColumn) }
func (c *ArrayArrayLCStringColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayLCStringType = reflect.TypeFor[[][]string]()

func (c *ArrayArrayLCStringColumn) Type() reflect.Type                  { return arrayArrayLCStringType }
func (c *ArrayArrayLCStringColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayLCStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayLCStringColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayLCStringColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type DateTimeColumn struct{ ColumnOf[unixtime.Nano] }

var _ Columnar = (*DateTimeColumn)(nil)

func NewDateTimeColumn() Columnar { return new(DateTimeColumn) }

var _DateTimeType = reflect.TypeFor[unixtime.Nano]()

func (c *DateTimeColumn) Type() reflect.Type { return _DateTimeType }
func (c *DateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.DateTime()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *DateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.DateTime(n)
	}
	return nil
}

type NullableDateTimeColumn struct {
	ColumnOf[*unixtime.Nano]
	Nulls UInt8Column
	Data  DateTimeColumn
}

var _ Columnar = (*NullableDateTimeColumn)(nil)

func NewNullableDateTimeColumn() Columnar { return &NullableDateTimeColumn{} }
func (c *NullableDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableDateTimeColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableDateTimeColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableDateTimeType = reflect.TypeFor[*unixtime.Nano]()

func (c *NullableDateTimeColumn) Type() reflect.Type { return nullableDateTimeType }
func (c *NullableDateTimeColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableDateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero unixtime.Nano
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayDateTimeColumn struct {
	ColumnOf[[]unixtime.Nano]
	elem DateTimeColumn
}

var _ Columnar = (*ArrayDateTimeColumn)(nil)

func NewArrayDateTimeColumn() Columnar { return new(ArrayDateTimeColumn) }
func (c *ArrayDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayDateTimeType = reflect.TypeFor[[]unixtime.Nano]()

func (c *ArrayDateTimeColumn) Type() reflect.Type                  { return arrayDateTimeType }
func (c *ArrayDateTimeColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayDateTimeColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayDateTimeColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayDateTimeColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableDateTimeColumn struct {
	ColumnOf[[]*unixtime.Nano]
	elem NullableDateTimeColumn
}

var _ Columnar = (*ArrayNullableDateTimeColumn)(nil)

func NewArrayNullableDateTimeColumn() Columnar { return new(ArrayNullableDateTimeColumn) }
func (c *ArrayNullableDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableDateTimeType = reflect.TypeFor[[]*unixtime.Nano]()

func (c *ArrayNullableDateTimeColumn) Type() reflect.Type { return arrayNullableDateTimeType }
func (c *ArrayNullableDateTimeColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableDateTimeColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableDateTimeColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayDateTimeColumn struct {
	ColumnOf[[][]unixtime.Nano]
	elem ArrayDateTimeColumn
}

var _ Columnar = (*ArrayArrayDateTimeColumn)(nil)

func NewArrayArrayDateTimeColumn() Columnar { return new(ArrayArrayDateTimeColumn) }
func (c *ArrayArrayDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayDateTimeType = reflect.TypeFor[[][]unixtime.Nano]()

func (c *ArrayArrayDateTimeColumn) Type() reflect.Type                  { return arrayArrayDateTimeType }
func (c *ArrayArrayDateTimeColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayDateTimeColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayDateTimeColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type GoDateTimeColumn struct{ ColumnOf[time.Time] }

var _ Columnar = (*GoDateTimeColumn)(nil)

func NewGoDateTimeColumn() Columnar { return new(GoDateTimeColumn) }

var _GoDateTimeType = reflect.TypeFor[time.Time]()

func (c *GoDateTimeColumn) Type() reflect.Type { return _GoDateTimeType }
func (c *GoDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.GoDateTime()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *GoDateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.GoDateTime(n)
	}
	return nil
}

type NullableGoDateTimeColumn struct {
	ColumnOf[*time.Time]
	Nulls UInt8Column
	Data  GoDateTimeColumn
}

var _ Columnar = (*NullableGoDateTimeColumn)(nil)

func NewNullableGoDateTimeColumn() Columnar { return &NullableGoDateTimeColumn{} }
func (c *NullableGoDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableGoDateTimeColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableGoDateTimeColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableGoDateTimeType = reflect.TypeFor[*time.Time]()

func (c *NullableGoDateTimeColumn) Type() reflect.Type { return nullableGoDateTimeType }
func (c *NullableGoDateTimeColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableGoDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableGoDateTimeColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero time.Time
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayGoDateTimeColumn struct {
	ColumnOf[[]time.Time]
	elem GoDateTimeColumn
}

var _ Columnar = (*ArrayGoDateTimeColumn)(nil)

func NewArrayGoDateTimeColumn() Columnar { return new(ArrayGoDateTimeColumn) }
func (c *ArrayGoDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayGoDateTimeType = reflect.TypeFor[[]time.Time]()

func (c *ArrayGoDateTimeColumn) Type() reflect.Type                  { return arrayGoDateTimeType }
func (c *ArrayGoDateTimeColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayGoDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayGoDateTimeColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayGoDateTimeColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayGoDateTimeColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableGoDateTimeColumn struct {
	ColumnOf[[]*time.Time]
	elem NullableGoDateTimeColumn
}

var _ Columnar = (*ArrayNullableGoDateTimeColumn)(nil)

func NewArrayNullableGoDateTimeColumn() Columnar { return new(ArrayNullableGoDateTimeColumn) }
func (c *ArrayNullableGoDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableGoDateTimeType = reflect.TypeFor[[]*time.Time]()

func (c *ArrayNullableGoDateTimeColumn) Type() reflect.Type { return arrayNullableGoDateTimeType }
func (c *ArrayNullableGoDateTimeColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableGoDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableGoDateTimeColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableGoDateTimeColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayGoDateTimeColumn struct {
	ColumnOf[[][]time.Time]
	elem ArrayGoDateTimeColumn
}

var _ Columnar = (*ArrayArrayGoDateTimeColumn)(nil)

func NewArrayArrayGoDateTimeColumn() Columnar { return new(ArrayArrayGoDateTimeColumn) }
func (c *ArrayArrayGoDateTimeColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayGoDateTimeType = reflect.TypeFor[[][]time.Time]()

func (c *ArrayArrayGoDateTimeColumn) Type() reflect.Type { return arrayArrayGoDateTimeType }
func (c *ArrayArrayGoDateTimeColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayArrayGoDateTimeColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayGoDateTimeColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayGoDateTimeColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type DateColumn struct{ ColumnOf[unixtime.Nano] }

var _ Columnar = (*DateColumn)(nil)

func NewDateColumn() Columnar { return new(DateColumn) }

var _DateType = reflect.TypeFor[unixtime.Nano]()

func (c *DateColumn) Type() reflect.Type { return _DateType }
func (c *DateColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		val, err := rd.Date()
		if err != nil {
			return err
		}
		c.Column[i] = val
	}
	return nil
}
func (c *DateColumn) WriteData(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Date(n)
	}
	return nil
}

type NullableDateColumn struct {
	ColumnOf[*unixtime.Nano]
	Nulls UInt8Column
	Data  DateColumn
}

var _ Columnar = (*NullableDateColumn)(nil)

func NewNullableDateColumn() Columnar { return &NullableDateColumn{} }
func (c *NullableDateColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableDateColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableDateColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableDateType = reflect.TypeFor[*unixtime.Nano]()

func (c *NullableDateColumn) Type() reflect.Type { return nullableDateType }
func (c *NullableDateColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableDateColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableDateColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero unixtime.Nano
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayDateColumn struct {
	ColumnOf[[]unixtime.Nano]
	elem DateColumn
}

var _ Columnar = (*ArrayDateColumn)(nil)

func NewArrayDateColumn() Columnar { return new(ArrayDateColumn) }
func (c *ArrayDateColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayDateType = reflect.TypeFor[[]unixtime.Nano]()

func (c *ArrayDateColumn) Type() reflect.Type                  { return arrayDateType }
func (c *ArrayDateColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayDateColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayDateColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayDateColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayDateColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableDateColumn struct {
	ColumnOf[[]*unixtime.Nano]
	elem NullableDateColumn
}

var _ Columnar = (*ArrayNullableDateColumn)(nil)

func NewArrayNullableDateColumn() Columnar { return new(ArrayNullableDateColumn) }
func (c *ArrayNullableDateColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableDateType = reflect.TypeFor[[]*unixtime.Nano]()

func (c *ArrayNullableDateColumn) Type() reflect.Type                  { return arrayNullableDateType }
func (c *ArrayNullableDateColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableDateColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableDateColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableDateColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayDateColumn struct {
	ColumnOf[[][]unixtime.Nano]
	elem ArrayDateColumn
}

var _ Columnar = (*ArrayArrayDateColumn)(nil)

func NewArrayArrayDateColumn() Columnar { return new(ArrayArrayDateColumn) }
func (c *ArrayArrayDateColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayDateType = reflect.TypeFor[[][]unixtime.Nano]()

func (c *ArrayArrayDateColumn) Type() reflect.Type                  { return arrayArrayDateType }
func (c *ArrayArrayDateColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayDateColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayDateColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayDateColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type NullableDateTime64Column struct {
	ColumnOf[*int64]
	Nulls UInt8Column
	Data  DateTime64Column
}

var _ Columnar = (*NullableDateTime64Column)(nil)

func NewNullableDateTime64Column() Columnar { return &NullableDateTime64Column{} }
func (c *NullableDateTime64Column) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableDateTime64Column) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableDateTime64Column) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableDateTime64Type = reflect.TypeFor[*int64]()

func (c *NullableDateTime64Column) Type() reflect.Type { return nullableDateTime64Type }
func (c *NullableDateTime64Column) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableDateTime64Column) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableDateTime64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero int64
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayDateTime64Column struct {
	ColumnOf[[]int64]
	elem DateTime64Column
}

var _ Columnar = (*ArrayDateTime64Column)(nil)

func NewArrayDateTime64Column() Columnar { return new(ArrayDateTime64Column) }
func (c *ArrayDateTime64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayDateTime64Type = reflect.TypeFor[[]int64]()

func (c *ArrayDateTime64Column) Type() reflect.Type                  { return arrayDateTime64Type }
func (c *ArrayDateTime64Column) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayDateTime64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayDateTime64Column) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayDateTime64Column) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayDateTime64Column) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableDateTime64Column struct {
	ColumnOf[[]*int64]
	elem NullableDateTime64Column
}

var _ Columnar = (*ArrayNullableDateTime64Column)(nil)

func NewArrayNullableDateTime64Column() Columnar { return new(ArrayNullableDateTime64Column) }
func (c *ArrayNullableDateTime64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableDateTime64Type = reflect.TypeFor[[]*int64]()

func (c *ArrayNullableDateTime64Column) Type() reflect.Type { return arrayNullableDateTime64Type }
func (c *ArrayNullableDateTime64Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableDateTime64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableDateTime64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableDateTime64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayDateTime64Column struct {
	ColumnOf[[][]int64]
	elem ArrayDateTime64Column
}

var _ Columnar = (*ArrayArrayDateTime64Column)(nil)

func NewArrayArrayDateTime64Column() Columnar { return new(ArrayArrayDateTime64Column) }
func (c *ArrayArrayDateTime64Column) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayDateTime64Type = reflect.TypeFor[[][]int64]()

func (c *ArrayArrayDateTime64Column) Type() reflect.Type { return arrayArrayDateTime64Type }
func (c *ArrayArrayDateTime64Column) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayArrayDateTime64Column) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayDateTime64Column) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayDateTime64Column) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type UUIDColumn struct{ ColumnOf[UUID] }

var _ Columnar = (*UUIDColumn)(nil)

func NewUUIDColumn() Columnar { return new(UUIDColumn) }

var _UUIDType = reflect.TypeFor[UUID]()

func (c *UUIDColumn) Type() reflect.Type { return _UUIDType }

type NullableUUIDColumn struct {
	ColumnOf[*UUID]
	Nulls UInt8Column
	Data  UUIDColumn
}

var _ Columnar = (*NullableUUIDColumn)(nil)

func NewNullableUUIDColumn() Columnar { return &NullableUUIDColumn{} }
func (c *NullableUUIDColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableUUIDColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableUUIDColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableUUIDType = reflect.TypeFor[*UUID]()

func (c *NullableUUIDColumn) Type() reflect.Type { return nullableUUIDType }
func (c *NullableUUIDColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableUUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableUUIDColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero UUID
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayUUIDColumn struct {
	ColumnOf[[]UUID]
	elem UUIDColumn
}

var _ Columnar = (*ArrayUUIDColumn)(nil)

func NewArrayUUIDColumn() Columnar { return new(ArrayUUIDColumn) }
func (c *ArrayUUIDColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayUUIDType = reflect.TypeFor[[]UUID]()

func (c *ArrayUUIDColumn) Type() reflect.Type                  { return arrayUUIDType }
func (c *ArrayUUIDColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayUUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayUUIDColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayUUIDColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayUUIDColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableUUIDColumn struct {
	ColumnOf[[]*UUID]
	elem NullableUUIDColumn
}

var _ Columnar = (*ArrayNullableUUIDColumn)(nil)

func NewArrayNullableUUIDColumn() Columnar { return new(ArrayNullableUUIDColumn) }
func (c *ArrayNullableUUIDColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableUUIDType = reflect.TypeFor[[]*UUID]()

func (c *ArrayNullableUUIDColumn) Type() reflect.Type                  { return arrayNullableUUIDType }
func (c *ArrayNullableUUIDColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayNullableUUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableUUIDColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableUUIDColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayUUIDColumn struct {
	ColumnOf[[][]UUID]
	elem ArrayUUIDColumn
}

var _ Columnar = (*ArrayArrayUUIDColumn)(nil)

func NewArrayArrayUUIDColumn() Columnar { return new(ArrayArrayUUIDColumn) }
func (c *ArrayArrayUUIDColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayUUIDType = reflect.TypeFor[[][]UUID]()

func (c *ArrayArrayUUIDColumn) Type() reflect.Type                  { return arrayArrayUUIDType }
func (c *ArrayArrayUUIDColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayUUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayUUIDColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayArrayUUIDColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}

type NullableTDigestColumn struct {
	ColumnOf[*[]float32]
	Nulls UInt8Column
	Data  TDigestColumn
}

var _ Columnar = (*NullableTDigestColumn)(nil)

func NewNullableTDigestColumn() Columnar { return &NullableTDigestColumn{} }
func (c *NullableTDigestColumn) Init(chType string, goType reflect.Type) error {
	return c.Data.Init(chArrayElemType(chType), nil)
}
func (c *NullableTDigestColumn) Clear() { c.ColumnOf.Clear(); c.Nulls.Clear(); c.Data.Clear() }
func (c *NullableTDigestColumn) Grow(numRow int) {
	c.ColumnOf.Grow(numRow)
	c.Nulls.Grow(numRow)
	c.Data.Grow(numRow)
}

var nullableTDigestType = reflect.TypeFor[*[]float32]()

func (c *NullableTDigestColumn) Type() reflect.Type { return nullableTDigestType }
func (c *NullableTDigestColumn) ConvertAssign(idx int, typ reflect.Type, ptr unsafe.Pointer) error {
	if c.Nulls.Column[idx] == 1 {
		return nil
	}
	dest := reflect.NewAt(typ, ptr).Elem()
	if dest.IsNil() {
		dest.Set(reflect.New(dest.Type().Elem()))
	}
	typ = typ.Elem()
	return c.Data.ConvertAssign(idx, typ, *(*unsafe.Pointer)(ptr))
}
func (c *NullableTDigestColumn) ReadData(rd *chproto.Reader, numRow int) error {
	if err := c.Nulls.ReadData(rd, numRow); err != nil {
		return err
	}
	if err := c.Data.ReadData(rd, numRow); err != nil {
		return err
	}
	c.Column = c.Column[:len(c.Nulls.Column)]
	for i, isNull := range c.Nulls.Column {
		if isNull == 1 {
			c.Column[i] = nil
		} else {
			c.Column[i] = &c.Data.Column[i]
		}
	}
	return nil
}
func (c *NullableTDigestColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		if el != nil {
			c.Nulls.Column = append(c.Nulls.Column, 0)
			c.Data.Column = append(c.Data.Column, *el)
		} else {
			c.Nulls.Column = append(c.Nulls.Column, 1)
			var zero []float32
			c.Data.Column = append(c.Data.Column, zero)
		}
	}
	if err := c.Nulls.WriteData(wr); err != nil {
		return err
	}
	return c.Data.WriteData(wr)
}

type ArrayTDigestColumn struct {
	ColumnOf[[][]float32]
	elem TDigestColumn
}

var _ Columnar = (*ArrayTDigestColumn)(nil)

func NewArrayTDigestColumn() Columnar { return new(ArrayTDigestColumn) }
func (c *ArrayTDigestColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), nil)
}

var arrayTDigestType = reflect.TypeFor[[][]float32]()

func (c *ArrayTDigestColumn) Type() reflect.Type                  { return arrayTDigestType }
func (c *ArrayTDigestColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayTDigestColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayTDigestColumn) WritePrefix(wr *chproto.Writer) error { return c.elem.WritePrefix(wr) }
func (c *ArrayTDigestColumn) WriteData(wr *chproto.Writer) error {
	_ = c.writeOffset(wr, 0)
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				c.elem.AddPointer(unsafe.Pointer(&el))
			}
		}
		return c.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		c.elem.Column = el
		if err := c.elem.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (c *ArrayTDigestColumn) writeOffset(wr *chproto.Writer, offset int) int {
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	return offset
}

type ArrayNullableTDigestColumn struct {
	ColumnOf[[]*[]float32]
	elem NullableTDigestColumn
}

var _ Columnar = (*ArrayNullableTDigestColumn)(nil)

func NewArrayNullableTDigestColumn() Columnar { return new(ArrayNullableTDigestColumn) }
func (c *ArrayNullableTDigestColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chType), goType)
}

var arrayNullableTDigestType = reflect.TypeFor[[]*[]float32]()

func (c *ArrayNullableTDigestColumn) Type() reflect.Type { return arrayNullableTDigestType }
func (c *ArrayNullableTDigestColumn) ReadPrefix(rd *chproto.Reader) error {
	return c.elem.ReadPrefix(rd)
}
func (c *ArrayNullableTDigestColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	if numElem := offsets[len(offsets)-1]; numElem > 0 {
		c.elem.Column = nil
		c.elem.Grow(numElem)
		if err := c.elem.ReadData(rd, numElem); err != nil {
			return err
		}
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayNullableTDigestColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayNullableTDigestColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	c.elem.Grow(len(c.Column))
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.AddPointer(unsafe.Pointer(&el))
		}
	}
	return c.elem.WriteData(wr)
}

type ArrayArrayTDigestColumn struct {
	ColumnOf[[][][]float32]
	elem ArrayTDigestColumn
}

var _ Columnar = (*ArrayArrayTDigestColumn)(nil)

func NewArrayArrayTDigestColumn() Columnar { return new(ArrayArrayTDigestColumn) }
func (c *ArrayArrayTDigestColumn) Init(chType string, goType reflect.Type) error {
	return c.elem.Init(chArrayElemType(chArrayElemType(chType)), nil)
}

var arrayArrayTDigestType = reflect.TypeFor[[][][]float32]()

func (c *ArrayArrayTDigestColumn) Type() reflect.Type                  { return arrayArrayTDigestType }
func (c *ArrayArrayTDigestColumn) ReadPrefix(rd *chproto.Reader) error { return c.elem.ReadPrefix(rd) }
func (c *ArrayArrayTDigestColumn) ReadData(rd *chproto.Reader, numRow int) error {
	offsets, err := readOffsets(rd, numRow)
	if err != nil {
		return err
	}
	c.Column = c.Column[:len(offsets)]
	numElem := offsets[len(offsets)-1]
	if numElem == 0 {
		return nil
	}
	c.elem.Column = nil
	c.elem.Grow(numElem)
	if err := c.elem.ReadData(rd, numElem); err != nil {
		return err
	}
	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.elem.Column[prev:offset:offset]
		prev = offset
	}
	return nil
}
func (c *ArrayArrayTDigestColumn) WritePrefix(wr *chproto.Writer) error {
	return c.elem.WritePrefix(wr)
}
func (c *ArrayArrayTDigestColumn) WriteData(wr *chproto.Writer) error {
	var offset int
	for _, el := range c.Column {
		offset += len(el)
		wr.UInt64(uint64(offset))
	}
	offset = 0
	for _, el := range c.Column {
		c.elem.Column = el
		offset = c.elem.writeOffset(wr, offset)
	}
	if _, ok := (any)(c.elem).(CustomEncoding); ok {
		c.elem.elem.Grow(len(c.Column))
		for _, el := range c.Column {
			for _, el := range el {
				for _, el := range el {
					c.elem.elem.AddPointer(unsafe.Pointer(&el))
				}
			}
		}
		return c.elem.elem.WriteData(wr)
	}
	for _, el := range c.Column {
		for _, el := range el {
			c.elem.elem.Column = el
			if err := c.elem.elem.WriteData(wr); err != nil {
				return err
			}
		}
	}
	return nil
}
