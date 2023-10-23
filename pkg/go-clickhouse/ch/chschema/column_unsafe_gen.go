//go:build amd64 || arm64

package chschema

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

func (c *Int8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 8 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Int8Column) WriteTo(wr *chproto.Writer) error {
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

func (c *UInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 8 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *UInt8Column) WriteTo(wr *chproto.Writer) error {
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

func (c *Int16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 16 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Int16Column) WriteTo(wr *chproto.Writer) error {
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

func (c *UInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 16 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *UInt16Column) WriteTo(wr *chproto.Writer) error {
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

func (c *Int32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Int32Column) WriteTo(wr *chproto.Writer) error {
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

func (c *UInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *UInt32Column) WriteTo(wr *chproto.Writer) error {
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

func (c *Int64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Int64Column) WriteTo(wr *chproto.Writer) error {
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

func (c *UInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *UInt64Column) WriteTo(wr *chproto.Writer) error {
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

func (c *Float32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 32 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Float32Column) WriteTo(wr *chproto.Writer) error {
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

func (c *Float64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = 64 / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *Float64Column) WriteTo(wr *chproto.Writer) error {
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
