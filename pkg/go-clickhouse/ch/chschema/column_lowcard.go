package chschema

import (
	"fmt"
	"math"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

type LCStringColumn struct {
	StringColumn
}

var _ Columnar = (*LCStringColumn)(nil)

func NewLCStringColumn() Columnar {
	return new(LCStringColumn)
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
	if numRow == 0 {
		return nil
	}

	flags, err := rd.Int64()
	if err != nil {
		return err
	}
	lcKey := newLCKeyType(flags & 0xf)

	dictSize, err := rd.UInt64()
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

	numKey, err := rd.UInt64()
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
	lcKey := newLCKey(int64(len(dict)))

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

type ArrayLCStringColumn struct {
	ArrayStringColumn
	lc LCStringColumn
}

var _ Columnar = (*ArrayLCStringColumn)(nil)

func NewArrayLCStringColumn() Columnar {
	return new(ArrayLCStringColumn)
}

func (c *ArrayLCStringColumn) ConvertAssign(idx int, dest reflect.Value) error {
	dest.Set(reflect.ValueOf(c.Column[idx]))
	return nil
}

func (c *ArrayLCStringColumn) ReadFrom(rd *chproto.Reader, numRow int) error {
	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	if err := c.lc.readPrefix(rd, numRow); err != nil {
		return err
	}

	offsets, err := c.readOffsets(rd, numRow)
	if err != nil {
		return err
	}

	if err := c.lc.readData(rd, offsets[len(offsets)-1]); err != nil {
		return err
	}

	var prev int
	for i, offset := range offsets {
		c.Column[i] = c.lc.Column[prev:offset]
		prev = offset
	}

	return nil
}

func (c *ArrayLCStringColumn) WriteTo(wr *chproto.Writer) error {
	c.lc.writePrefix(wr)
	_ = c.WriteOffset(wr, 0)
	return c.WriteData(wr)
}

func (c *ArrayLCStringColumn) WriteData(wr *chproto.Writer) error {
	for _, ss := range c.Column {
		c.lc.Column = ss
		c.lc.writeData(wr)
	}
	return nil
}

//------------------------------------------------------------------------------

type lcKey struct {
	typ   int8
	read  func(*chproto.Reader) (int, error)
	write func(*chproto.Writer, int)
}

func newLCKey(numKey int64) lcKey {
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
				n, err := rd.UInt8()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.UInt8(uint8(n))
			},
		}
	case 1:
		return lcKey{
			typ: int8(1),
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.UInt16()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.UInt16(uint16(n))
			},
		}
	case 2:
		return lcKey{
			typ: 2,
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.UInt32()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.UInt32(uint32(n))
			},
		}
	case 3:
		return lcKey{
			typ: 3,
			read: func(rd *chproto.Reader) (int, error) {
				n, err := rd.UInt64()
				return int(n), err
			},
			write: func(wr *chproto.Writer, n int) {
				wr.UInt64(uint64(n))
			},
		}
	default:
		panic("not reached")
	}
}

//------------------------------------------------------------------------------

type lowCard struct {
	slice sliceMap
	dict  map[string]int
}

func (lc *lowCard) Add(word string) int {
	if i, ok := lc.dict[word]; ok {
		return i
	}

	if lc.dict == nil {
		lc.dict = make(map[string]int)
	}

	i := lc.slice.Add(word)
	lc.dict[word] = i

	return i
}

func (lc *lowCard) Dict() []string {
	return lc.slice.Slice()
}

//------------------------------------------------------------------------------

type sliceMap struct {
	ss []string
}

func (m sliceMap) Len() int {
	return len(m.ss)
}

func (m sliceMap) Get(word string) (int, bool) {
	for i, s := range m.ss {
		if s == word {
			return i, true
		}
	}
	return 0, false
}

func (m *sliceMap) Add(word string) int {
	m.ss = append(m.ss, word)
	return len(m.ss) - 1
}

func (m sliceMap) Slice() []string {
	return m.ss
}
