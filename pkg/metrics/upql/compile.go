package upql

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
)

//------------------------------------------------------------------------------

type Expr interface{}

type NamedExpr struct {
	Part  *QueryPart
	Expr  Expr // *TimeseriesExpr | *ExprChain
	Alias string
}

type TimeseriesExpr struct {
	Metric     string
	Func       string
	Filters    []ast.Filter
	Where      [][]ast.Filter
	Grouping   []string
	GroupByAll bool

	Part       *QueryPart
	Timeseries []Timeseries
}

type RefExpr struct {
	Metric string
}

type BinaryExpr struct {
	AST *ast.BinaryExpr

	Op  ast.BinaryOp
	LHS Expr
	RHS Expr
}

type ParenExpr struct {
	Expr Expr
}

type compiler struct {
	storage    Storage
	exprs      []NamedExpr
	timeseries []*TimeseriesExpr
}

func compile(storage Storage, parts []*QueryPart) []NamedExpr {
	c := &compiler{
		storage: storage,
	}

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch ast := part.AST.(type) {
		case *ast.Selector:
			pos := len(c.timeseries)
			expr := c.selector(ast.Expr.Expr)

			for _, ts := range c.timeseries[pos:] {
				if ast.GroupByAll {
					ts.GroupByAll = true
				} else {
					ts.Grouping = append(ts.Grouping, ast.Grouping...)
				}
				ts.Part = part
			}

			c.exprs = append(c.exprs, NamedExpr{
				Part:  part,
				Expr:  expr,
				Alias: ast.Expr.Alias,
			})
		case *ast.Where, *ast.Grouping:
			// see below
		default:
			panic(fmt.Errorf("unknown ast: %T", ast))
		}
	}

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch ast := part.AST.(type) {
		case *ast.Where:
			if err := c.where(ast); err != nil {
				part.Error.Wrapped = err
			}
		case *ast.Grouping:
			if err := c.grouping(ast); err != nil {
				part.Error.Wrapped = err
			}
		}
	}

	var wg sync.WaitGroup

	for _, expr := range c.timeseries {
		expr := expr

		wg.Add(1)
		go func() {
			defer wg.Done()

			timeseries, err := storage.SelectTimeseries(&TimeseriesFilter{
				Metric:     expr.Metric,
				Func:       expr.Func,
				Filters:    expr.Filters,
				Where:      expr.Where,
				Grouping:   expr.Grouping,
				GroupByAll: expr.GroupByAll,
			})
			if err != nil {
				if _, ok := err.(*ch.Error); ok {
					expr.Part.Error.Wrapped = errors.New("internal error")
				} else {
					expr.Part.Error.Wrapped = err
				}
			} else {
				expr.Timeseries = timeseries
			}
		}()
	}

	wg.Wait()

	return c.exprs
}

func (c *compiler) selector(expr ast.Expr) Expr {
	switch expr := expr.(type) {
	case *ast.Name:
		if strings.HasPrefix(expr.Name, "$") {
			ts := &TimeseriesExpr{
				Metric: strings.TrimPrefix(expr.Name, "$"),
				Func:   expr.Func,
			}
			c.timeseries = append(c.timeseries, ts)
			return ts
		}
		return &RefExpr{
			Metric: expr.Name,
		}
	case *ast.FilteredName:
		ts := &TimeseriesExpr{
			Metric:  strings.TrimPrefix(expr.Name.Name, "$"),
			Func:    expr.Name.Func,
			Filters: expr.Filters,
		}
		c.timeseries = append(c.timeseries, ts)
		return ts
	case *ast.BinaryExpr:
		return &BinaryExpr{
			AST: expr,
			Op:  expr.Op,
			LHS: c.selector(expr.LHS),
			RHS: c.selector(expr.RHS),
		}
	case ast.ParenExpr:
		return ParenExpr{
			Expr: c.selector(expr.Expr),
		}
	case *ast.Number:
		return expr
	default:
		panic(fmt.Errorf("unknown selector expr: %T", expr))
	}
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
