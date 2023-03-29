package upql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
)

//------------------------------------------------------------------------------

type Expr interface {
	String() string
}

type NamedExpr struct {
	Part  *QueryPart
	Expr  Expr
	Alias string
}

type TimeseriesExpr struct {
	AST *ast.Name

	Metric     string
	Func       string
	Attr       string
	Filters    []ast.Filter
	Where      [][]ast.Filter
	Grouping   []string
	GroupByAll bool

	Part       *QueryPart
	Timeseries []Timeseries
}

func (e *TimeseriesExpr) String() string {
	ts := Timeseries{
		Metric:  e.Metric,
		Func:    e.Func,
		Filters: e.Filters,
	}
	return ts.Name()
}

type RefExpr struct {
	Metric string
}

func (e *RefExpr) String() string {
	return e.Metric
}

type BinaryExpr struct {
	AST *ast.BinaryExpr

	Op  ast.BinaryOp
	LHS Expr
	RHS Expr
}

func (e *BinaryExpr) String() string {
	return e.AST.String()
}

type ParenExpr struct {
	Expr
}

type FuncCall struct {
	AST *ast.FuncCall

	Func string
	Args []Expr
}

func (fn *FuncCall) String() string {
	return fn.AST.String()
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

		switch expr := part.AST.(type) {
		case *ast.Selector:
			pos := len(c.timeseries)
			sel := c.selector(expr.Expr.Expr)

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

	var wg sync.WaitGroup

	for _, expr := range c.timeseries {
		expr := expr

		wg.Add(1)
		go func() {
			defer wg.Done()

			f := &TimeseriesFilter{
				Metric:     expr.Metric,
				Func:       expr.Func,
				Attr:       expr.Attr,
				Filters:    expr.Filters,
				Where:      expr.Where,
				Grouping:   expr.Grouping,
				GroupByAll: expr.GroupByAll,
			}

			timeseries, err := storage.SelectTimeseries(f)
			if err != nil {
				expr.Part.Error.Wrapped = err
				return
			}

			expr.Timeseries = timeseries
			return
		}()
	}

	wg.Wait()

	return c.exprs
}

func (c *compiler) selector(expr ast.Expr) Expr {
	switch expr := expr.(type) {
	case *ast.Name:
		if !strings.HasPrefix(expr.Name, "$") {
			return &RefExpr{
				Metric: expr.Name,
			}
		}

		metric := strings.TrimPrefix(expr.Name, "$")
		var attr string
		if i := strings.IndexByte(metric, '.'); i >= 0 {
			attr = metric[i+1:]
			metric = metric[:i]
		}

		ts := &TimeseriesExpr{
			AST:     expr,
			Metric:  metric,
			Func:    expr.Func,
			Attr:    attr,
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

	case *ast.FuncCall:
		return c.funcCall(expr)

	default:
		panic(fmt.Errorf("unknown selector expr: %T", expr))
	}
}

func (c *compiler) funcCall(fn *ast.FuncCall) Expr {
	switch fn.Func {
	case "delta":
		// continue below
	default:
		if len(fn.Args) == 1 {
			switch arg := fn.Args[0].(type) {
			case *ast.Name:
				if arg.Func == "" {
					arg.Func = fn.Func
					return c.selector(arg)
				}
			}
		}
	}

	args := make([]Expr, len(fn.Args))
	for i, arg := range fn.Args {
		args[i] = c.selector(arg)
	}
	return &FuncCall{
		AST:  fn,
		Func: fn.Func,
		Args: args,
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

func hasWherePrefix(s string) bool {
	l := len("where ")

	if len(s) < l {
		return false
	}
	return strings.EqualFold(s[:l], "where ")
}
