package chschema

import (
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"math"
)

const lcVersion = 1

type LCStringColumn struct {
	CustomEncoding
	StringColumn
}

var _ Columnar = (*LCStringColumn)(nil)

func NewLCStringColumn() Columnar { return new(LCStringColumn) }
func (c *LCStringColumn) ReadPrefix(rd *chproto.Reader) error {
	version, err := rd.Int64()
	if err != nil {
		return err
	}
	if version != lcVersion {
		return fmt.Errorf("chschema: unsupported LowCardinality version: %d", version)
	}
	return nil
}
func (c *LCStringColumn) ReadData(rd *chproto.Reader, numRow int) error {
	flags, err := rd.Int64()
	if err != nil {
		return err
	}
	lcKey := newLCKeyType(flags & 0xf)
	dictSize, err := rd.UInt64()
	if err != nil {
		return err
	}
	dict := rd.AllocDict(int(dictSize))
	for i := range dict {
		s, err := rd.String()
		if err != nil {
			return err
		}
		dict[i] = s
	}
	numKey, err := rd.UInt64()
	if err != nil {
		return err
	}
	if int(numKey) != numRow {
		return fmt.Errorf("%d != %d", numKey, numRow)
	}
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		key, err := lcKey.read(rd)
		if err != nil {
			return err
		}
		c.Column[i] = dict[key]
	}
	return nil
}
func (c *LCStringColumn) WritePrefix(wr *chproto.Writer) error {
	wr.Int64(lcVersion)
	return nil
}
func (c *LCStringColumn) WriteData(wr *chproto.Writer) error {
	if len(c.Column) == 0 {
		return nil
	}
	lc := wr.LowCard()
	keys := lc.MakeKeys(len(c.Column))
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
	return nil
}

type LCUInt64Column struct {
	CustomEncoding
	UInt64Column
}

var _ Columnar = (*LCUInt64Column)(nil)

func NewLCUInt64Column() Columnar { return new(LCUInt64Column) }
func (c *LCUInt64Column) ReadPrefix(rd *chproto.Reader) error {
	version, err := rd.Int64()
	if err != nil {
		return err
	}
	if version != lcVersion {
		return fmt.Errorf("chschema: unsupported LowCardinality version: %d", version)
	}
	return nil
}
func (c *LCUInt64Column) ReadData(rd *chproto.Reader, numRow int) error {
	flags, err := rd.Int64()
	if err != nil {
		return err
	}
	lcKey := newLCKeyType(flags & 0xf)
	dictSize, err := rd.UInt64()
	if err != nil {
		return err
	}
	dict := make([]uint64, dictSize)
	for i := range dict {
		num, err := rd.UInt64()
		if err != nil {
			return err
		}
		dict[i] = num
	}
	numKey, err := rd.UInt64()
	if err != nil {
		return err
	}
	if int(numKey) != numRow {
		return fmt.Errorf("%d != %d", numKey, numRow)
	}
	c.Column = c.Column[:numRow]
	for i := range c.Column {
		key, err := lcKey.read(rd)
		if err != nil {
			return err
		}
		c.Column[i] = dict[key]
	}
	return nil
}
func (c *LCUInt64Column) WriteTo(wr *chproto.Writer) error { panic("not implemented") }

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
		return lcKey{typ: 0, read: func(rd *chproto.Reader) (int, error) {
			n, err := rd.UInt8()
			return int(n), err
		}, write: func(wr *chproto.Writer, n int) { wr.UInt8(uint8(n)) }}
	case 1:
		return lcKey{typ: int8(1), read: func(rd *chproto.Reader) (int, error) {
			n, err := rd.UInt16()
			return int(n), err
		}, write: func(wr *chproto.Writer, n int) { wr.UInt16(uint16(n)) }}
	case 2:
		return lcKey{typ: 2, read: func(rd *chproto.Reader) (int, error) {
			n, err := rd.UInt32()
			return int(n), err
		}, write: func(wr *chproto.Writer, n int) { wr.UInt32(uint32(n)) }}
	case 3:
		return lcKey{typ: 3, read: func(rd *chproto.Reader) (int, error) {
			n, err := rd.UInt64()
			return int(n), err
		}, write: func(wr *chproto.Writer, n int) { wr.UInt64(uint64(n)) }}
	default:
		panic("not reached")
	}
}
