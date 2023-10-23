//go:build !amd64 && !arm64

package chschema

import (
	"github.com/uptrace/go-clickhouse/ch/chproto"
)

func (c *Int8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int8()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Int8Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int8(n)
	}
	return nil
}

func (c *UInt8Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.UInt8()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *UInt8Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.UInt8(n)
	}
	return nil
}

func (c *Int16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int16()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Int16Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int16(n)
	}
	return nil
}

func (c *UInt16Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.UInt16()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *UInt16Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.UInt16(n)
	}
	return nil
}

func (c *Int32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Int32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int32(n)
	}
	return nil
}

func (c *UInt32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.UInt32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *UInt32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.UInt32(n)
	}
	return nil
}

func (c *Int64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Int64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Int64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Int64(n)
	}
	return nil
}

func (c *UInt64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.UInt64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *UInt64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.UInt64(n)
	}
	return nil
}

func (c *Float32Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Float32()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Float32Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Float32(n)
	}
	return nil
}

func (c *Float64Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.Float64()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *Float64Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.Float64(n)
	}
	return nil
}
