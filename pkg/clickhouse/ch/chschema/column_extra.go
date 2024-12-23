package chschema

import (
	"fmt"
	"github.com/uptrace/pkg/clickhouse/bfloat16"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"reflect"
)

type BFloat16HistColumn struct {
	ColumnOf[map[bfloat16.T]uint64]
}

var _ Columnar = (*BFloat16HistColumn)(nil)

func NewBFloat16HistColumn() Columnar           { return new(BFloat16HistColumn) }
func (c BFloat16HistColumn) Type() reflect.Type { return bfloat16MapType }
func (c *BFloat16HistColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
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
func (c BFloat16HistColumn) WriteData(wr *chproto.Writer) error {
	for _, m := range c.Column {
		wr.Uvarint(uint64(len(m)))
		for k, v := range m {
			wr.UInt16(uint16(k))
			wr.UInt64(v)
		}
	}
	return nil
}

type QTimingColumn struct{ ColumnOf[[]uint64] }

var _ Columnar = (*QTimingColumn)(nil)

func NewQTimingColumn() Columnar           { return new(QTimingColumn) }
func (c QTimingColumn) Type() reflect.Type { return sliceUint64Type }
func (c *QTimingColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		kind, err := rd.UInt8()
		if err != nil {
			return err
		}
		switch kind {
		case 1:
			data, err := c.readTiny(rd)
			if err != nil {
				return err
			}
			c.Column[i] = data
		case 2:
			data, err := c.readMedium(rd)
			if err != nil {
				return err
			}
			c.Column[i] = data
		case 3:
			data, err := c.readLarge(rd)
			if err != nil {
				return err
			}
			c.Column[i] = data
		default:
			return fmt.Errorf("ch: unknown quantileTiming kind: %d", kind)
		}
	}
	return nil
}
func (c *QTimingColumn) readTiny(rd *chproto.Reader) ([]uint64, error) {
	n, err := rd.UInt16()
	if err != nil {
		return nil, err
	}
	data := make([]uint64, 0, 2*n)
	for i := 0; i < int(n); i++ {
		value, err := rd.UInt16()
		if err != nil {
			return nil, err
		}
		data = append(data, uint64(value), 1)
	}
	return data, nil
}
func (c *QTimingColumn) readMedium(rd *chproto.Reader) ([]uint64, error) {
	n, err := rd.UInt64()
	if err != nil {
		return nil, err
	}
	data := make([]uint64, 0, 2*n)
	for i := 0; i < int(n); i++ {
		value, err := rd.UInt16()
		if err != nil {
			return nil, err
		}
		data = append(data, uint64(value), 1)
	}
	return data, nil
}
func (c *QTimingColumn) readLarge(rd *chproto.Reader) ([]uint64, error) {
	const smallThreshold = 1024
	const bigThreshold = 30000
	const bigPrecision = 16
	const bigSize = (bigThreshold - smallThreshold) / bigPrecision
	n, err := rd.UInt64()
	if err != nil {
		return nil, err
	}
	data := make([]uint64, 0, 2*n)
	if n*2 > smallThreshold+bigSize {
		for i := 0; i < smallThreshold; i++ {
			count, err := rd.UInt64()
			if err != nil {
				return nil, err
			}
			data = append(data, uint64(i), uint64(count))
		}
		for i := 0; i < bigSize; i++ {
			count, err := rd.UInt64()
			if err != nil {
				return nil, err
			}
			value := i*bigPrecision + smallThreshold
			data = append(data, uint64(value), uint64(count))
		}
	} else {
		for i := 0; i < smallThreshold; i++ {
			value, err := rd.UInt16()
			if err != nil {
				return nil, err
			}
			count, err := rd.UInt64()
			if err != nil {
				return nil, err
			}
			data = append(data, uint64(value), uint64(count))
		}
		for i := 0; i < bigSize; i++ {
			idx, err := rd.UInt16()
			if err != nil {
				return nil, err
			}
			count, err := rd.UInt64()
			if err != nil {
				return nil, err
			}
			value := idx*bigPrecision + smallThreshold
			data = append(data, uint64(value), uint64(count))
		}
	}
	return data, nil
}
func (c QTimingColumn) WriteData(wr *chproto.Writer) error { panic("not implemented") }
