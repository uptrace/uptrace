package ast

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/pkg/unsafeconv"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/xhit/go-str2duration/v2"
)

type Expr interface {
	AppendString(b []byte) []byte
	AppendTemplate(b []byte) []byte
}

var (
	_ Expr = (*ParenExpr)(nil)
	_ Expr = (*MetricExpr)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*FuncCall)(nil)
	_ Expr = (*Number)(nil)
)

func String(expr Expr) string {
	return unsafeconv.String(expr.AppendString(nil))
}

type QueryPart interface {
	fmt.Stringer
}

var (
	_ QueryPart = (*Selector)(nil)
	_ QueryPart = (*Grouping)(nil)
	_ QueryPart = (*Where)(nil)
)

type Selector struct {
	Expr NamedExpr
}

func (sel *Selector) String() string {
	var b []byte
	b = sel.Expr.AppendString(b)
	return unsafeconv.String(b)
}

type NamedExpr struct {
	Expr     Expr
	HasAlias bool
	Alias    string
}

func (e *NamedExpr) AppendString(b []byte) []byte {
	b = e.Expr.AppendString(b)
	if e.HasAlias {
		b = append(b, " as "...)
		b = append(b, e.Alias...)
	}
	return b
}

func defaultAliasForExpr(expr Expr) string {
	switch expr := expr.(type) {
	case *MetricExpr:
		if len(expr.Filters) == 0 && strings.HasPrefix(expr.Name, "$") {
			return strings.TrimPrefix(expr.Name, "$")
		}
	}
	return String(expr)
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
	Name         string
	Filters      []Filter
	RollupWindow time.Duration
	Offset       time.Duration
	Grouping     GroupingElems
}

func (me *MetricExpr) AppendString(b []byte) []byte {
	b = append(b, me.Name...)

	if len(me.Filters) > 0 {
		b = append(b, '{')
		for i := range me.Filters {
			if i > 0 {
				b = append(b, ',')
			}
			b = me.Filters[i].AppendString(b, false)
		}
		b = append(b, '}')
	}

	if me.RollupWindow != 0 {
		b = append(b, '[')
		b = append(b, str2duration.String(me.RollupWindow)...)
		b = append(b, ']')
	}

	if len(me.Grouping) > 0 {
		b = append(b, " by ("...)
		b = me.Grouping.AppendString(b)
		b = append(b, ')')
	}

	if me.Offset != 0 {
		b = append(b, " offset "...)
		b = append(b, str2duration.String(me.Offset)...)
	}

	return b
}

func (me *MetricExpr) AppendTemplate(b []byte) []byte {
	if len(me.Filters) == 0 && strings.HasPrefix(me.Name, "$") {
		b = append(b, me.Name[1:]...)
	} else {
		b = append(b, me.Name...)
	}
	b = append(b, "$$"...)

	if me.Offset != 0 {
		b = append(b, " offset "...)
		b = append(b, str2duration.String(me.Offset)...)
	}

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
	if len(fn.Grouping) > 0 {
		b = append(b, " by ("...)
		b = fn.Grouping.AppendString(b)
		b = append(b, ')')
	}
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
	Name  *MetricExpr
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
	b = append(b, "uniq("...)
	b = uq.Name.AppendTemplate(b)
	for _, attr := range uq.Attrs {
		b = append(b, ", "...)
		b = append(b, attr...)
	}
	b = append(b, ')')
	return b
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
	Elems GroupingElems
}

func (g *Grouping) String() string {
	var b []byte
	b = append(b, "group by "...)
	b = g.Elems.AppendString(b)
	return unsafeconv.String(b)
}

type GroupingElems []GroupingElem

func (els GroupingElems) AppendString(b []byte) []byte {
	for i := range els {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = els[i].AppendString(b)
	}
	return b
}

func (els GroupingElems) Attrs() []string {
	attrs := make([]string, len(els))
	for i, el := range els {
		attrs[i] = el.Alias
	}
	return attrs
}

type GroupingElem struct {
	Func     string
	Name     string
	HasAlias bool
	Alias    string
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
	if g.HasAlias {
		b = append(b, " as "...)
		b = append(b, g.Alias...)
	}
	return b
}

func defaultAliasForGrouping(elem *GroupingElem) string {
	b := elem.AppendString(nil)
	return unsafeconv.String(b)
}

//------------------------------------------------------------------------------

type Where struct {
	Filters Filters
}

func (w *Where) String() string {
	var b []byte
	b = append(b, "where "...)
	b = w.Filters.AppendString(b)
	return unsafeconv.String(b)
}

type Filters []Filter

func (filters Filters) AppendString(b []byte) []byte {
	for i := range filters {
		f := &filters[i]

		if i > 0 {
			b = append(b, ' ')
			if f.BoolOp != "" {
				b = append(b, f.BoolOp...)
			} else {
				b = append(b, BoolAnd...)
			}
			b = append(b, ' ')
		}
		b = f.AppendString(b, true)
	}
	return b
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

func (f *Filter) AppendString(b []byte, spaceAround bool) []byte {
	b = append(b, f.LHS...)

	switch f.Op {
	case FilterEqual, FilterNotEqual, FilterRegexp, FilterNotRegexp:
		if spaceAround {
			b = append(b, ' ')
		}
		b = append(b, f.Op...)
		if spaceAround && f.RHS != nil {
			b = append(b, ' ')
		}
	default:
		b = append(b, ' ')
		b = append(b, f.Op...)
		if f.RHS != nil {
			b = append(b, ' ')
		}
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
		b = strconv.AppendQuote(b, text)
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
