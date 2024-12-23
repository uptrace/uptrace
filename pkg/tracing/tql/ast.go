package tql

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/pkg/unsafeconv"
)

type Expr interface {
	AppendString([]byte) []byte
}

var (
	_ Expr = (*Attr)(nil)
	_ Expr = (*NumberValue)(nil)
	_ Expr = (*FuncCall)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*ParenExpr)(nil)
)

func String(expr Expr) string {
	return unsafeconv.String(expr.AppendString(nil))
}

type AST interface {
	fmt.Stringer
}

type Selector struct {
	Columns
}

func (sel *Selector) String() string {
	var b []byte
	b = sel.Columns.AppendString(b)
	return unsafeconv.String(b)
}

type Grouping struct {
	Columns
}

func (g *Grouping) String() string {
	var b []byte
	b = append(b, "group by "...)
	b = g.Columns.AppendString(b)
	return unsafeconv.String(b)
}

type Columns []Column

func (cols Columns) AppendString(b []byte) []byte {
	for i := range cols {
		col := &cols[i]

		if i > 0 {
			b = append(b, ", "...)
		}
		b = col.AppendString(b)
	}
	return b
}

type Column struct {
	Value Expr
	Alias string
}

func (col *Column) AppendString(b []byte) []byte {
	b = col.Value.AppendString(b)
	if col.Alias != "" {
		b = append(b, " AS "...)
		b = append(b, col.Alias...)
	}
	return b
}

type Where struct {
	Filters []Filter
}

func (w *Where) String() string {
	var b []byte
	b = append(b, "where "...)
	for i := range w.Filters {
		f := &w.Filters[i]

		if i > 0 {
			b = append(b, ' ')
			b = append(b, f.BoolOp...)
			b = append(b, ' ')
		}
		b = f.AppendString(b)
	}
	return unsafeconv.String(b)
}

type ParenExpr struct {
	Expr
}

func (e ParenExpr) AppendString(b []byte) []byte {
	b = append(b, '(')
	b = e.Expr.AppendString(b)
	b = append(b, ')')
	return b
}

type Attr struct {
	Name string
}

func (a Attr) AppendString(b []byte) []byte {
	return append(b, a.Name...)
}

type BinaryExpr struct {
	Op       BinaryOp
	LHS, RHS Expr
}

type BinaryOp string

func (e *BinaryExpr) AppendString(b []byte) []byte {
	b = e.LHS.AppendString(b)
	b = append(b, ' ')
	b = append(b, e.Op...)
	b = append(b, ' ')
	b = e.RHS.AppendString(b)
	return b
}

type FuncCall struct {
	Func string
	Arg  Expr
}

func (c *FuncCall) AppendString(b []byte) []byte {
	b = append(b, c.Func...)
	b = append(b, '(')
	b = c.Arg.AppendString(b)
	b = append(b, ')')
	return b
}

//------------------------------------------------------------------------------

type Filter struct {
	BoolOp BoolOp
	LHS    Expr
	Op     FilterOp
	RHS    Value
}

func (f *Filter) AppendString(b []byte) []byte {
	b = f.LHS.AppendString(b)
	b = append(b, ' ')
	b = append(b, f.Op...)
	if f.RHS != nil {
		b = append(b, ' ')
		b = f.RHS.AppendString(b)
	}
	return b
}

type BoolOp string

const (
	BoolAnd BoolOp = "AND"
	BoolOr  BoolOp = "OR"
)

type FilterOp string

const (
	FilterEqual    FilterOp = "="
	FilterNotEqual FilterOp = "!="

	FilterIn    FilterOp = "in"
	FilterNotIn FilterOp = "not in"

	FilterLike    FilterOp = "like"
	FilterNotLike FilterOp = "not like"

	FilterContains    FilterOp = "contains"
	FilterNotContains FilterOp = "not contains"

	FilterExists    FilterOp = "exists"
	FilterNotExists FilterOp = "not exists"

	// For compatibility with metrics.
	FilterRegexp    FilterOp = "~"
	FilterNotRegexp FilterOp = "!~"
)

