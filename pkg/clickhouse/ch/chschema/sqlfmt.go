package chschema

import (
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"github.com/uptrace/pkg/unsafeconv"
	"reflect"
	"strings"
)

type QueryAppender interface {
	AppendQuery(fmter Formatter, b []byte) ([]byte, error)
}
type ColumnsAppender interface {
	AppendColumns(fmter Formatter, b []byte) ([]byte, error)
}
type Safe string

var _ QueryAppender = (*Safe)(nil)

func (s Safe) AppendQuery(fmter Formatter, b []byte) ([]byte, error) { return append(b, s...), nil }

type Name string

var _ QueryAppender = (*Name)(nil)

func (s Name) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return AppendName(b, string(s)), nil
}
func QuoteName(field string) string            { return string(AppendName(nil, field)) }
func AppendName(b []byte, field string) []byte { return appendName(b, unsafeconv.Bytes(field)) }
func appendName(buf, src []byte) []byte {
	const quote = '`'
	buf = append(buf, quote)
	for _, ch := range src {
		switch ch {
		case quote:
			buf = append(buf, quote, quote)
		case '\\':
			buf = append(buf, `\\`...)
		default:
			buf = append(buf, ch)
		}
	}
	buf = append(buf, quote)
	return buf
}

type Ident string

var _ QueryAppender = (*Ident)(nil)

func (s Ident) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return AppendIdent(b, string(s)), nil
}
func AppendIdent(b []byte, field string) []byte { return appendIdent(b, unsafeconv.Bytes(field)) }
func QuoteIdent(field string) string            { return string(AppendIdent(nil, field)) }
func appendIdent(buf, src []byte) []byte {
	const quote = '`'
	var quoted bool
loop:
	for _, ch := range src {
		switch ch {
		case '*':
			if !quoted {
				buf = append(buf, '*')
				continue loop
			}
		case '.':
			if quoted {
				buf = append(buf, quote)
				quoted = false
			}
			buf = append(buf, '.')
			continue loop
		}
		if !quoted {
			buf = append(buf, quote)
			quoted = true
		}
		switch ch {
		case quote:
			buf = append(buf, quote, quote)
		case '\\':
			buf = append(buf, `\\`...)
		default:
			buf = append(buf, ch)
		}
	}
	if quoted {
		buf = append(buf, quote)
	}
	return buf
}

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
	return QueryWithArgs{Query: query, Args: args}
}
func UnsafeName(ident string) QueryWithArgs { return QueryWithArgs{Query: ident} }
func (q QueryWithArgs) IsZero() bool        { return q.Query == "" && q.Args == nil }
func (q QueryWithArgs) IsEmpty() bool       { return q.Query == "" }
func (q QueryWithArgs) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	if q.Args == nil {
		return AppendName(b, q.Query), nil
	}
	return fmter.AppendQuery(b, q.Query, q.Args...), nil
}
func (q QueryWithArgs) Value() Safe { b, _ := q.AppendQuery(NopFmter, nil); return Safe(b) }

type QueryWithSep struct {
	QueryWithArgs
	Sep string
}

func SafeQueryWithSep(query string, args []any, sep string) QueryWithSep {
	return QueryWithSep{QueryWithArgs: SafeQuery(query, args), Sep: sep}
}

type ArrayValue struct{ v reflect.Value }

func Array(vi interface{}) *ArrayValue { return &ArrayValue{v: reflect.ValueOf(vi)} }

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
