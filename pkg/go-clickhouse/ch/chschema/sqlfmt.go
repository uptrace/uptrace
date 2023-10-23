package chschema

import (
	"reflect"
	"strings"

	"github.com/uptrace/go-clickhouse/ch/internal"
)

type QueryAppender interface {
	AppendQuery(fmter Formatter, b []byte) ([]byte, error)
}

type ColumnsAppender interface {
	AppendColumns(fmter Formatter, b []byte) ([]byte, error)
}

//------------------------------------------------------------------------------

// Safe represents a safe SQL query.
type Safe string

var _ QueryAppender = (*Safe)(nil)

func (s Safe) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return append(b, s...), nil
}

//------------------------------------------------------------------------------

// Name represents a SQL identifier, for example, table or column name.
type Name string

var _ QueryAppender = (*Name)(nil)

func (s Name) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return fmter.AppendName(b, string(s)), nil
}

func AppendName(b []byte, field string) []byte {
	return appendName(b, internal.Bytes(field))
}

func appendName(b, src []byte) []byte {
	const quote = '"'

	b = append(b, quote)
	for _, c := range src {
		if c == quote {
			b = append(b, quote, quote)
		} else {
			b = append(b, c)
		}
	}
	b = append(b, quote)
	return b
}

//------------------------------------------------------------------------------

// Ident represents a fully qualified SQL name, for example, table or column name.
type Ident string

var _ QueryAppender = (*Name)(nil)

func (s Ident) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return fmter.AppendIdent(b, string(s)), nil
}

func AppendIdent(b []byte, field string) []byte {
	return appendIdent(b, internal.Bytes(field))
}

func appendIdent(b, src []byte) []byte {
	const quote = '"'

	var quoted bool
loop:
	for _, c := range src {
		switch c {
		case '*':
			if !quoted {
				b = append(b, '*')
				continue loop
			}
		case '.':
			if quoted {
				b = append(b, quote)
				quoted = false
			}
			b = append(b, '.')
			continue loop
		}

		if !quoted {
			b = append(b, quote)
			quoted = true
		}
		if c == quote {
			b = append(b, quote, quote)
		} else {
			b = append(b, c)
		}
	}
	if quoted {
		b = append(b, quote)
	}
	return b
}

//------------------------------------------------------------------------------

type QueryWithArgs struct {
	Query string
	Args  []any
}

var _ QueryAppender = (*QueryWithArgs)(nil)

func SafeQuery(query string, args []any) QueryWithArgs {
	if args == nil {
		args = make([]any, 0)
	} else if len(query) > 0 && strings.IndexByte(query, '?') == -1 {
		internal.Warn.Printf("query %q has %v args, but no placeholders", query, args)
	}
	return QueryWithArgs{
		Query: query,
		Args:  args,
	}
}

func UnsafeName(ident string) QueryWithArgs {
	return QueryWithArgs{Query: ident}
}

func (q QueryWithArgs) IsZero() bool {
	return q.Query == "" && q.Args == nil
}

func (q QueryWithArgs) IsEmpty() bool {
	return q.Query == ""
}

func (q QueryWithArgs) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	if q.Args == nil {
		return fmter.AppendName(b, q.Query), nil
	}
	return fmter.AppendQuery(b, q.Query, q.Args...), nil
}

func (q QueryWithArgs) Value() Safe {
	b, _ := q.AppendQuery(emptyFmter, nil)
	return Safe(b)
}

//------------------------------------------------------------------------------

type QueryWithSep struct {
	QueryWithArgs
	Sep string
}

func SafeQueryWithSep(query string, args []any, sep string) QueryWithSep {
	return QueryWithSep{
		QueryWithArgs: SafeQuery(query, args),
		Sep:           sep,
	}
}

//------------------------------------------------------------------------------

type ArrayValue struct {
	v reflect.Value
}

func Array(vi interface{}) *ArrayValue {
	return &ArrayValue{
		v: reflect.ValueOf(vi),
	}
}

var _ QueryAppender = (*ArrayValue)(nil)

func (a *ArrayValue) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	if !a.v.IsValid() || a.v.Len() == 0 {
		b = append(b, "[]"...)
		return b, nil
	}

	typ := a.v.Type()
	elemType := typ.Elem()
	appendElem := Appender(elemType)

	b = append(b, '[')

	ln := a.v.Len()
	for i := 0; i < ln; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		elem := a.v.Index(i)
		b = appendElem(fmter, b, elem)
	}

	b = append(b, ']')

	return b, nil
}