type Value interface {
	fmt.Stringer
	AppendString([]byte) []byte
	Values() []string
}

var (
	_ Value = (*StringValue)(nil)
	_ Value = (*StringValues)(nil)
	_ Value = (*NumberValue)(nil)
)

func NewValue(v any) Value {
	switch v := v.(type) {
	case string:
		return StringValue{Text: v}
	case float64:
		return NumberValue{
			Kind: NumberUnitless,
			Text: strconv.FormatFloat(v, 'f', -1, 64),
		}
	case json.Number:
		return NumberValue{
			Kind: NumberUnitless,
			Text: v.String(),
		}
	case []string:
		return StringValues{Strings: v}
	case []any:
		values := make([]string, len(v))
		for i, value := range v {
			values[i] = fmt.Sprint(value)
		}
		return StringValues{Strings: values}
	default:
		return StringValue{Text: fmt.Sprint(v)}
	}
}

type StringValue struct {
	Text string
}

func (v StringValue) String() string {
	return v.Text
}

func (v StringValue) AppendString(b []byte) []byte {
	return strconv.AppendQuote(b, v.Text)
}

func (v StringValue) Values() []string {
	return []string{v.Text}
}

type StringValues struct {
	Strings []string
}

func (v StringValues) String() string {
	return strings.Join(v.Strings, "|")
}

func (v StringValues) AppendString(b []byte) []byte {
	b = append(b, '(')
	for i, text := range v.Strings {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = strconv.AppendQuote(b, text)
	}
	b = append(b, ')')
	return b
}

func (v StringValues) Values() []string {
	return v.Strings
}

type NumberKind int

const (
	NumberUnitless NumberKind = iota
	NumberDuration
	NumberBytes
)

type NumberValue struct {
	Kind NumberKind
	Text string
}

func (n NumberValue) String() string {
	return n.Text
}

func (v NumberValue) Values() []string {
	return []string{v.Text}
}

func (n NumberValue) AppendString(b []byte) []byte {
	return append(b, n.Text...)
}

func clean(attrKey string) string {
	if strings.HasPrefix(attrKey, "span.") {
		return strings.TrimPrefix(attrKey, "span")
	}
	return attrKey
}

//------------------------------------------------------------------------------

var opPrecedence = [][]BinaryOp{
	{"^"},
	{"*", "/", "%"},
	{"+", "-"},
	{"+", "-"},
	{"==", "!=", "<=", "<", ">=", ">"},
	{"and", "unless"},
	{"or"},
}

func binaryExprPrecedence(expr *BinaryExpr) *BinaryExpr {
	for _, ops := range opPrecedence {
		expr = unwrapBinaryExpr(exprPrecedence(expr, ops))
	}
	return expr
}

func exprPrecedence(anyexpr Expr, ops []BinaryOp) Expr {
	expr, ok := anyexpr.(*BinaryExpr)
	if !ok {
		return anyexpr
	}

	if slices.Index(ops, expr.Op) == -1 {
		expr.RHS = exprPrecedence(expr.RHS, ops)
		return expr
	}

	switch rhs := expr.RHS.(type) {
	case *BinaryExpr:
		expr = &BinaryExpr{
			Op: rhs.Op,
			LHS: ParenExpr{
				Expr: &BinaryExpr{
					Op:  expr.Op,
					LHS: expr.LHS,
					RHS: rhs.LHS,
				},
			},
			RHS: rhs.RHS,
		}
		expr = unwrapBinaryExpr(exprPrecedence(expr, ops))
		expr.RHS = exprPrecedence(expr.RHS, ops)
		return expr
	case ParenExpr:
		return expr
	default:
		return ParenExpr{Expr: expr}
	}
}

func unwrapBinaryExpr(expr Expr) *BinaryExpr {
	switch expr := expr.(type) {
	case *BinaryExpr:
		return expr
	case ParenExpr:
		return unwrapBinaryExpr(expr.Expr)
	default:
		panic("not reached")
	}
}
