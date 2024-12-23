package chschema

import (
	"fmt"
	"github.com/uptrace/pkg/clickhouse/bfloat16"
	"github.com/uptrace/pkg/clickhouse/ch/chtype"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"github.com/uptrace/pkg/unixtime"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type UUID [16]byte

var (
	anyType          = reflect.TypeFor[any]()
	boolType         = reflect.TypeFor[bool]()
	int8Type         = reflect.TypeFor[int8]()
	int16Type        = reflect.TypeFor[int16]()
	int32Type        = reflect.TypeFor[int32]()
	int64Type        = reflect.TypeFor[int64]()
	uint8Type        = reflect.TypeFor[uint8]()
	uint16Type       = reflect.TypeFor[uint16]()
	uint32Type       = reflect.TypeFor[uint32]()
	uint64Type       = reflect.TypeFor[uint64]()
	float32Type      = reflect.TypeFor[float32]()
	float64Type      = reflect.TypeFor[float64]()
	stringType       = reflect.TypeFor[string]()
	bytesType        = reflect.TypeFor[[]byte]()
	ipType           = reflect.TypeFor[net.IP]()
	ipNetType        = reflect.TypeFor[net.IPNet]()
	bfloat16MapType  = reflect.TypeFor[map[bfloat16.T]uint64]()
	variantType      = reflect.TypeFor[Variant]()
	timeType         = reflect.TypeFor[time.Time]()
	unixNanoType     = reflect.TypeFor[unixtime.Nano]()
	sliceUint64Type  = reflect.TypeFor[[]uint64]()
	sliceFloat32Type = reflect.TypeFor[[]float32]()
)
var (
	int64PtrType   = reflect.TypeFor[*int64]()
	uint64PtrType  = reflect.TypeFor[*uint64]()
	float64PtrType = reflect.TypeFor[*float64]()
	stringPtrType  = reflect.TypeFor[*string]()
)
var chTypes = [...]string{reflect.Bool: chtype.Bool, reflect.Int: chtype.Int64, reflect.Int8: chtype.Int8, reflect.Int16: chtype.Int16, reflect.Int32: chtype.Int32, reflect.Int64: chtype.Int64, reflect.Uint: chtype.UInt64, reflect.Uint8: chtype.UInt8, reflect.Uint16: chtype.UInt16, reflect.Uint32: chtype.UInt32, reflect.Uint64: chtype.UInt64, reflect.Uintptr: "", reflect.Float32: chtype.Float32, reflect.Float64: chtype.Float64, reflect.Complex64: "", reflect.Complex128: "", reflect.Array: "", reflect.Chan: "", reflect.Func: "", reflect.Interface: chtype.Any, reflect.Map: chtype.JSON, reflect.Ptr: "", reflect.Slice: "", reflect.String: chtype.String, reflect.Struct: chtype.JSON, reflect.UnsafePointer: ""}

type NewColumnFunc func() Columnar

func NewColumn(chType string, goType reflect.Type) Columnar {
	col := ColumnFactory(chType, goType)()
	col.Init(chType, goType)
	return col
}
func ColumnFactory(chType string, typ reflect.Type) NewColumnFunc {
	switch chType {
	case chtype.String:
		if typ == bytesType {
			return NewBytesColumn
		}
		return NewStringColumn
	case "LowCardinality(String)":
		return NewLCStringColumn
	case "LowCardinality(UInt64)":
		return NewLCUInt64Column
	case chtype.Int8:
		return NewInt8Column
	case chtype.Int16:
		return NewInt16Column
	case chtype.Int32:
		return NewInt32Column
	case chtype.Int64:
		return NewInt64Column
	case chtype.UInt8:
		return NewUInt8Column
	case chtype.UInt16:
		return NewUInt16Column
	case chtype.UInt32:
		return NewUInt32Column
	case chtype.UInt64:
		return NewUInt64Column
	case chtype.Float32:
		return NewFloat32Column
	case chtype.Float64:
		return NewFloat64Column
	case chtype.Date:
		return NewDateColumn
	case chtype.DateTime:
		switch typ {
		case nil, unixNanoType:
			return NewDateTimeColumn
		case timeType:
			return NewGoDateTimeColumn
		default:
			panic(fmt.Errorf("got type %s, wanted %s", typ, unixNanoType))
		}
	case chtype.Bool:
		return NewBoolColumn
	case chtype.UUID:
		return NewUUIDColumn
	case chtype.IPv6:
		return NewIPColumn
	case chtype.Dynamic:
		return NewDynamicColumn
	case chtype.JSON:
		return NewJSONColumn
	case "Array(Int8)":
		return NewArrayInt8Column
	case "Array(UInt8)":
		return NewArrayUInt8Column
	case "Array(Int16)":
		return NewArrayInt16Column
	case "Array(UInt16)":
		return NewArrayUInt16Column
	case "Array(Int32)":
		return NewArrayInt32Column
	case "Array(UInt32)":
		return NewArrayUInt32Column
	case "Array(Int64)":
		return NewArrayInt64Column
	case "Array(UInt64)":
		return NewArrayUInt64Column
	case "Array(Float32)":
		return NewArrayFloat32Column
	case "Array(Float64)":
		return NewArrayFloat64Column
	case "Array(String)":
		return NewArrayStringColumn
	case "Array(LowCardinality(String))":
		return NewArrayLCStringColumn
	case "Array(DateTime)":
		return NewArrayDateTimeColumn
	case "Array(DateTime64)":
		return NewArrayDateTime64Column
	case "Array(Bool)":
		return NewArrayBoolColumn
	case "Array(AggregateFunction(quantilesTDigest(0.5), Float32))":
		return NewArrayTDigestColumn
	case "Nullable(String)":
		return NewNullableStringColumn
	case "Array(Nullable(String))":
		return NewArrayNullableStringColumn
	case "Array(Nullable(Int64))":
		return NewArrayNullableInt64Column
	case "Array(Nullable(Float64))":
		return NewArrayNullableFloat64Column
	case "Array(Array(Int8))":
		return NewArrayArrayInt8Column
	case "Array(Array(UInt8))":
		return NewArrayArrayUInt8Column
	case "Array(Array(Int16))":
		return NewArrayArrayInt16Column
	case "Array(Array(UInt16))":
		return NewArrayArrayUInt16Column
	case "Array(Array(Int32))":
		return NewArrayArrayInt32Column
	case "Array(Array(UInt32))":
		return NewArrayArrayUInt32Column
	case "Array(Array(Int64))":
		return NewArrayArrayInt64Column
	case "Array(Array(UInt64))":
		return NewArrayArrayUInt64Column
	case "Array(Array(Float32))":
		return NewArrayArrayFloat32Column
	case "Array(Array(Float64))":
		return NewArrayArrayFloat64Column
	case "Array(Array(String))":
		return NewArrayArrayStringColumn
	case "Array(Array(DateTime))":
		return NewArrayArrayGoDateTimeColumn
	case "Array(Array(Bool))":
		return NewArrayArrayStringColumn
	case chtype.Any:
		return nil
	}
	if chType := chEnumType(chType); chType != "" {
		return NewEnumColumn
	}
	if chType := chArrayElemType(chType); chType != "" {
		if strings.HasPrefix(chType, "Variant(") {
			return NewArrayVariantColumn
		}
		if chType := chEnumType(chType); chType != "" {
			return NewArrayEnumColumn
		}
	}
	if isDateTime64Type(chType) {
		switch typ {
		case nil, unixNanoType:
			return NewDateTime64Column
		case timeType:
			return NewGoDateTime64Column
		default:
			panic(fmt.Errorf("got type %s, wanted %s", typ, unixNanoType))
		}
	}
	if chType := chDateTimeType(chType); chType != "" {
		return ColumnFactory(chType, typ)
	}
	if chType := chSimpleAggFunc(chType); chType != "" {
		return ColumnFactory(chType, typ)
	}
	if strings.HasPrefix(chType, "JSON(") {
		return NewJSONColumn
	}
	if strings.HasPrefix(chType, "Variant(") {
		return NewVariantColumn
	}
	if funcName, _ := aggFuncNameAndType(chType); funcName != "" {
		switch funcName {
		case "quantileTDigest", "quantilesTDigest", "quantileTDigestWeighted", "quantilesTDigestWeighted":
			if typ == bfloat16MapType {
				return NewTDigestMapColumn
			}
			return NewTDigestColumn
		case "quantileTiming", "quantilesTiming":
			return NewQTimingColumn
		case "quantileBFloat16", "quantilesBFloat16":
			return NewBFloat16HistColumn
		default:
			panic(fmt.Errorf("unsupported ClickHouse type: %s", chType))
		}
	}
	if typ == nil {
		panic(fmt.Errorf("unsupported ClickHouse column: %s", chType))
	}
	kind := typ.Kind()
	switch kind {
	case reflect.Ptr:
		switch typ.Elem().Kind() {
		case reflect.Struct:
			return NewJSONColumn
		case reflect.String:
			return NewNullableStringColumn
		}
	case reflect.Slice:
		switch elem := typ.Elem(); elem.Kind() {
		case reflect.Ptr:
			if elem.Elem().Kind() == reflect.Struct {
				return NewJSONColumn
			}
		case reflect.Struct:
			if elem != timeType {
				return NewJSONColumn
			}
		}
	}
	panic(fmt.Errorf("unsupported ClickHouse column: %s", chType))
}
func CHType(typ reflect.Type) string {
	switch typ {
	case timeType:
		return chtype.DateTime
	case unixNanoType:
		return chtype.DateTime64
	case ipType:
		return chtype.IPv6
	}
	kind := typ.Kind()
	switch kind {
	case reflect.Ptr:
		if typ.Elem().Kind() == reflect.Struct {
			return chtype.JSON
		}
		return fmt.Sprintf("Nullable(%s)", CHType(typ.Elem()))
	case reflect.Slice:
		switch elem := typ.Elem(); elem.Kind() {
		case reflect.Ptr:
			if elem.Elem().Kind() == reflect.Struct {
				return chtype.JSON
			}
		case reflect.Struct:
			if elem != timeType {
				return chtype.JSON
			}
		case reflect.Uint8:
			return chtype.String
		}
		return "Array(" + CHType(typ.Elem()) + ")"
	case reflect.Array:
		if isUUID(typ) {
			return chtype.UUID
		}
	}
	if s := chTypes[kind]; s != "" {
		return s
	}
	panic(fmt.Errorf("ch: unsupported Go type: %s", typ))
}
func chArrayElemType(s string) string {
	if s := chSubType(s, "SimpleAggregateFunction("); s != "" {
		if i := strings.Index(s, ", "); i >= 0 {
			s = s[i+2:]
		}
		return chSubType(s, "Array(")
	}
	s = chSubType(s, "Array(")
	if s == "" {
		return ""
	}
	elemType := s
	s = chSubType(s, "SimpleAggregateFunction(")
	if s == "" {
		return elemType
	}
	if i := strings.Index(s, ", "); i >= 0 {
		return s[i+2:]
	}
	return s
}
func chEnumType(s string) string { return chSubType(s, "Enum8(") }
func chSimpleAggFunc(s string) string {
	s = chSubType(s, "SimpleAggregateFunction(")
	if s == "" {
		return ""
	}
	i := strings.Index(s, ", ")
	if i == -1 {
		return ""
	}
	return s[i+2:]
}
func chDateTimeType(s string) string {
	s = chSubType(s, "DateTime(")
	if s == "" {
		return ""
	}
	if s != "'UTC'" {
		internal.Logger.Printf("DateTime has timezeone=%q, expected UTC", s)
	}
	return chtype.DateTime
}
func isDateTime64Type(s string) bool {
	if s == chtype.DateTime64 {
		return true
	}
	return chSubType(s, "DateTime64(") != ""
}
func parseDateTime64Prec(str string) int {
	const defaultPrec = 9
	const wanted = "DateTime64("
	i := strings.Index(str, wanted)
	if i == -1 {
		return defaultPrec
	}
	str = str[i+len(wanted):]
	prec, err := strconv.Atoi(str[:1])
	if err != nil {
		return defaultPrec
	}
	return prec
}
func chNullableType(s string) string { return chSubType(s, "Nullable(") }
func aggFuncNameAndType(chType string) (funcName, funcType string) {
	var s string
	for _, prefix := range []string{"SimpleAggregateFunction(", "AggregateFunction("} {
		s = chSubType(chType, prefix)
		if s != "" {
			break
		}
	}
	if s == "" {
		return "", ""
	}
	const sep = ", "
	idx := strings.LastIndex(s, sep)
	if idx == -1 {
		return "", ""
	}
	funcName = s[:idx]
	funcType = s[idx+len(sep):]
	if idx := strings.IndexByte(funcName, '('); idx >= 0 {
		funcName = funcName[:idx]
	}
	return funcName, funcType
}
func chSubType(s, prefix string) string {
	if strings.HasPrefix(s, prefix) && strings.HasSuffix(s, ")") {
		return s[len(prefix) : len(s)-1]
	}
	return ""
}
func isUUID(typ reflect.Type) bool { return typ.Len() == 16 && typ.Elem().Kind() == reflect.Uint8 }
