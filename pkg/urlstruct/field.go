package urlstruct

import (
	"database/sql"
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/pkg/idgen"
	"github.com/uptrace/pkg/unixtime"
	"github.com/vmihailenco/tagparser"
)

type Field struct {
	Type      reflect.Type
	Name      string
	Index     []int
	Tag       *tagparser.Tag
	noDecode  bool
	scanValue scannerFunc
}

func (f *Field) init() {
	_, f.noDecode = f.Tag.Options["nodecode"]
	if f.Type.Kind() == reflect.Slice {
		f.scanValue = sliceScanner(f.Type)
	} else {
		f.scanValue = scanner(f.Type)
	}
}
func (f *Field) Value(strct reflect.Value) reflect.Value { return strct.FieldByIndex(f.Index) }

var (
	textUnmarshalerType      = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	timeType                 = reflect.TypeOf((*time.Time)(nil)).Elem()
	nanoTimeType             = reflect.TypeFor[unixtime.Nano]()
	durationType             = reflect.TypeOf((*time.Duration)(nil)).Elem()
	nullBoolType             = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	nullInt64Type            = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	nullFloat64Type          = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	nullStringType           = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	traceIDType              = reflect.TypeOf((*idgen.TraceID)(nil)).Elem()
	spanIDType               = reflect.TypeFor[idgen.SpanID]()
	mapStringStringType      = reflect.TypeOf((*map[string]string)(nil)).Elem()
	mapStringStringSliceType = reflect.TypeOf((*map[string][]string)(nil)).Elem()
)

type scannerFunc func(v reflect.Value, values []string) error

