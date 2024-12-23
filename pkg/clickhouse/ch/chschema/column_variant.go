package chschema

import (
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/segmentio/encoding/json"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"github.com/uptrace/pkg/unsafeconv"
	"log/slog"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const variantVersion = 0
const nullVariant = 255

type VariantColumn struct {
	CustomEncoding
	ColumnOf[Variant]
	variant *variantInfo
	layout  []uint8
	offsets []int
	counts  []int
	columns []Columnar
}

var _ Columnar = (*VariantColumn)(nil)

func NewVariantColumn() Columnar { return new(VariantColumn) }
func (c *VariantColumn) Init(chType string, goType reflect.Type) error {
	variant, err := parseVariantInfo(chType)
	if err != nil {
		return err
	}
	c.variant = variant
	c.counts = make([]int, c.variant.NumType())
	c.columns = make([]Columnar, c.variant.NumType())
	return nil
}
func (c *VariantColumn) Type() reflect.Type { return anyType }
func (c *VariantColumn) ReadPrefix(rd *chproto.Reader) error {
	version, err := rd.UInt64()
	if err != nil {
		return fmt.Errorf("failed to read Variant version: %w", err)
	}
	if version != variantVersion {
		return fmt.Errorf("unsupported Variant version: %d", version)
	}
	return nil
}
func (c *VariantColumn) Clear() {
	c.ColumnOf.Clear()
	for _, col := range c.columns {
		col.Clear()
	}
}
func (c *VariantColumn) ReadData(rd *chproto.Reader, numRow int) error {
	c.layout = slices.Grow(c.layout, numRow)
	c.layout = c.layout[:numRow]
	c.offsets = slices.Grow(c.offsets, numRow)
	c.offsets = c.offsets[:numRow]
	defer clear(c.counts)
	for i := 0; i < numRow; i++ {
		typeIndex, err := rd.UInt8()
		if err != nil {
			return fmt.Errorf("failed to read Dynamic type at typeIndex %d: %w", i, err)
		}
		c.layout[i] = typeIndex
		if typeIndex == nullVariant {
			continue
		}
		count := c.counts[typeIndex]
		c.counts[typeIndex] = count + 1
		c.offsets[i] = count
	}
	for typeIndex, numRow := range c.counts {
		if numRow == 0 {
			continue
		}
		col := c.columns[typeIndex]
		if col == nil {
			col = NewColumn(c.variant.Type(typeIndex), nil)
			c.columns[typeIndex] = col
		}
		col.Grow(numRow)
		if err := col.ReadData(rd, numRow); err != nil {
			return err
		}
	}
	c.Column = c.Column[:len(c.layout)]
	for i, typeIndex := range c.layout {
		if typeIndex == nullVariant {
			c.Column[i] = NilVariant()
			continue
		}
		col := c.columns[typeIndex]
		offset := c.offsets[i]
		c.Column[i] = AnyVariant(col.Index(offset))
	}
	return nil
}
func (c *VariantColumn) WritePrefix(wr *chproto.Writer) error {
	wr.UInt64(variantVersion)
	return nil
}
func (c *VariantColumn) WriteData(wr *chproto.Writer) error {
	defer clear(c.counts)
	for _, col := range c.columns {
		if col != nil {
			col.Grow(0)
		}
	}
	for _, el := range c.Column {
		index := c.variant.KindIndex(el.Kind())
		wr.UInt8(uint8(index))
		if index == nullVariant {
			continue
		}
		c.counts[index]++
		col := c.columns[index]
		if col == nil {
			col = NewColumn(c.variant.Type(index), nil)
			c.columns[index] = col
		}
		addVariant(col, el)
	}
	for _, col := range c.columns {
		if col != nil && col.Len() > 0 {
			col.WriteData(wr)
		}
	}
	return nil
}

var variantInfoCache = xsync.NewMapOf[string, *variantInfo]()

func parseVariantInfo(str string) (*variantInfo, error) {
	if v, ok := variantInfoCache.Load(str); ok {
		return v, nil
	}
	v, err := _parseVariantInfo(str)
	if err != nil {
		return nil, err
	}
	v, _ = variantInfoCache.LoadOrStore(str, v)
	return v, nil
}
func _parseVariantInfo(str string) (*variantInfo, error) {
	str, ok := strings.CutPrefix(str, "Variant(")
	if !ok {
		return nil, fmt.Errorf("invalid Variant type: %q", str)
	}
	str, ok = strings.CutSuffix(str, ")")
	if !ok {
		return nil, fmt.Errorf("invalid Variant type: %q", str)
	}
	types := make([]string, 0, 3)
	for {
		i := strings.Index(str, ", ")
		if i == -1 {
			types = append(types, str)
			return newVariantInfo(types), nil
		}
		chType := str[:i]
		types = append(types, chType)
		str = str[i+2:]
	}
}

type variantInfo struct {
	types   []string
	typeMap []int
}

func newVariantInfo(types []string) *variantInfo {
	slices.Sort(types)
	typeMap := make([]int, numKind)
	for i, typ := range types {
		typeMap[int(kindFromString(typ))] = i
	}
	return &variantInfo{types: slices.Clip(types), typeMap: typeMap}
}
func (v *variantInfo) NumType() int            { return len(v.types) }
func (v *variantInfo) Type(i int) string       { return v.types[i] }
func (v *variantInfo) KindIndex(kind Kind) int { return v.typeMap[int(kind)] }
func addVariant(col Columnar, val Variant) {
	switch val.Kind() {
	case KindString:
		val := val.string()
		col.AddPointer(unsafe.Pointer(&val))
	case KindInt64:
		val := val.int64()
		col.AddPointer(unsafe.Pointer(&val))
	case KindUInt64:
		val := val.uint64()
		col.AddPointer(unsafe.Pointer(&val))
	case KindFloat64:
		val := val.float64()
		col.AddPointer(unsafe.Pointer(&val))
	case KindArrayString:
		val := val.arrayString()
		col.AddPointer(unsafe.Pointer(&val))
	default:
		panic(fmt.Sprintf("bad kind: %s", val.Kind()))
	}
}

type Variant struct {
	_   [0]func()
	num uint64
	any any
}
type Kind int

const (
	KindNil Kind = iota
	KindString
	KindInt64
	KindUInt64
	KindFloat64
	KindArrayString
)
const numKind = int(KindArrayString) + 1

type (
	stringPtr *byte
	arrayPtr  *string
)

var kindStringSlice = []string{"Nil", "String", "Int64", "UInt64", "Float64", "Array(String)"}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(kindStringSlice) {
		return kindStringSlice[k]
	}
	return "<unknown chschema.Kind>"
}
func kindFromString(s string) Kind {
	switch s {
	case "String":
		return KindString
	case "Int64":
		return KindInt64
	case "UInt64":
		return KindUInt64
	case "Float64":
		return KindFloat64
	case "Array(String)":
		return KindArrayString
	case "Nil":
		return KindNil
	default:
		slog.Error("unsupported kind string", slog.String("kind", s))
		return KindNil
	}
}
func (v Variant) IsZero() bool { return v.num == 0 && v.any == nil }
func (v Variant) Kind() Kind {
	switch x := v.any.(type) {
	case Kind:
		return x
	case stringPtr:
		return KindString
	case arrayPtr:
		return KindArrayString
	default:
		return KindNil
	}
}
func NilVariant() Variant { return Variant{} }
func AnyVariant(v any) Variant {
	switch v := v.(type) {
	case string:
		return StringVariant(v)
	case int64:
		return Int64Variant(v)
	case uint64:
		return UInt64Variant(v)
	case float64:
		return Float64Variant(v)
	case []string:
		return ArrayStringVariant(v)
	case nil:
		return NilVariant()
	default:
		slog.Error("unsupported variant", slog.String("type", reflect.TypeOf(v).String()))
		return NilVariant()
	}
}
func Int64Variant(v int64) Variant     { return Variant{num: uint64(v), any: KindInt64} }
func UInt64Variant(v uint64) Variant   { return Variant{num: v, any: KindUInt64} }
func Float64Variant(v float64) Variant { return Variant{num: math.Float64bits(v), any: KindFloat64} }
func StringVariant(v string) Variant {
	return Variant{num: uint64(len(v)), any: stringPtr(unsafe.StringData(v))}
}
func ArrayStringVariant(v []string) Variant {
	if v == nil {
		return NilVariant()
	}
	return Variant{num: uint64(len(v)), any: arrayPtr(unsafe.SliceData(v))}
}
func (v Variant) Int64() int64 {
	if g, w := v.Kind(), KindInt64; g != w {
		panic(fmt.Sprintf("Variant kind is %s, not %s", g, w))
	}
	return v.int64()
}
func (v Variant) int64() int64 { return int64(v.num) }
func (v Variant) Uint64() uint64 {
	if g, w := v.Kind(), KindUInt64; g != w {
		panic(fmt.Sprintf("Variant kind is %s, not %s", g, w))
	}
	return v.uint64()
}
func (v Variant) uint64() uint64 { return v.num }
func (v Variant) Float64() float64 {
	if g, w := v.Kind(), KindFloat64; g != w {
		panic(fmt.Sprintf("Variant kind is %s, not %s", g, w))
	}
	return v.float64()
}
func (v Variant) float64() float64 { return math.Float64frombits(v.num) }
func (v Variant) String() string {
	if sp, ok := v.any.(stringPtr); ok {
		return unsafe.String(sp, v.num)
	}
	return unsafeconv.String(v.Append(nil))
}
func (v Variant) string() string { return unsafe.String(v.any.(stringPtr), v.num) }
func (v Variant) ArrayString() []string {
	if ptr, ok := v.any.(arrayPtr); ok {
		return unsafe.Slice((*string)(ptr), v.num)
	}
	panic("ArrayString: bad kind")
}
func (v Variant) arrayString() []string { return unsafe.Slice((*string)(v.any.(arrayPtr)), v.num) }
func (v Variant) Append(dest []byte) []byte {
	switch v.Kind() {
	case KindString:
		return append(dest, v.string()...)
	case KindInt64:
		return strconv.AppendInt(dest, v.int64(), 10)
	case KindUInt64:
		return strconv.AppendUint(dest, v.uint64(), 10)
	case KindFloat64:
		return strconv.AppendFloat(dest, v.float64(), 'g', -1, 64)
	case KindArrayString:
		if res, err := json.Append(dest, v.arrayString(), 0); err == nil {
			return res
		}
		return fmt.Append(dest, v.any)
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}
