package msgp

import (
	"github.com/uptrace/pkg/msgp/msgpcode"
	"math"
	"slices"
	"strings"
	"sync"
	"time"
)

const timeExtID = 255

func AppendNil(b []byte) []byte { return append(b, msgpcode.Nil) }
func AppendBool(b []byte, v bool) []byte {
	if v {
		return append(b, msgpcode.True)
	}
	return append(b, msgpcode.False)
}
func AppendTypedUint(b []byte, n uint64) []byte {
	if n <= math.MaxUint8 {
		return append(b, msgpcode.Uint8, byte(n))
	}
	if n <= math.MaxUint16 {
		return AppendUint16(b, uint16(n))
	}
	if n <= math.MaxUint32 {
		return AppendUint32(b, uint32(n))
	}
	return AppendUint64(b, n)
}
func AppendUvarint(b []byte, n uint64) []byte {
	if n <= math.MaxInt8 {
		return append(b, byte(n))
	}
	if n <= math.MaxUint8 {
		return append(b, msgpcode.Uint8, byte(n))
	}
	if n <= math.MaxUint16 {
		return AppendUint16(b, uint16(n))
	}
	if n <= math.MaxUint32 {
		return AppendUint32(b, uint32(n))
	}
	return AppendUint64(b, n)
}
func AppendUint16(b []byte, n uint16) []byte { return append(b, msgpcode.Uint16, byte(n>>8), byte(n)) }
func AppendUint32(b []byte, n uint32) []byte {
	return append(b, msgpcode.Uint32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendUint64(b []byte, n uint64) []byte {
	return append(b, msgpcode.Uint64, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendTypedInt(b []byte, n int64) []byte {
	if n < 0 {
		if n >= int64(int8(msgpcode.NegFixedNumLow)) {
			return append(b, byte(n))
		}
		if n >= math.MinInt8 {
			return append(b, msgpcode.Int8, byte(n))
		}
		if n >= math.MinInt16 {
			return append(b, msgpcode.Int16, byte(n>>8), byte(n))
		}
		if n >= math.MinInt32 {
			return append(b, msgpcode.Int32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
		}
		return append(b, msgpcode.Int64, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	if n <= math.MaxInt8 {
		return append(b, msgpcode.Int8, byte(n))
	}
	if n <= math.MaxInt16 {
		return append(b, msgpcode.Int16, byte(n>>8), byte(n))
	}
	if n <= math.MaxInt32 {
		return append(b, msgpcode.Int32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	return append(b, msgpcode.Int64, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendVarint(b []byte, n int64) []byte {
	if n > 0 {
		return AppendUvarint(b, uint64(n))
	}
	if n >= int64(int8(msgpcode.NegFixedNumLow)) {
		return append(b, byte(n))
	}
	if n >= math.MinInt8 {
		return append(b, msgpcode.Int8, byte(n))
	}
	if n >= math.MinInt16 {
		return append(b, msgpcode.Int16, byte(n>>8), byte(n))
	}
	if n >= math.MinInt32 {
		return append(b, msgpcode.Int32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	return append(b, msgpcode.Int64, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendFloat32(b []byte, f float32) []byte {
	n := math.Float32bits(f)
	return append(b, msgpcode.Float, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendFloat64(b []byte, f float64) []byte {
	n := math.Float64bits(f)
	return append(b, msgpcode.Double, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendString(b []byte, s string) []byte {
	b = appendStringLen(b, len(s))
	b = append(b, s...)
	return b
}
func appendStringLen(b []byte, n int) []byte {
	if n < 32 {
		return append(b, msgpcode.FixedStrLow|byte(n))
	}
	if n < 256 {
		return append(b, msgpcode.Str8, byte(n))
	}
	if n <= math.MaxUint16 {
		return append(b, msgpcode.Str16, byte(n>>8), byte(n))
	}
	return append(b, msgpcode.Str32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendBytes(b []byte, bs []byte) []byte {
	if bs == nil {
		return AppendNil(b)
	}
	b = appendBytesLen(b, len(bs))
	b = append(b, bs...)
	return b
}
func appendBytesLen(b []byte, n int) []byte {
	if n < 256 {
		return append(b, msgpcode.Bin8, byte(n))
	}
	if n <= math.MaxUint16 {
		return append(b, msgpcode.Bin16, byte(n>>8), byte(n))
	}
	return append(b, msgpcode.Bin32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendTime(b []byte, tm time.Time) []byte {
	if tm.IsZero() {
		return AppendNil(b)
	}
	seconds := uint64(tm.Unix())
	if seconds>>34 == 0 {
		data := uint64(tm.Nanosecond())<<34 | seconds
		if data&0xffffffff00000000 == 0 {
			b = append(b, msgpcode.FixExt4, timeExtID)
			return append4(b, uint32(data))
		}
		b = append(b, msgpcode.FixExt8, timeExtID)
		return append8(b, uint64(data))
	}
	b = append(b, msgpcode.Ext8, 12, timeExtID)
	b = append4(b, uint32(tm.Nanosecond()))
	b = append8(b, seconds)
	return b
}
func AppendMapInterfaceInterface(b []byte, m map[any]any, flags AppendFlags) (_ []byte, err error) {
	if m == nil {
		return AppendNil(b), nil
	}
	b = AppendMapLen(b, len(m))
	for k, v := range m {
		b, err = Append(b, k, flags)
		if err != nil {
			return b, err
		}
		b, err = Append(b, v, flags)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}
func AppendMapStringInterface(b []byte, m map[string]any, flags AppendFlags) (_ []byte, err error) {
	if m == nil {
		return AppendNil(b), nil
	}
	b = AppendMapLen(b, len(m))
	if flags&SortedMapKeys == 0 {
		for k, v := range m {
			b = AppendString(b, k)
			b, err = Append(b, v, flags)
			if err != nil {
				return b, err
			}
		}
		return b, nil
	}
	s := getMapEntrySlice(len(m))
	defer putMapEntrySlice(s)
	for k, v := range m {
		s.entries = append(s.entries, mapEntry{key: k, val: v})
	}
	slices.SortFunc(s.entries, func(a, b mapEntry) int { return strings.Compare(a.key, b.key) })
	for _, entry := range s.entries {
		b = AppendString(b, entry.key)
		b, err = Append(b, entry.val, flags)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}
func AppendMapStringString(b []byte, m map[string]string, flags AppendFlags) (_ []byte, err error) {
	b = AppendMapLen(b, len(m))
	if flags&SortedMapKeys == 0 {
		for k, v := range m {
			b = AppendString(b, k)
			b = AppendString(b, v)
		}
		return b, nil
	}
	s := getMapEntrySlice(len(m))
	defer putMapEntrySlice(s)
	for k, v := range m {
		s.entries = append(s.entries, mapEntry{key: k, val: &v})
	}
	slices.SortFunc(s.entries, func(a, b mapEntry) int { return strings.Compare(a.key, b.key) })
	for _, entry := range s.entries {
		b = AppendString(b, entry.key)
		b = AppendString(b, *entry.val.(*string))
	}
	return b, nil
}
func AppendMapStringBool(b []byte, m map[string]bool, flags AppendFlags) (_ []byte, err error) {
	b = AppendMapLen(b, len(m))
	if flags&SortedMapKeys == 0 {
		for k, v := range m {
			b = AppendString(b, k)
			b = AppendBool(b, v)
		}
		return b, nil
	}
	s := getMapEntrySlice(len(m))
	defer putMapEntrySlice(s)
	for k, v := range m {
		s.entries = append(s.entries, mapEntry{key: k, val: v})
	}
	slices.SortFunc(s.entries, func(a, b mapEntry) int { return strings.Compare(a.key, b.key) })
	for _, entry := range s.entries {
		b = AppendString(b, entry.key)
		b = AppendBool(b, entry.val.(bool))
	}
	return b, nil
}
func AppendMapLen(b []byte, n int) []byte {
	if n < 16 {
		return AppendMapLen8(b, n)
	}
	if n <= math.MaxUint16 {
		return AppendMapLen16(b, n)
	}
	return append(b, msgpcode.Map32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func AppendMapLen8(b []byte, n int) []byte  { return append(b, msgpcode.FixedMapLow|byte(n)) }
func AppendMapLen16(b []byte, n int) []byte { return append(b, msgpcode.Map16, byte(n>>8), byte(n)) }
func AppendStringSlice(b []byte, slice []string) []byte {
	if slice == nil {
		return AppendNil(b)
	}
	b = AppendArrayLen(b, len(slice))
	for _, str := range slice {
		b = AppendString(b, str)
	}
	return b
}
func AppendArrayLen(b []byte, n int) []byte {
	if n < 16 {
		return append(b, msgpcode.FixedArrayLow|byte(n))
	}
	if n <= math.MaxUint16 {
		return append(b, msgpcode.Array16, byte(n>>8), byte(n))
	}
	return append(b, msgpcode.Array32, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func append4(b []byte, n uint32) []byte {
	return append(b, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func append8(b []byte, n uint64) []byte {
	return append(b, byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

type mapEntry struct {
	key string
	val any
}
type mapEntrySlice struct{ entries []mapEntry }

var mapEntrySlicePool = sync.Pool{New: func() any { return new(mapEntrySlice) }}

func getMapEntrySlice(size int) *mapEntrySlice {
	s := mapEntrySlicePool.Get().(*mapEntrySlice)
	if cap(s.entries) < size {
		s.entries = make([]mapEntry, 0, align(10, uintptr(size)))
	}
	return s
}
func putMapEntrySlice(s *mapEntrySlice) {
	clear(s.entries)
	s.entries = s.entries[:0]
	mapEntrySlicePool.Put(s)
}
