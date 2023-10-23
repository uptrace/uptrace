package chschema

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/go-clickhouse/ch/chtype"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

var (
	boolType    = reflect.TypeOf(false)
	int8Type    = reflect.TypeOf(int8(0))
	int16Type   = reflect.TypeOf(int16(0))
	int32Type   = reflect.TypeOf(int32(0))
	int64Type   = reflect.TypeOf(int64(0))
	uint8Type   = reflect.TypeOf(uint8(0))
	uint16Type  = reflect.TypeOf(uint16(0))
	uint32Type  = reflect.TypeOf(uint32(0))
	uint64Type  = reflect.TypeOf(uint64(0))
	float32Type = reflect.TypeOf(float32(0))
	float64Type = reflect.TypeOf(float64(0))

	stringType      = reflect.TypeOf("")
	bytesType       = reflect.TypeOf((*[]byte)(nil)).Elem()
	uuidType        = reflect.TypeOf((*UUID)(nil)).Elem()
	timeType        = reflect.TypeOf((*time.Time)(nil)).Elem()
	ipType          = reflect.TypeOf((*net.IP)(nil)).Elem()
	ipNetType       = reflect.TypeOf((*net.IPNet)(nil)).Elem()
	bfloat16MapType = reflect.TypeOf((*map[bfloat16.T]uint64)(nil)).Elem()

	sliceUint64Type  = reflect.TypeOf((*[]uint64)(nil)).Elem()
	sliceFloat32Type = reflect.TypeOf((*[]float32)(nil)).Elem()
)

var chTypes = [...]string{
	reflect.Bool:          chtype.Bool,
	reflect.Int:           chtype.Int64,
	reflect.Int8:          chtype.Int8,
	reflect.Int16:         chtype.Int16,
	reflect.Int32:         chtype.Int32,
	reflect.Int64:         chtype.Int64,
	reflect.Uint:          chtype.UInt64,
	reflect.Uint8:         chtype.UInt8,
	reflect.Uint16:        chtype.UInt16,
	reflect.Uint32:        chtype.UInt32,
	reflect.Uint64:        chtype.UInt64,
	reflect.Uintptr:       "",
	reflect.Float32:       chtype.Float32,
	reflect.Float64:       chtype.Float64,
	reflect.Complex64:     "",
	reflect.Complex128:    "",
	reflect.Array:         "",
	reflect.Chan:          "",
	reflect.Func:          "",
	reflect.Interface:     chtype.Any,
	reflect.Map:           chtype.String,
	reflect.Ptr:           "",
	reflect.Slice:         "",
	reflect.String:        chtype.String,
	reflect.Struct:        chtype.String,
	reflect.UnsafePointer: "",
}

type NewColumnFunc func() Columnar

func NewColumn(chType string, typ reflect.Type) Columnar {
	col := ColumnFactory(chType, typ)()
	col.Init(chType)
	return col
}

func ColumnFactory(chType string, typ reflect.Type) NewColumnFunc {
	switch chType {
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

	case chtype.String:
		if typ == bytesType {
			return NewBytesColumn
		}
		return NewStringColumn
	case "LowCardinality(String)":
		return NewLCStringColumn
	case chtype.Bool:
		return NewBoolColumn
	case chtype.UUID:
		return NewUUIDColumn
	case chtype.IPv6:
		return NewIPColumn

	case chtype.DateTime:
		return NewDateTimeColumn
	case chtype.Date:
		return NewDateColumn

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
	case "Array(Bool)":
		return NewArrayBoolColumn

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
		return NewArrayArrayDateTimeColumn
	case "Array(Array(Bool))":
		return NewArrayArrayStringColumn

	case chtype.Any:
		return nil
	}

	if chType := chEnumType(chType); chType != "" {
		return NewEnumColumn
	}
	if chType := chArrayElemType(chType); chType != "" {
		if chType := chEnumType(chType); chType != "" {
			return NewArrayEnumColumn
		}
	}
	if isDateTime64Type(chType) {
		return NewDateTime64Column
	}
	if chType := chDateTimeType(chType); chType != "" {
		return ColumnFactory(chType, typ)
	}
	if chType := chNullableType(chType); chType != "" {
		if typ != nil {
			typ = typ.Elem()
		}
		return NewNullableColumnFunc(ColumnFactory(chType, typ))
	}

	if chType := chSimpleAggFunc(chType); chType != "" {
		return ColumnFactory(chType, typ)
	}

	if funcName, _ := aggFuncNameAndType(chType); funcName != "" {
		switch funcName {
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
		if typ.Elem().Kind() == reflect.Struct {
			return NewJSONColumn
		}
		return NewNullableColumnFunc(ColumnFactory(chNullableType(chType), typ.Elem()))
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

func chType(typ reflect.Type) string {
	switch typ {
	case timeType:
		return chtype.DateTime
	case ipType:
		return chtype.IPv6
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Ptr:
		if typ.Elem().Kind() == reflect.Struct {
			return chtype.String
		}
		return fmt.Sprintf("Nullable(%s)", chType(typ.Elem()))
	case reflect.Slice:
		switch elem := typ.Elem(); elem.Kind() {
		case reflect.Ptr:
			if elem.Elem().Kind() == reflect.Struct {
				return chtype.String // json
			}
		case reflect.Struct:
			if elem != timeType {
				return chtype.String // json
			}
		case reflect.Uint8:
			return chtype.String // []byte
		}

		return "Array(" + chType(typ.Elem()) + ")"
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

func chEnumType(s string) string {
	return chSubType(s, "Enum8(")
}

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
	return chSubType(s, "DateTime64(") != ""
}

func parseDateTime64Prec(s string) int {
	s = chSubType(s, "DateTime64(")
	if s == "" {
		return 0
	}
	prec, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return prec
}

func chNullableType(s string) string {
	return chSubType(s, "Nullable(")
}

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

func isUUID(typ reflect.Type) bool {
	return typ.Len() == 16 && typ.Elem().Kind() == reflect.Uint8
}