func scanner(typ reflect.Type) scannerFunc {
	switch typ {
	case timeType:
		return scanTime
	case nanoTimeType:
		return scanNanoTime
	}
	if typ.Implements(textUnmarshalerType) {
		return scanTextUnmarshaler
	}
	if reflect.PtrTo(typ).Implements(textUnmarshalerType) {
		return scanTextUnmarshalerAddr
	}
	switch typ {
	case durationType:
		return scanDuration
	case nullBoolType:
		return scanNullBool
	case nullInt64Type:
		return scanNullInt64
	case nullFloat64Type:
		return scanNullFloat64
	case nullStringType:
		return scanNullString
	case traceIDType:
		return scanTraceID
	case spanIDType:
		return scanSpanID
	case mapStringStringType:
		return scanMapStringString
	case mapStringStringSliceType:
		return scanMapStringStringSlice
	}
	switch typ.Kind() {
	case reflect.Bool:
		return scanBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return scanInt64
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return scanUint64
	case reflect.Float32:
		return scanFloat32
	case reflect.Float64:
		return scanFloat64
	case reflect.String:
		return scanString
	}
	return nil
}
func sliceScanner(typ reflect.Type) scannerFunc {
	switch typ.Elem().Kind() {
	case reflect.Int:
		return scanIntSlice
	case reflect.Int32:
		return scanInt32Slice
	case reflect.Int64:
		return scanInt64Slice
	case reflect.String:
		return scanStringSlice
	}
	if elementScanner := scanner(typ.Elem()); elementScanner != nil {
		return func(v reflect.Value, values []string) error {
			nn := reflect.MakeSlice(typ, 0, len(values))
			for _, s := range values {
				n := reflect.New(typ.Elem())
				err := elementScanner(n.Elem(), []string{s})
				if err != nil {
					return err
				}
				nn = reflect.Append(nn, n.Elem())
			}
			v.Set(nn)
			return nil
		}
	}
	return nil
}
func scanTextUnmarshaler(v reflect.Value, values []string) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	u := v.Interface().(encoding.TextUnmarshaler)
	return u.UnmarshalText([]byte(values[0]))
}
func scanTextUnmarshalerAddr(v reflect.Value, values []string) error {
	if !v.CanAddr() {
		return fmt.Errorf("urlstruct: Scan(nonsettable %s)", v.Type())
	}
	u := v.Addr().Interface().(encoding.TextUnmarshaler)
	return u.UnmarshalText([]byte(values[0]))
}
func scanBool(v reflect.Value, values []string) error {
	f, err := strconv.ParseBool(values[0])
	if err != nil {
		return err
	}
	v.SetBool(f)
	return nil
}
func scanInt64(v reflect.Value, values []string) error {
	s := values[0]
	if s == "" {
		v.SetInt(0)
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	v.SetInt(n)
	return nil
}
func scanUint64(v reflect.Value, values []string) error {
	s := values[0]
	if s == "" {
		v.SetUint(0)
		return nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	v.SetUint(n)
	return nil
}
func scanFloat32(v reflect.Value, values []string) error { return scanFloat(v, values, 32) }
func scanFloat64(v reflect.Value, values []string) error { return scanFloat(v, values, 64) }
func scanFloat(v reflect.Value, values []string, bits int) error {
	s := values[0]
	if s == "" {
		v.SetFloat(0)
		return nil
	}
	n, err := strconv.ParseFloat(values[0], bits)
	if err != nil {
		return err
	}
	v.SetFloat(n)
	return nil
}
func scanString(v reflect.Value, values []string) error {
	v.SetString(values[0])
	return nil
}
func scanTime(v reflect.Value, values []string) error {
	tm, err := parseTime(values[0])
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(tm))
	return nil
}
func scanNanoTime(v reflect.Value, values []string) error {
	tm, err := parseTime(values[0])
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(unixtime.ToNano(tm)))
	return nil
}
func parseTime(s string) (time.Time, error) {
	secs, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return time.Unix(secs, 0), nil
	}
	if len(s) >= 5 && s[4] == '-' {
		return time.Parse(time.RFC3339Nano, s)
	}
	if len(s) == 15 {
		const basicFormat = "20060102T150405"
		return time.Parse(basicFormat, s)
	}
	const basicFormat = "20060102T150405-07:00"
	return time.Parse(basicFormat, s)
}
func scanDuration(v reflect.Value, values []string) error {
	dur, err := parseDuration(values[0])
	if err != nil {
		return err
	}
	v.SetInt(int64(dur))
	return nil
}
func parseDuration(val string) (time.Duration, error) {
	if ms, err := strconv.ParseFloat(val, 64); err == nil {
		return time.Duration(float64(ms) * float64(time.Millisecond)), nil
	}
	return time.ParseDuration(val)
}
func scanTraceID(v reflect.Value, values []string) error {
	u, err := idgen.ParseTraceID(values[0])
	if err != nil {
		return err
	}
	v.Addr().SetBytes(u[:])
	return nil
}
func scanSpanID(v reflect.Value, values []string) error {
	id, err := idgen.ParseSpanID(values[0])
	if err != nil {
		return err
	}
	v.SetUint(uint64(id))
	return nil
}
func scanNullBool(v reflect.Value, values []string) error {
	value := sql.NullBool{Valid: true}
	s := values[0]
	if s == "" {
		v.Set(reflect.ValueOf(value))
		return nil
	}
	f, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	value.Bool = f
	v.Set(reflect.ValueOf(value))
	return nil
}
func scanNullInt64(v reflect.Value, values []string) error {
	value := sql.NullInt64{Valid: true}
	s := values[0]
	if s == "" {
		v.Set(reflect.ValueOf(value))
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	value.Int64 = n
	v.Set(reflect.ValueOf(value))
	return nil
}
func scanNullFloat64(v reflect.Value, values []string) error {
	value := sql.NullFloat64{Valid: true}
	s := values[0]
	if s == "" {
		v.Set(reflect.ValueOf(value))
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	value.Float64 = n
	v.Set(reflect.ValueOf(value))
	return nil
}
func scanNullString(v reflect.Value, values []string) error {
	value := sql.NullString{Valid: true}
	s := values[0]
	if s == "" {
		v.Set(reflect.ValueOf(value))
		return nil
	}
	value.String = s
	v.Set(reflect.ValueOf(value))
	return nil
}
func scanMapStringString(v reflect.Value, values []string) error {
	if len(values) == 1 {
		s := values[0]
		if s == "" {
			v.Set(reflect.New(mapStringStringType).Elem())
			return nil
		}
		var m map[string]string
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(m))
		return nil
	}
	m := make(map[string]string, len(values)/3)
	var key string
	for _, value := range values {
		if key == "" {
			key = value
			continue
		}
		if value == endOfValues {
			key = ""
			continue
		}
		if _, ok := m[key]; !ok {
			m[key] = value
		}
	}
	v.Set(reflect.ValueOf(m))
	return nil
}
func scanMapStringStringSlice(v reflect.Value, values []string) error {
	if len(values) == 1 {
		s := values[0]
		if s == "" {
			v.Set(reflect.New(mapStringStringSliceType).Elem())
			return nil
		}
		var m map[string][]string
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(m))
		return nil
	}
	m := make(map[string][]string, len(values)/3)
	var key string
	for _, value := range values {
		if key == "" {
			key = value
			continue
		}
		if value == endOfValues {
			key = ""
			continue
		}
		m[key] = append(m[key], value)
	}
	v.Set(reflect.ValueOf(m))
	return nil
}
func scanIntSlice(v reflect.Value, values []string) error {
	nn := make([]int, 0, len(values))
	for _, s := range values {
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		nn = append(nn, n)
	}
	v.Set(reflect.ValueOf(nn))
	return nil
}
func scanInt32Slice(v reflect.Value, values []string) error {
	nn := make([]int32, 0, len(values))
	for _, s := range values {
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}
		nn = append(nn, int32(n))
	}
	v.Set(reflect.ValueOf(nn))
	return nil
}
func scanInt64Slice(v reflect.Value, values []string) error {
	nn := make([]int64, 0, len(values))
	for _, s := range values {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		nn = append(nn, n)
	}
	v.Set(reflect.ValueOf(nn))
	return nil
}
func scanStringSlice(v reflect.Value, values []string) error {
	v.Set(reflect.ValueOf(values))
	return nil
}
