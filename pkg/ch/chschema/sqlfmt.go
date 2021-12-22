package chschema

import (
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

// FQN represents a fully qualified SQL name, for example, table or column name.
type FQN string

var _ QueryAppender = (*FQN)(nil)

func (s FQN) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return fmter.AppendIdent(b, string(s)), nil
}

func AppendFQN(b []byte, field string) []byte {
	return appendFQN(b, internal.Bytes(field))
}

func appendFQN(b, src []byte) []byte {
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

// Ident represents a SQL identifier, for example, table or column name.
type Ident string

var _ QueryAppender = (*Ident)(nil)

func (s Ident) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	return fmter.AppendIdent(b, string(s)), nil
}

func AppendIdent(b []byte, field string) []byte {
	return appendIdent(b, internal.Bytes(field))
}

func appendIdent(b, src []byte) []byte {
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

type QueryWithArgs struct {
	Query string
	Args  []any
}

var _ QueryAppender = (*QueryWithArgs)(nil)

func SafeQuery(query string, args []any) QueryWithArgs {
	if query != "" && args == nil {
		args = make([]any, 0)
	}
	return QueryWithArgs{Query: query, Args: args}
}

func UnsafeIdent(ident string) QueryWithArgs {
	return QueryWithArgs{Query: ident}
}

func (q QueryWithArgs) IsZero() bool {
	return q.Query == "" && q.Args == nil
}

func (q QueryWithArgs) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	if q.Args == nil {
		return fmter.AppendIdent(b, q.Query), nil
	}
	return fmter.AppendQuery(b, q.Query, q.Args...), nil
}

func (q QueryWithArgs) Value() Safe {
	return Safe(FormatQuery(q.Query, q.Args...))
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
