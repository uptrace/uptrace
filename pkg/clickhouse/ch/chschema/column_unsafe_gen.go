//go:build amd64 || arm64

package chschema

import (
	"github.com/segmentio/asm/bswap"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"io"
	"reflect"
	"unsafe"
)

func (c *Int8Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 8 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Int8Column) WriteData(wr *chproto.Writer) error {
	const size = 8 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *UInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 8 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *UInt8Column) WriteData(wr *chproto.Writer) error {
	const size = 8 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *Int16Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 16 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Int16Column) WriteData(wr *chproto.Writer) error {
	const size = 16 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *UInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 16 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *UInt16Column) WriteData(wr *chproto.Writer) error {
	const size = 16 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *Int32Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Int32Column) WriteData(wr *chproto.Writer) error {
	const size = 32 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *UInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *UInt32Column) WriteData(wr *chproto.Writer) error {
	const size = 32 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *Int64Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Int64Column) WriteData(wr *chproto.Writer) error {
	const size = 64 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *UInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *UInt64Column) WriteData(wr *chproto.Writer) error {
	const size = 64 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *Float32Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Float32Column) WriteData(wr *chproto.Writer) error {
	const size = 32 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *Float64Column) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	return nil
}
func (c *Float64Column) WriteData(wr *chproto.Writer) error {
	const size = 64 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}
func (c *UUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	const size = 128 / 8
	c.Column = c.Column[:numRow]
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	dest := *(*[]byte)(unsafe.Pointer(&slice))
	if _, err := io.ReadFull(rd, dest); err != nil {
		return err
	}
	bswap.Swap64(dest)
	return nil
}
func (c *UUIDColumn) WriteData(wr *chproto.Writer) error {
	const size = 128 / 8
	if len(c.Column) == 0 {
		return nil
	}
	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size
	src := *(*[]byte)(unsafe.Pointer(&slice))
	bswap.Swap64(src)
	wr.Write(src)
	return nil
}
