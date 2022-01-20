package chschema

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chtype"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

var chType = [...]string{
	reflect.Bool:          chtype.UInt8,
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

// keep in sync with ColumnFactory
func clickhouseType(typ reflect.Type) string {
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

		return "Array(" + clickhouseType(typ.Elem()) + ")"
	case reflect.Array:
		if isUUID(typ) {
			return chtype.UUID
		}
	}

	if s := chType[kind]; s != "" {
		return s
	}

	panic(fmt.Errorf("ch: unsupported Go type: %s", typ))
}

type NewColumnFunc func(typ reflect.Type, chType string, numRow int) Columnar

var kindToColumn = [...]NewColumnFunc{
	reflect.Bool:          NewBoolColumn,
	reflect.Int:           NewInt64Column,
	reflect.Int8:          NewInt8Column,
	reflect.Int16:         NewInt16Column,
	reflect.Int32:         NewInt32Column,
	reflect.Int64:         NewInt64Column,
	reflect.Uint:          NewUint64Column,
	reflect.Uint8:         NewUint8Column,
	reflect.Uint16:        NewUint16Column,
	reflect.Uint32:        NewUint32Column,
	reflect.Uint64:        NewUint64Column,
	reflect.Uintptr:       nil,
	reflect.Float32:       NewFloat32Column,
	reflect.Float64:       NewFloat64Column,
	reflect.Complex64:     nil,
	reflect.Complex128:    nil,
	reflect.Array:         nil,
	reflect.Chan:          nil,
	reflect.Func:          nil,
	reflect.Interface:     nil,
	reflect.Map:           NewJSONColumn,
	reflect.Ptr:           nil,
	reflect.Slice:         nil,
	reflect.String:        NewStringColumn,
	reflect.Struct:        NewJSONColumn,
	reflect.UnsafePointer: nil,
}

// keep in sync with clickhouseType
func ColumnFactory(typ reflect.Type, chType string) NewColumnFunc {
	if chType == chtype.Any {
		return nil
	}

	if s := lowCardinalityType(chType); s != "" {
		switch s {
		case chtype.String:
			return NewLCStringColumn
		}
		panic(fmt.Errorf("got %s, wanted LowCardinality(String)", chType))
	}

	if s := enumType(chType); s != "" {
		return NewEnumColumn
	}

	if strings.HasPrefix(chType, "SimpleAggregateFunction(") {
		chType = chSubType(chType, "SimpleAggregateFunction(")
	} else if s := dateTimeType(chType); s != "" {
		chType = s
	}

	switch typ {
	case timeType:
		switch chType {
		case chtype.DateTime:
			return NewDateTimeColumn
		case chtype.Date:
			return NewDateColumn
		case chtype.Int64:
			return NewTimeColumn
		}
	case ipType:
		return NewIPColumn
	}

	kind := typ.Kind()

	switch kind {
	case reflect.Ptr:
		if typ.Elem().Kind() == reflect.Struct {
			return NewJSONColumn
		}
	case reflect.Slice:
		switch elem := typ.Elem(); elem.Kind() {
		case reflect.Ptr:
			if elem.Elem().Kind() == reflect.Struct {
				return NewJSONColumn
			}
		case reflect.Uint8:
			if chType == chtype.String {
				return NewBytesColumn
			}
		case reflect.String:
			return NewStringArrayColumn
		case reflect.Struct:
			if elem != timeType {
				return NewJSONColumn
			}
		}

		return NewArrayColumn
	case reflect.Array:
		if isUUID(typ) {
			return NewUUIDColumn
		}
	case reflect.Interface:
		return columnFromCHType(chType)
	}

	switch chType {
	case chtype.DateTime:
		switch typ {
		case uint32Type:
			return NewUint32Column
		case int64Type:
			return NewInt64TimeColumn
		default:
			return NewDateTimeColumn
		}
	}

	fn := kindToColumn[kind]
	if fn != nil {
		return fn
	}

	panic(fmt.Errorf("unsupported go_type=%q ch_type=%q", typ.String(), chType))
}

