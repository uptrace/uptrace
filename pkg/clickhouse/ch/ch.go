package ch

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/uptrace/pkg/clickhouse/ch/chpool"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
)

type (
	Safe             = chschema.Safe
	Name             = chschema.Name
	Ident            = chschema.Ident
	CHModel          = chschema.CHModel
	AfterScanRowHook = chschema.AfterScanRowHook
)

var ErrClosed = chpool.ErrClosed

func SafeQuery(query string, args ...any) chschema.QueryWithArgs {
	return chschema.SafeQuery(query, args)
}

type AnyValue struct{ Any any }

func SafeValue(v any) AnyValue { return AnyValue{v} }
func (v AnyValue) AppendQuery(fmter chschema.Formatter, buf []byte) ([]byte, error) {
	return chschema.Append(fmter, buf, v.Any), nil
}

type Result struct {
	model    Model
	affected int64
	progress Progress
}

var _ sql.Result = (*Result)(nil)

func NewResult() *Result                         { return new(Result) }
func (res *Result) Model() Model                 { return res.model }
func (res *Result) RowsAffected() (int64, error) { return int64(res.affected), nil }
func (res *Result) LastInsertId() (int64, error) { return 0, errors.New("not implemented") }
func (res *Result) Progress() *Progress          { return &res.progress }

type Error struct {
	Code       int32
	Name       string
	Message    string
	StackTrace string
	Nested     error
	Result     *Result
}

func (err *Error) Error() string { return fmt.Sprintf("%s: %s (%d)", err.Name, err.Message, err.Code) }
func (err *Error) Timeout() bool {
	const (
		timeoutExceeded = 159
		tooSlow         = 160
	)
	switch err.Code {
	case timeoutExceeded, tooSlow:
		return true
	default:
		return false
	}
}
func isBadConn(err error, allowTimeout bool) bool {
	if err == nil {
		return false
	}
	if allowTimeout {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return !netErr.Temporary()
		}
	}
	return true
}
func RegisterEnum(name string, ss []string) { chschema.RegisterEnum(name, ss) }

type ListValues struct{ slice any }

var _ chschema.QueryAppender = ListValues{}

func List(slice any) ListValues { return ListValues{slice: slice} }
func (in ListValues) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	v := reflect.ValueOf(in.slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("ch: In(non-slice %T)", in.slice)
	}
	b = appendList(fmter, b, v)
	return b, nil
}

type InValues struct{ slice any }

var _ chschema.QueryAppender = InValues{}

func In(slice any) InValues { return InValues{slice: slice} }
func (in InValues) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	v := reflect.ValueOf(in.slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("ch: In(non-slice %T)", in.slice)
	}
	b = append(b, '(')
	b = appendList(fmter, b, v)
	b = append(b, ')')
	return b, nil
}

type ArrayValues struct{ slice any }

var _ chschema.QueryAppender = ArrayValues{}

func Array(slice any) ArrayValues { return ArrayValues{slice: slice} }
func (in ArrayValues) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	v := reflect.ValueOf(in.slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("ch: Array(non-slice %T)", in.slice)
	}
	b = append(b, '[')
	b = appendList(fmter, b, v)
	b = append(b, ']')
	return b, nil
}
func appendList(fmter chschema.Formatter, b []byte, slice reflect.Value) []byte {
	sliceLen := slice.Len()
	if sliceLen == 0 {
		return append(b, "NULL"...)
	}
	for i := 0; i < sliceLen; i++ {
		if i > 0 {
			b = append(b, ", "...)
		}
		elem := slice.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		b = chschema.AppendValue(fmter, b, elem)
	}
	return b
}
