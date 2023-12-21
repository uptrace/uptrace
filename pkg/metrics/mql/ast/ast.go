package ast

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/slices"
)

type Expr interface {
	AppendString(b []byte) []byte
	AppendTemplate(b []byte) []byte
}

func String(expr Expr) string {
	return unsafeconv.String(expr.AppendString(nil))
}

var (
	_ Expr = (*ParenExpr)(nil)
	_ Expr = (*MetricExpr)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*FuncCall)(nil)
)

type Selector struct {
	Expr NamedExpr
}

type NamedExpr struct {
	Expr  Expr
	Alias string
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

func (e ParenExpr) AppendTemplate(b []byte) []byte {
	b = append(b, '(')
	b = e.Expr.AppendTemplate(b)
	b = append(b, ')')
	return b
}

type MetricExpr struct {
	Name     string
	Filters  []Filter
	Grouping GroupingElems
}

func (n *MetricExpr) AppendString(b []byte) []byte {
	b = append(b, n.Name...)

	if len(n.Filters) > 0 {
		b = append(b, '{')
		for i := range n.Filters {
			if i > 0 {
				b = append(b, ',')
			}
			b = n.Filters[i].AppendString(b)
		}
		b = append(b, '}')
	}

	return b
}

func (n *MetricExpr) AppendTemplate(b []byte) []byte {
	b = append(b, n.Name...)
	b = append(b, "$$"...)
	return b
}

type NumberKind int

const (
	NumberUnitless NumberKind = iota
	NumberDuration
	NumberBytes
)

type Number struct {
	Text string
	Kind NumberKind
}

func (n Number) String() string {
	return n.Text
}

func (n Number) AppendString(b []byte) []byte {
	return append(b, n.Text...)
}

func (n Number) AppendTemplate(b []byte) []byte {
	return append(b, n.Text...)
}

func (n Number) ConvertValue(unit string) (float64, error) {
	switch n.Kind {
	case NumberDuration:
		dur, err := time.ParseDuration(n.Text)
		if err != nil {
			return 0, err
		}
		return bunconv.ConvertValue(float64(dur), bunconv.UnitNanoseconds, unit)
	case NumberBytes:
		bytes, err := bunconv.ParseBytes(n.Text)
		if err != nil {
			return 0, err
		}
		return bunconv.ConvertValue(float64(bytes), bunconv.UnitBytes, unit)
	default:
		f, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	}
}

func (n Number) Float64() float64 {
	switch n.Kind {
	case NumberDuration:
		dur, err := time.ParseDuration(n.Text)
		if err != nil {
			panic(err)
		}
		return float64(dur)
	case NumberBytes:
		bytes, err := bunconv.ParseBytes(n.Text)
		if err != nil {
			panic(err)
		}
		return float64(bytes)
	default:
		f, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			panic(err)
		}
		return f
	}
}

type FuncCall struct {
	Func     string
	Arg      Expr
	Grouping GroupingElems
}

func (fn *FuncCall) AppendString(b []byte) []byte {
	b = append(b, fn.Func...)
	b = append(b, '(')
	b = fn.Arg.AppendString(b)
	b = append(b, ')')
	return b
}

func (fn *FuncCall) AppendTemplate(b []byte) []byte {
	b = append(b, fn.Func...)
	b = append(b, '(')
	b = fn.Arg.AppendTemplate(b)
	b = append(b, ')')
	return b
}

type UniqExpr struct {
	Name  MetricExpr
	Attrs []string
}

func (uq *UniqExpr) AppendString(b []byte) []byte {
	b = append(b, "uniq("...)
	for i, attr := range uq.Attrs {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = append(b, uq.Name.Name...)
		b = append(b, '.')
		b = append(b, attr...)
	}
	b = append(b, ')')
	return b
}

