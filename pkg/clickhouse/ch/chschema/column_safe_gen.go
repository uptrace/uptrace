//go:build !amd64 && !arm64

package chschema

import (
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
)

func (c *Int8Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Int8()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Int8Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Int8(el)
	}
	return nil
}
func (c *UInt8Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.UInt8()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *UInt8Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.UInt8(el)
	}
	return nil
}
func (c *Int16Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Int16()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Int16Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Int16(el)
	}
	return nil
}
func (c *UInt16Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.UInt16()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *UInt16Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.UInt16(el)
	}
	return nil
}
func (c *Int32Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Int32()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Int32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Int32(el)
	}
	return nil
}
func (c *UInt32Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.UInt32()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *UInt32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.UInt32(el)
	}
	return nil
}
func (c *Int64Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Int64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Int64(el)
	}
	return nil
}
func (c *UInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.UInt64()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *UInt64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.UInt64(el)
	}
	return nil
}
func (c *Float32Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Float32()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Float32Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Float32(el)
	}
	return nil
}
func (c *Float64Column) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.Float64()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *Float64Column) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.Float64(el)
	}
	return nil
}
func (c *UUIDColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		el, err := rd.UUID()
		if err != nil {
			return err
		}
		c.Column[i] = el
	}
	return nil
}
func (c *UUIDColumn) WriteData(wr *chproto.Writer) error {
	for _, el := range c.Column {
		wr.UUID(el)
	}
	return nil
}
