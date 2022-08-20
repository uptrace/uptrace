package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type Selector struct {
	Expr       NamedExpr
	Grouping   []string
	GroupByAll bool
}

type NamedExpr struct {
	Expr  Expr // *Name | *FilteredName | *BinaryExpr | *FuncCall
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
	Func string
	Name string
}

func (n Name) String() string {
	if n.Func != "" {
		return n.Func + "(" + n.Name + ")"
	}
	return n.Name
}

type Number struct {
	Text string
}

func (n *Number) String() string {
	return n.Text
}

func (n *Number) Float32() float32 {
	return float32(n.Float64())
}

func (n *Number) Float64() float64 {
	f, err := strconv.ParseFloat(n.Text, 64)
	if err != nil {
		panic(err)
	}
	return f
}

type FilteredName struct {
	Name    Name
	Filters []Filter
}

func (n *FilteredName) String() string {
	var b []byte
	for i := range n.Filters {
		if i > 0 {
			b = append(b, ',')
		}
		b = n.Filters[i].AppendString(b)
	}

	return n.Name.String() + "{" + string(b) + "}"
}

type FuncCall struct {
	Func string
	Args []Expr
}

type BinaryExpr struct {
	Op       BinaryOp
	LHS, RHS Expr
	JoinOn   []string
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
	RHS string
}

func (f *Filter) String() string {
	b := make([]byte, len(f.LHS)+len(f.Op)+len(f.RHS))
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

	if isIdent(f.RHS) {
		b = append(b, f.RHS...)
	} else {
		b = strconv.AppendQuote(b, f.RHS)
	}

	return b
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
