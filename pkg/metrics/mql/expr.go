package mql

import (
	"time"

	"github.com/uptrace/pkg/unsafeconv"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
)

type NamedExpr struct {
	Part     *QueryPart
	AST      ast.Expr
	Expr     Expr
	HasAlias bool
	Alias    string
}

func (e *NamedExpr) NameTemplate() string {
	if e.HasAlias {
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
	Metric       string
	RollupWindow time.Duration
	Offset       time.Duration

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
	Name string
}

type ParenExpr struct {
	Expr Expr
}