func (uq *UniqExpr) AppendTemplate(b []byte) []byte {
	return uq.AppendString(b)
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

func (e *BinaryExpr) AppendTemplate(b []byte) []byte {
	b = e.LHS.AppendTemplate(b)
	b = append(b, ' ')
	b = append(b, e.Op...)
	b = append(b, ' ')
	b = e.RHS.AppendTemplate(b)
	return b
}

//------------------------------------------------------------------------------

type Grouping struct {
	Elems []GroupingElem
}

type GroupingElems []GroupingElem

func (els GroupingElems) Attrs() []string {
	attrs := make([]string, len(els))
	for i, el := range els {
		attrs[i] = el.Alias
	}
	return attrs
}

type GroupingElem struct {
	Func  string
	Name  string
	Alias string
}

func (g GroupingElem) String() string {
	return unsafeconv.String(g.AppendString(nil))
}

func (g GroupingElem) AppendString(b []byte) []byte {
	if g.Func != "" {
		b = append(b, g.Func...)
		b = append(b, '(')
	}
	b = append(b, g.Name...)
	if g.Func != "" {
		b = append(b, ')')
	}
	if g.Alias != "" {
		b = append(b, " AS "...)
		b = append(b, g.Alias...)
	}
	return b
}

//------------------------------------------------------------------------------

type Where struct {
	Filters []Filter
}

type FilterOp string

const (
	FilterEqual     FilterOp = "="
	FilterNotEqual  FilterOp = "!="
	FilterLT        FilterOp = "<"
	FilterLTE       FilterOp = "<="
	FilterGT        FilterOp = ">"
	FilterGTE       FilterOp = ">="
	FilterIn        FilterOp = "in"
	FilterNotIn     FilterOp = "not in"
	FilterRegexp    FilterOp = "~"
	FilterNotRegexp FilterOp = "!~"
	FilterLike      FilterOp = "like"
	FilterNotLike   FilterOp = "not like"
	FilterExists    FilterOp = "exists"
	FilterNotExists FilterOp = "not exists"
)

type BoolOp string

const (
	BoolAnd BoolOp = "AND"
	BoolOr  BoolOp = "OR"
)

type Filter struct {
	BoolOp BoolOp
	LHS    string
	Op     FilterOp
	RHS    Value
}

type Value interface {
	AppendString(b []byte) []byte
}

func (f *Filter) String() string {
	b := make([]byte, 0, 100)
	b = f.AppendString(b)
	return unsafeconv.String(b)
}

func (f *Filter) AppendString(b []byte) []byte {
	b = append(b, f.LHS...)

	switch f.Op {
	case FilterEqual, FilterNotEqual, FilterRegexp, FilterNotRegexp:
		b = append(b, f.Op...)
	default:
		b = append(b, ' ')
		b = append(b, f.Op...)
		b = append(b, ' ')
	}

	if f.RHS != nil {
		b = f.RHS.AppendString(b)
	}

	return b
}

type StringValue struct {
	Text string
}

func (v StringValue) AppendString(b []byte) []byte {
	if IsIdent(v.Text) {
		return append(b, v.Text...)
	}
	return strconv.AppendQuote(b, v.Text)
}

type StringValues struct {
	Values []string
}

func (v StringValues) AppendString(b []byte) []byte {
	b = append(b, '(')
	for i, text := range v.Values {
		if i > 0 {
			b = append(b, ", "...)
		}
		if IsIdent(text) {
			b = append(b, text...)
		} else {
			b = strconv.AppendQuote(b, text)
		}
	}
	b = append(b, ')')
	return b
}

func SplitAliasName(s string) (string, string) {
	if s == "" {
		return "", ""
	}
	if s[0] != '$' {
		return "", s
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

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

//------------------------------------------------------------------------------

func clean(attrKey string) string {
	if strings.HasPrefix(attrKey, "span.") {
		return strings.TrimPrefix(attrKey, "span")
	}
	return attrKey
}

func applyGrouping(expr Expr, grouping []GroupingElem) {
	switch expr := expr.(type) {
	case *MetricExpr:
		expr.Grouping = append(expr.Grouping, grouping...)
	case *FuncCall:
		applyGrouping(expr.Arg, grouping)
		expr.Grouping = append(expr.Grouping, grouping...)
	case *UniqExpr:
		// nothing
	case *BinaryExpr:
		applyGrouping(expr.LHS, grouping)
		applyGrouping(expr.RHS, grouping)
	case ParenExpr:
		applyGrouping(expr.Expr, grouping)
	case Number:
		// nothing
	default:
		panic(fmt.Errorf("unsupported expr: %T", expr))
	}
}
