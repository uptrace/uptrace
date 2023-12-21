package mql

import (
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type NamedExpr struct {
	Part  *QueryPart
	AST   ast.Expr
	Expr  Expr
	Alias string
}

func (e *NamedExpr) String() string {
	if e.Alias != "" {
		return e.Alias
	}
	return unsafeconv.String(e.AST.AppendString(nil))
}

func (e *NamedExpr) NameTemplate() string {
	if e.Alias != "" {
		return e.Alias + "$$"
	}
	return unsafeconv.String(e.AST.AppendTemplate(nil))
}

type Expr interface{}

var (
	_ Expr = (*TimeseriesExpr)(nil)
	_ Expr = (*FuncCall)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*ParenExpr)(nil)
	_ Expr = (*RefExpr)(nil)
)

type TimeseriesExpr struct {
	Metric string

	CHFunc string
	Attr   string

	Uniq []string

	Filters  []ast.Filter
	Where    [][]ast.Filter
	Grouping []ast.GroupingElem

	Part       *QueryPart
	Timeseries []*Timeseries
}

type FuncCall struct {
	Func     string
	Arg      Expr
	Grouping []string
}

type BinaryExpr struct {
	Op  ast.BinaryOp
	LHS Expr
	RHS Expr
}

type RefExpr struct {
	Expr *ast.MetricExpr
}

type ParenExpr struct {
	Expr Expr
}
