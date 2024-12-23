package chschema

import (
	"database/sql/driver"
	"fmt"
	"github.com/uptrace/pkg/msgp"
	"github.com/uptrace/pkg/unsafeconv"
	"net"
	"reflect"
	"strconv"
	"time"
)

var (
	driverValuerType  = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	queryAppenderType = reflect.TypeOf((*QueryAppender)(nil)).Elem()
)

type AppenderFunc func(fmter Formatter, b []byte, v reflect.Value) []byte

var valueAppenders []AppenderFunc

func init() {
	valueAppenders = []AppenderFunc{reflect.Bool: appendBoolValue, reflect.Int: appendIntValue, reflect.Int8: appendIntValue, reflect.Int16: appendIntValue, reflect.Int32: appendIntValue, reflect.Int64: appendIntValue, reflect.Uint: appendUintValue, reflect.Uint8: appendUintValue, reflect.Uint16: appendUintValue, reflect.Uint32: appendUintValue, reflect.Uint64: appendUintValue, reflect.Uintptr: nil, reflect.Float32: appendFloat32Value, reflect.Float64: appendFloat64Value, reflect.Complex64: nil, reflect.Complex128: nil, reflect.Array: nil, reflect.Chan: nil, reflect.Func: nil, reflect.Interface: appendIfaceValue, reflect.Map: nil, reflect.Ptr: nil, reflect.Slice: nil, reflect.String: appendStringValue, reflect.Struct: nil, reflect.UnsafePointer: nil}
}
func Appender(typ reflect.Type) AppenderFunc {
	switch typ {
	case timeType:
		return appendTimeValue
	case ipType:
		return appendIPValue
	case ipNetType:
		return appendIPNetValue
	}
	if typ.Implements(queryAppenderType) {
		return appendQueryAppenderValue
	}
	if typ.Implements(driverValuerType) {
		return appendDriverValuerValue
	}
	kind := typ.Kind()
	if kind != reflect.Ptr {
		ptr := reflect.PtrTo(typ)
		if ptr.Implements(queryAppenderType) {
			return addrAppender(appendQueryAppenderValue)
		}
		if ptr.Implements(driverValuerType) {
			return addrAppender(appendDriverValuerValue)
		}
	}
	switch kind {
	case reflect.Ptr:
		return ptrAppenderFunc(typ)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return appendBytesValue
		}
		return arrayAppender(typ)
	case reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return appendArrayBytesValue
		}
	}
	return valueAppenders[kind]
}
func ptrAppenderFunc(typ reflect.Type) AppenderFunc {
	appender := Appender(typ.Elem())
	return func(fmter Formatter, b []byte, v reflect.Value) []byte {
		if v.IsNil() {
			return AppendNull(b)
		}
		return appender(fmter, b, v.Elem())
	}
}
func AppendValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return AppendNull(b)
	}
	appender := Appender(v.Type())
	return appender(fmter, b, v)
}
func appendIfaceValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return Append(fmter, b, v.Interface())
}
func appendBoolValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendBool(b, v.Bool())
}
func appendIntValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return strconv.AppendInt(b, v.Int(), 10)
}
func appendUintValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return strconv.AppendUint(b, v.Uint(), 10)
}
func appendFloat32Value(fmter Formatter, b []byte, v reflect.Value) []byte {
	return appendFloat(b, v.Float(), 32)
}
func appendFloat64Value(fmter Formatter, b []byte, v reflect.Value) []byte {
	return appendFloat(b, v.Float(), 64)
}
func appendBytesValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendBytes(b, v.Bytes())
}
func appendArrayBytesValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendBytes(b, v.Slice(0, v.Len()).Bytes())
}
func appendStringValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendString(b, v.String())
}
func appendTimeValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	tm := v.Interface().(time.Time)
	return AppendTime(b, tm)
}
func appendIPValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	ip := v.Interface().(net.IP)
	return AppendString(b, ip.String())
}
func appendIPNetValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	ipnet := v.Interface().(net.IPNet)
	return AppendString(b, ipnet.String())
}
func appendJSONRawMessageValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendString(b, unsafeconv.String(v.Bytes()))
}
func appendQueryAppenderValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return AppendQueryAppender(fmter, b, v.Interface().(QueryAppender))
}
func appendDriverValuerValue(fmter Formatter, b []byte, v reflect.Value) []byte {
	return appendDriverValue(fmter, b, v.Interface().(driver.Valuer))
}
func addrAppender(fn AppenderFunc) AppenderFunc {
	return func(fmter Formatter, b []byte, v reflect.Value) []byte {
		if !v.CanAddr() {
			err := fmt.Errorf("ch: Append(nonaddressable %T)", v.Interface())
			return AppendError(b, err)
		}
		return fn(fmter, b, v.Addr())
	}
}
func AppendQueryAppender(fmter Formatter, b []byte, app QueryAppender) []byte {
	bb, err := app.AppendQuery(fmter, b)
	if err != nil {
		return AppendError(b, err)
	}
	return bb
}
func appendDriverValue(fmter Formatter, b []byte, v driver.Valuer) []byte {
	value, err := v.Value()
	if err != nil {
		return AppendError(b, err)
	}
	return Append(fmter, b, value)
}
func msgpackAppender(_ reflect.Type) AppenderFunc {
	return func(fmter Formatter, b []byte, v reflect.Value) []byte {
		bs, err := msgp.Marshal(v.Interface(), msgp.SortedMapKeys)
		if err != nil {
			return AppendError(b, err)
		}
		return AppendBytes(b, bs)
	}
}
func arrayAppender(typ reflect.Type) AppenderFunc {
	elemType := typ.Elem()
	appendElem := Appender(elemType)
	return func(fmter Formatter, b []byte, v reflect.Value) []byte {
		kind := v.Kind()
		switch kind {
		case reflect.Ptr, reflect.Slice:
			if v.IsNil() {
				return AppendNull(b)
			}
		}
		if kind == reflect.Ptr {
			v = v.Elem()
		}
		b = append(b, '[')
		sliceLen := v.Len()
		for i := 0; i < sliceLen; i++ {
			elem := v.Index(i)
			b = appendElem(fmter, b, elem)
			b = append(b, ',')
		}
		if sliceLen > 0 {
			b[len(b)-1] = ']'
		} else {
			b = append(b, ']')
		}
		return b
	}
}
