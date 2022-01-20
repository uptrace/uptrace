package ch

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type (
	Safe             = chschema.Safe
	Ident            = chschema.Ident
	CHModel          = chschema.CHModel
	AfterScanRowHook = chschema.AfterScanRowHook
)

func SafeQuery(query string, args ...any) chschema.QueryWithArgs {
	return chschema.SafeQuery(query, args)
}

//------------------------------------------------------------------------------

type result struct {
	model    Model
	affected int
}

var _ sql.Result = (*result)(nil)

func (res *result) Model() Model {
	return res.model
}

func (res *result) RowsAffected() (int64, error) {
	return int64(res.affected), nil
}

func (res *result) LastInsertId() (int64, error) {
	return 0, errors.New("not implemented")
}

//------------------------------------------------------------------------------

type Error struct {
	Code       int32
	Name       string
	Message    string
	StackTrace string
	nested     error // TODO: wrap/unwrap
}

func (exc *Error) Error() string {
	return exc.Name + ": " + exc.Message
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

//------------------------------------------------------------------------------

type InValues struct {
	slice reflect.Value
	err   error
}

var _ chschema.QueryAppender = InValues{}

func In(slice any) InValues {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return InValues{
			err: fmt.Errorf("ch: In(non-slice %T)", slice),
		}
	}
	return InValues{
		slice: v,
	}
}

func (in InValues) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if in.err != nil {
		return nil, in.err
	}
	return appendIn(fmter, b, in.slice), nil
}

func appendIn(fmter chschema.Formatter, b []byte, slice reflect.Value) []byte {
	sliceLen := slice.Len()
	for i := 0; i < sliceLen; i++ {
		if i > 0 {
			b = append(b, ", "...)
		}

		elem := slice.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		if elem.Kind() == reflect.Slice {
			b = append(b, '(')
			b = appendIn(fmter, b, elem)
			b = append(b, ')')
		} else {
			b = chschema.AppendValue(fmter, b, elem)
		}
	}
	return b
}
