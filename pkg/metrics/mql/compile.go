package mql

import (
	"errors"
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

		switch value := part.AST.(type) {
		case *ast.Selector:
			pos := len(c.timeseries)

			sel, err := c.selector(value.Expr.Expr)
			if err != nil {
				part.Error.Wrapped = err
				break
			}

			for _, ts := range c.timeseries[pos:] {
				ts.Part = part
			}

			c.exprs = append(c.exprs, NamedExpr{
				Part:     part,
				AST:      value.Expr.Expr,
				Expr:     sel,
				HasAlias: value.Expr.HasAlias,
				Alias:    value.Expr.Alias,
			})
		case *ast.Where, *ast.Grouping:
			// see below
		default:
			panic(fmt.Errorf("unknown ast: %T", value))
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
			if err := c.applyGlobalGrouping(expr); err != nil {
				part.Error.Wrapped = err
			}
		}
	}

	return c.exprs, c.timeseries
}

type compiler struct {
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
	case *ast.MetricExpr:
		if !strings.HasPrefix(expr.Name, "$") {
			return RefExpr{
				Expr: expr,
			}
		}
		return c.metricExpr(expr)

	case *ast.BinaryExpr:
		return &BinaryExpr{
			Op:  expr.Op,
			LHS: c.panickySelector(expr.LHS),
			RHS: c.panickySelector(expr.RHS),
		}

	case ast.ParenExpr:
		return ParenExpr{
			Expr: c.panickySelector(expr.Expr),
		}

	case ast.Number:
		return expr

	case *ast.FuncCall:
		return c.funcCall(expr)

	case *ast.UniqExpr:
		return c.uniqExpr(expr)

	default:
		panic(fmt.Errorf("unknown selector expr: %T", expr))
	}
}

func (c *compiler) metricExpr(expr *ast.MetricExpr) *TimeseriesExpr {
	metric, attr := ast.SplitAliasName(expr.Name)
	ts := &TimeseriesExpr{
		Metric:   metric,
		Attr:     attr,
		Filters:  expr.Filters,
		Grouping: expr.Grouping,
	}
	c.timeseries = append(c.timeseries, ts)
	return ts
}

func (c *compiler) funcCall(fn *ast.FuncCall) Expr {
	switch arg := c.panickySelector(fn.Arg); arg := arg.(type) {
	case *TimeseriesExpr:
		if arg.CHFunc == "" {
			if isCHFunc(fn.Func) {
				arg.CHFunc = fn.Func
				return arg
			}
			arg.CHFunc = "_"
		}

		return &FuncCall{
			Func:     fn.Func,
			Arg:      arg,
			Grouping: fn.Grouping.Attrs(),
		}
	default:
		return &FuncCall{
			Func:     fn.Func,
			Arg:      arg,
			Grouping: fn.Grouping.Attrs(),
		}
	}
}

func (c *compiler) uniqExpr(uq *ast.UniqExpr) *TimeseriesExpr {
	expr := c.metricExpr(&uq.Name)
	expr.CHFunc = CHAggUniq
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

func (c *compiler) applyGlobalGrouping(expr *ast.Grouping) error {
	for _, elem := range expr.Elems {
		if alias, _ := ast.SplitAliasName(elem.Name); alias != "" {
			return errors.New("global grouping can't reference a metric")
		}
	}
	for _, namedExpr := range c.exprs {
		applyGrouping(namedExpr.Expr, expr.Elems)
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

func applyGrouping(expr Expr, grouping []ast.GroupingElem) {
	switch expr := expr.(type) {
	case *TimeseriesExpr:
		expr.Grouping = append(expr.Grouping, grouping...)
	case *FuncCall:
		for _, elem := range grouping {
			expr.Grouping = append(expr.Grouping, elem.Alias)
		}
		applyGrouping(expr.Arg, grouping)
	case *BinaryExpr:
		applyGrouping(expr.LHS, grouping)
		applyGrouping(expr.RHS, grouping)
	case ParenExpr:
		applyGrouping(expr.Expr, grouping)
	case RefExpr, ast.Number:
		// nothing
	default:
		panic(fmt.Errorf("unsupported expr: %T", expr))
	}
}