func columnFromCHType(chType string) NewColumnFunc {
	switch chType {
	case chtype.String:
		return NewStringColumn
	case chtype.UUID:
		return NewUUIDColumn
	case chtype.Int8:
		return NewInt8Column
	case chtype.Int16:
		return NewInt16Column
	case chtype.Int32:
		return NewInt32Column
	case chtype.Int64:
		return NewInt64Column
	case chtype.UInt8:
		return NewUint8Column
	case chtype.UInt16:
		return NewUint16Column
	case chtype.UInt32:
		return NewUint32Column
	case chtype.UInt64:
		return NewUint64Column
	case chtype.Float32:
		return NewFloat32Column
	case chtype.Float64:
		return NewFloat64Column
	case chtype.DateTime:
		return NewDateTimeColumn
	case chtype.Date:
		return NewDateColumn
	case chtype.IPv6:
		return NewIPColumn
	default:
		return nil
	}
}

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

	stringType       = reflect.TypeOf("")
	bytesType        = reflect.TypeOf((*[]byte)(nil)).Elem()
	uuidType         = reflect.TypeOf((*UUID)(nil)).Elem()
	timeType         = reflect.TypeOf((*time.Time)(nil)).Elem()
	ipType           = reflect.TypeOf((*net.IP)(nil)).Elem()
	ipNetType        = reflect.TypeOf((*net.IPNet)(nil)).Elem()
	bfloat16HistType = reflect.TypeOf((*map[chtype.BFloat16]uint64)(nil)).Elem()

	int64SliceType   = reflect.TypeOf((*[]int64)(nil)).Elem()
	uint64SliceType  = reflect.TypeOf((*[]uint64)(nil)).Elem()
	float32SliceType = reflect.TypeOf((*[]float32)(nil)).Elem()
	float64SliceType = reflect.TypeOf((*[]float64)(nil)).Elem()
	stringSliceType  = reflect.TypeOf((*[]string)(nil)).Elem()
)

func goType(chType string) reflect.Type {
	switch chType {
	case chtype.Int8:
		return int8Type
	case chtype.Int32:
		return int32Type
	case chtype.Int64:
		return int64Type
	case chtype.UInt8:
		return uint8Type
	case chtype.UInt16:
		return uint16Type
	case chtype.UInt32:
		return uint32Type
	case chtype.UInt64:
		return uint64Type
	case chtype.Float32:
		return float32Type
	case chtype.Float64:
		return float64Type
	case chtype.String:
		return stringType
	case chtype.UUID:
		return uuidType
	case chtype.DateTime:
		return timeType
	case chtype.Date:
		return timeType
	case chtype.IPv6:
		return ipType
	default:
	}

	if s := chArrayElemType(chType); s != "" {
		return reflect.SliceOf(goType(s))
	}
	if s := lowCardinalityType(chType); s != "" {
		return goType(s)
	}
	if s := enumType(chType); s != "" {
		return stringType
	}
	if s := dateTimeType(chType); s != "" {
		return timeType
	}
	if _, funcType := aggFuncNameAndType(chType); funcType != "" {
		return goType(funcType)
	}

	panic(fmt.Errorf("unsupported ClickHouse type=%q", chType))
}

func chArrayElemType(s string) string {
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

func lowCardinalityType(s string) string {
	return chSubType(s, "LowCardinality(")
}

func enumType(s string) string {
	return chSubType(s, "Enum8(")
}

func dateTimeType(s string) string {
	s = chSubType(s, "DateTime(")
	if s == "" {
		return ""
	}
	if s != "'UTC'" {
		internal.Logger.Printf("DateTime has timezeone=%q, expected UTC", s)
	}
	return chtype.DateTime
}

func aggFuncNameAndType(chType string) (funcName, funcType string) {
	s := chSubType(chType, "SimpleAggregateFunction(")
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
