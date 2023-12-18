package mql

import (
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type NamedExpr struct {
	Part  *QueryPart
	Expr  Expr
	Alias string
}

func (e *NamedExpr) String() string {
	if e.Alias != "" {
		return e.Alias
	}
	return unsafeconv.String(e.Expr.AppendString(nil))
}

func (e *NamedExpr) NameTemplate() string {
	if e.Alias != "" {
		return e.Alias + "$$"
	}
	return unsafeconv.String(e.Expr.AppendTemplate(nil))
}

type Expr interface {
	ast.Expr
}

var (
	_ Expr = (*TimeseriesExpr)(nil)
	_ Expr = (*FuncCall)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*ParenExpr)(nil)
	_ Expr = (*RefExpr)(nil)
)

type TimeseriesExpr struct {
	ast.Expr

	Metric string

	AggFunc string
	Attr    string

	TableFunc string
	Uniq      []string

	Filters []ast.Filter
	Where   [][]ast.Filter

	Grouping []ast.NamedExpr

	Part       *QueryPart
	Timeseries []Timeseries
}

func (e *TimeseriesExpr) String() string {
	var b []byte

	if e.AggFunc != "" {
		b = append(b, e.AggFunc...)
		b = append(b, '(')
	}

	b = append(b, e.Metric...)

	if e.AggFunc != "" {
		b = append(b, ')')
	}

	return unsafeconv.String(b)
}

type FuncCall struct {
	*ast.FuncCall

	Func string
	Args []Expr
}

type BinaryExpr struct {
	*ast.BinaryExpr

	Op  ast.BinaryOp
	LHS Expr
	RHS Expr
}

type RefExpr struct {
	*ast.Name
}

type ParenExpr struct {
	ast.ParenExpr

	Expr Expr
}
