package ast

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type Selector struct {
	Expr       NamedExpr
	Grouping   []string
	GroupByAll bool
}

type NamedExpr struct {
	Expr  Expr // *Name | *BinaryExpr | *FuncCall
	Alias string
}

type ParenExpr struct {
	Expr
}

func (e ParenExpr) String() string {
	return "(" + e.Expr.String() + ")"
}

type Expr interface {
	fmt.Stringer
}

type Name struct {
	Func    string
	Name    string
	Filters []Filter
}

func (n *Name) String() string {
	var b []byte

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

	return unsafeconv.String(b)
}

type Number struct {
	Text string
	Kind ValueKind
}

func (n *Number) String() string {
	return n.Text
}

func (n *Number) Float32() float32 {
	return float32(n.Float64())
}

func (n *Number) Float64() float64 {
	switch n.Kind {
	case ValueDuration:
		dur, err := time.ParseDuration(n.Text)
		if err != nil {
			panic(err)
		}
		return float64(dur)
	case ValueBytes:
		bytes, err := bununit.ParseBytes(n.Text)
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
	Func string
	Args []Expr
}

func (fn *FuncCall) String() string {
	args := make([]string, len(fn.Args))
	for i, arg := range fn.Args {
		args[i] = arg.String()
	}
	return fn.Func + "(" + strings.Join(args, ", ") + ")"
}

type BinaryExpr struct {
	Op       BinaryOp
	LHS, RHS Expr
}

func (e *BinaryExpr) String() string {
	return e.LHS.String() + " " + string(e.Op) + " " + e.RHS.String()
}

type BinaryOp string

//------------------------------------------------------------------------------

type Grouping struct {
	Names      []string
	GroupByAll bool
}

//------------------------------------------------------------------------------

type Where struct {
	Filters []Filter
}

type FilterOp string

const (
	FilterEqual     FilterOp = "="
	FilterNotEqual  FilterOp = "!="
	FilterRegexp    FilterOp = "~"
	FilterNotRegexp FilterOp = "!~"
	FilterLike      FilterOp = "like"
	FilterNotLike   FilterOp = "not like"
)

type BoolOp string

const (
	BoolAnd BoolOp = "AND"
	BoolOr  BoolOp = "OR"
)

type Filter struct {
	Sep BoolOp
	LHS string
	Op  FilterOp
	RHS Value
}

func (f *Filter) String() string {
	b := make([]byte, len(f.LHS)+len(f.Op)+len(f.RHS.Text))
	b = f.AppendString(b)
	return unsafeconv.String(b)
}

func (f *Filter) AppendString(b []byte) []byte {
	b = append(b, f.LHS...)

	switch f.Op {
	case FilterLike, FilterNotLike:
		b = append(b, ' ')
		b = append(b, f.Op...)
		b = append(b, ' ')
	default:
		b = append(b, f.Op...)
	}

	if isIdent(f.RHS.Text) {
		b = append(b, f.RHS.Text...)
	} else {
		b = strconv.AppendQuote(b, f.RHS.Text)
	}

	return b
}

type Value struct {
	Text string
	Kind ValueKind
}

type ValueKind int

const (
	ValueText ValueKind = iota
	ValueDuration
	ValueBytes
)

func (v *Value) Value(unit string) (any, error) {
	switch v.Kind {
	case ValueDuration:
		dur, err := time.ParseDuration(v.Text)
		if err != nil {
			return nil, err
		}
		return bununit.ConvertValue(float64(dur), bununit.Nanoseconds, unit)
	case ValueBytes:
		bytes, err := bununit.ParseBytes(v.Text)
		if err != nil {
			return nil, err
		}
		return bununit.ConvertValue(float64(bytes), bununit.Bytes, unit)
	default:
		return v.Text, nil
	}
}

//------------------------------------------------------------------------------

// SplitAliasName splits metric alias and attr name.
// Alias must start with the `$` sign.
func SplitAliasName(s string) (string, string) {
	if s == "" {
		return "", ""
	}
	if s[0] != '$' {
		return "", s
	}
	s = strings.TrimPrefix(s, "$")
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, s
}

func Alias(s string) string {
	alias, _ := SplitAliasName(s)
	return alias
}

//------------------------------------------------------------------------------

var opPrecedence = []BinaryOp{
	"^",
	"*",
	"/",
	"%",
	"+",
	"-",
	"==",
	"!=",
	"<=",
	"<",
	">=",
	">",
	"and",
	"unless",
	"or",
}

func binaryOpPrecedence(expr *BinaryExpr) *BinaryExpr {
	for _, op := range opPrecedence {
		expr = unwrapBinaryExpr(exprPrecedence(expr, op))
	}
	return expr
}

func exprPrecedence(anyexpr Expr, op BinaryOp) Expr {
	expr, ok := anyexpr.(*BinaryExpr)
	if !ok {
		return anyexpr
	}

	if expr.Op != op {
		expr.RHS = exprPrecedence(expr.RHS, op)
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
		expr = unwrapBinaryExpr(exprPrecedence(expr, op))
		expr.RHS = exprPrecedence(expr.RHS, op)
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
