package chschema

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/uptrace/go-clickhouse/ch/internal"
	"github.com/uptrace/go-clickhouse/ch/internal/parser"
)

var emptyFmter Formatter

func FormatQuery(query string, args ...any) string {
	return emptyFmter.FormatQuery(query, args...)
}

func AppendQuery(b []byte, query string, args ...any) []byte {
	return emptyFmter.AppendQuery(b, query, args...)
}

type Formatter struct {
	args *namedArgList
}

func NewFormatter() Formatter {
	return Formatter{}
}

func (f Formatter) AppendIdent(b []byte, ident string) []byte {
	return AppendIdent(b, ident)
}

func (f Formatter) WithArg(arg NamedArgAppender) Formatter {
	return Formatter{
		args: f.args.WithArg(arg),
	}
}

func (f Formatter) WithNamedArg(name string, value any) Formatter {
	return Formatter{
		args: f.args.WithArg(&namedArg{name: name, value: value}),
	}
}

func (f Formatter) FormatQuery(query string, args ...any) string {
	if (args == nil && f.args == nil) || strings.IndexByte(query, '?') == -1 {
		return query
	}
	return internal.String(f.AppendQuery(nil, query, args...))
}

func (f Formatter) AppendQuery(b []byte, query string, args ...any) []byte {
	if (args == nil && f.args == nil) || strings.IndexByte(query, '?') == -1 {
		return append(b, query...)
	}
	return f.append(b, parser.NewString(query), args)
}

func (f Formatter) append(dst []byte, p *parser.Parser, args []any) []byte {
	var namedArgs NamedArgAppender
	if len(args) == 1 {
		if v, ok := args[0].(NamedArgAppender); ok {
			namedArgs = v
		} else if v, ok := newStructArgs(f, args[0]); ok {
			namedArgs = v
		}
	}

	var argIndex int
	for p.Valid() {
		b, ok := p.ReadSep('?')
		if !ok {
			dst = append(dst, b...)
			continue
		}
		if len(b) > 0 && b[len(b)-1] == '\\' {
			dst = append(dst, b[:len(b)-1]...)
			dst = append(dst, '?')
			continue
		}
		dst = append(dst, b...)

		name, numeric := p.ReadIdentifier()
		if name != "" {
			if numeric {
				idx, err := strconv.Atoi(name)
				if err != nil {
					goto restore_arg
				}

				if idx >= len(args) {
					goto restore_arg
				}

				dst = f.appendArg(dst, args[idx])
				continue
			}

			if namedArgs != nil {
				dst, ok = namedArgs.AppendNamedArg(f, dst, name)
				if ok {
					continue
				}
			}

			dst, ok = f.args.AppendNamedArg(f, dst, name)
			if ok {
				continue
			}

		restore_arg:
			dst = append(dst, '?')
			dst = append(dst, name...)
			continue
		}

		if argIndex >= len(args) {
			dst = append(dst, '?')
			continue
		}

		arg := args[argIndex]
		argIndex++

		dst = f.appendArg(dst, arg)
	}

	return dst
}

func (f Formatter) appendArg(b []byte, arg any) []byte {
	switch arg := arg.(type) {
	case QueryAppender:
		bb, err := arg.AppendQuery(f, b)
		if err != nil {
			return AppendError(b, err)
		}
		return bb
	default:
		return Append(f, b, arg)
	}
}

//------------------------------------------------------------------------------

type NamedArgAppender interface {
	AppendNamedArg(fmter Formatter, b []byte, name string) ([]byte, bool)
}

type namedArgList struct {
	arg  NamedArgAppender
	next *namedArgList
}

func (l *namedArgList) WithArg(arg NamedArgAppender) *namedArgList {
	return &namedArgList{
		arg:  arg,
		next: l,
	}
}

func (l *namedArgList) AppendNamedArg(fmter Formatter, b []byte, name string) ([]byte, bool) {
	for l != nil && l.arg != nil {
		if b, ok := l.arg.AppendNamedArg(fmter, b, name); ok {
			return b, true
		}
		l = l.next
	}
	return b, false
}

//------------------------------------------------------------------------------

type namedArg struct {
	name  string
	value any
}

var _ NamedArgAppender = (*namedArg)(nil)

func (a *namedArg) AppendNamedArg(fmter Formatter, b []byte, name string) ([]byte, bool) {
	if a.name == name {
		return fmter.appendArg(b, a.value), true
	}
	return b, false
}

//------------------------------------------------------------------------------

type structArgs struct {
	table *Table
	strct reflect.Value
}

var _ NamedArgAppender = (*structArgs)(nil)

func newStructArgs(fmter Formatter, strct any) (*structArgs, bool) {
	v := reflect.ValueOf(strct)
	if !v.IsValid() {
		return nil, false
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	return &structArgs{
		table: TableForType(v.Type()),
		strct: v,
	}, true
}

func (m *structArgs) AppendNamedArg(fmter Formatter, b []byte, name string) ([]byte, bool) {
	return m.table.AppendNamedArg(fmter, b, name, m.strct)
}
