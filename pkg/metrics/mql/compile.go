package mql

import (
	"fmt"
	"strings"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
)

func compile(parts []*QueryPart) ([]NamedExpr, []*TimeseriesExpr) {
	c := new(compiler)

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch expr := part.AST.(type) {
		case *ast.Selector:
			pos := len(c.timeseries)

			sel, err := c.selector(expr.Expr.Expr)
			if err != nil {
				part.Error.Wrapped = err
				break
			}

			for _, ts := range c.timeseries[pos:] {
				if expr.GroupByAll {
					ts.GroupByAll = true
				} else {
					ts.Grouping = append(ts.Grouping, expr.Grouping...)
				}
				ts.Part = part
			}

			c.exprs = append(c.exprs, NamedExpr{
				Part:  part,
				Expr:  sel,
				Alias: expr.Expr.Alias,
			})
		case *ast.Where, *ast.Grouping:
			// see below
		default:
			panic(fmt.Errorf("unknown ast: %T", expr))
		}
	}

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch expr := part.AST.(type) {
		case *ast.Where:
			if !hasWherePrefix(part.Query) {
				part.Query = "where " + part.Query
			}

			if err := c.where(expr); err != nil {
				part.Error.Wrapped = err
			}
		case *ast.Grouping:
			if err := c.grouping(expr); err != nil {
				part.Error.Wrapped = err
			}
		}
	}

	return c.exprs, c.timeseries
}

type compiler struct {
	storage    Storage
	exprs      []NamedExpr
	timeseries []*TimeseriesExpr
}

func (c *compiler) selector(astExpr ast.Expr) (_ Expr, retErr error) {
	defer func() {
		if v := recover(); v != nil {
			var ok bool
			retErr, ok = v.(error)
			if !ok {
				panic(v)
			}
		}
	}()
	expr := c.panickySelector(astExpr)
	return expr, retErr
}

func (c *compiler) panickySelector(expr ast.Expr) Expr {
	switch expr := expr.(type) {
	case *ast.Name:
		if !strings.HasPrefix(expr.Name, "$") {
			return &RefExpr{
				Name: expr,
			}
		}
		return c.name(expr)

	case *ast.BinaryExpr:
		return &BinaryExpr{
			BinaryExpr: expr,
			Op:         expr.Op,
			LHS:        c.panickySelector(expr.LHS),
			RHS:        c.panickySelector(expr.RHS),
		}

	case ast.ParenExpr:
		return ParenExpr{
			ParenExpr: expr,
			Expr:      c.panickySelector(expr.Expr),
		}

	case *ast.Number:
		return expr

	case *ast.FuncCall:
		return c.funcCall(expr)

	case *ast.UniqExpr:
		return c.uniqExpr(expr)

	default:
		panic(fmt.Errorf("unknown selector expr: %T", expr))
	}
}

func (c *compiler) name(name *ast.Name) *TimeseriesExpr {
	ts := &TimeseriesExpr{
		Name:    name,
		Metric:  name.Name,
		Filters: name.Filters,
	}
	c.timeseries = append(c.timeseries, ts)
	return ts
}

func (c *compiler) funcCall(fn *ast.FuncCall) Expr {
	if isAggFunc(fn.Func) {
		if len(fn.Args) != 1 {
			panic(fmt.Errorf("%q requires a single arg", fn.Func))
		}

		expr, ok := c.panickySelector(fn.Args[0]).(*TimeseriesExpr)
		if !ok {
			panic(fmt.Errorf("%q can be only applied to a timeseries", fn.Func))
		}

		if expr.AggFunc == "" {
			expr.AggFunc = fn.Func
			return expr
		}
	}

	if isTableFunc(fn.Func) {
		if len(fn.Args) != 1 {
			panic(fmt.Errorf("%q requires a single arg", fn.Func))
		}

		expr, ok := c.panickySelector(fn.Args[0]).(*TimeseriesExpr)
		if !ok {
			panic(fmt.Errorf("%q can be only applied to a timeseries", fn.Func))
		}

		if expr.TableFunc != "" {
			panic(fmt.Errorf("can't apply %q to %q", fn.Func, expr))
		}

		expr.TableFunc = fn.Func
		return expr
	}

	if !isOpFunc(fn.Func) {
		if isAggFunc(fn.Func) {
			panic(fmt.Errorf("can't apply %q in this context", fn.Func))
		}
		panic(fmt.Errorf("unsupported func: %q", fn.Func))
	}

	args := make([]Expr, len(fn.Args))
	for i, arg := range fn.Args {
		sel := c.panickySelector(arg)
		args[i] = sel

		if expr, ok := sel.(*TimeseriesExpr); ok && expr.TableFunc == "" {
			expr.TableFunc = fn.Func
		}
	}

	return &FuncCall{
		FuncCall: fn,
		Func:     fn.Func,
		Args:     args,
	}
}

func (c *compiler) uniqExpr(uq *ast.UniqExpr) *TimeseriesExpr {
	expr := c.name(&uq.Name)
	expr.AggFunc = AggUniq
	expr.Uniq = uq.Attrs
	return expr
}

func (c *compiler) where(expr *ast.Where) error {
	var alias string

	for i := range expr.Filters {
		filter := &expr.Filters[i]

		var filterAlias string
		filterAlias, filter.LHS = ast.SplitAliasName(filter.LHS)

		if i == 0 {
			filterAlias = alias
			continue
		}

		if filterAlias != alias {
			return fmt.Errorf("where must reference a single metric: %q != %q", filterAlias, alias)
		}
	}

	if alias == "" {
		for _, ts := range c.timeseries {
			ts.Where = append(ts.Where, expr.Filters)
		}
		return nil
	}

	var found bool

	for _, ts := range c.timeseries {
		if ts.Metric == alias {
			ts.Where = append(ts.Where, expr.Filters)
			found = true
		}
	}

	if !found {
		return fmt.Errorf("can't find metric with alias %q", alias)
	}
	return nil
}

func (c *compiler) grouping(expr *ast.Grouping) error {
	if expr.GroupByAll {
		for _, ts := range c.timeseries {
			ts.GroupByAll = true
		}
		return nil
	}

	for _, name := range expr.Names {
		if err := c.groupingName(name); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) groupingName(name string) error {
	var found bool
	alias, name := ast.SplitAliasName(name)

	for _, ts := range c.timeseries {
		if alias != "" && ts.Metric != alias {
			continue
		}
		ts.Grouping = append(ts.Grouping, name)
		found = true
	}

	if alias != "" && !found {
		return fmt.Errorf("can't find metric with alias %q", alias)
	}
	return nil
}

func hasWherePrefix(s string) bool {
	l := len("where ")

	if len(s) < l {
		return false
	}
	return strings.EqualFold(s[:l], "where ")
}
