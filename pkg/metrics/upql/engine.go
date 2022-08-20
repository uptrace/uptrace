package upql

import (
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
	"golang.org/x/exp/slices"
)

type Engine struct {
	storage Storage
	vars    map[string][]Timeseries
	buf     []byte
}

type Storage interface {
	MakeTimeseries() []Timeseries
	SelectTimeseries(f *TimeseriesFilter) ([]Timeseries, error)
}

func NewEngine(storage Storage) *Engine {
	return &Engine{
		storage: storage,
		vars:    make(map[string][]Timeseries),
	}
}

func (e *Engine) Run(query []*QueryPart) []Timeseries {
	for k := range e.vars {
		delete(e.vars, k)
	}

	var timeseries []Timeseries

	exprs := compile(e.storage, query)
	for _, expr := range exprs {
		if expr.Part.Error.Wrapped != nil {
			continue
		}

		tmp, err := e.eval(expr.Expr)
		if err != nil {
			expr.Part.Error.Wrapped = err
			continue
		}

		if expr.Alias == "" {
			if expr, ok := expr.Expr.(*BinaryExpr); ok {
				setTimeseriesMetric(tmp, expr.AST.String())
			}

			timeseries = append(timeseries, tmp...)
			continue
		}

		for i := range tmp {
			ts := &tmp[i]
			ts.Metric = expr.Alias
			ts.Filters = nil
		}

		if _, ok := e.vars[expr.Alias]; ok {
			expr.Part.Error.Wrapped = fmt.Errorf("alias %q already exists", expr.Alias)
			continue
		}
		e.vars[expr.Alias] = tmp

		if !strings.HasPrefix(expr.Alias, "_") {
			timeseries = append(timeseries, tmp...)
		}
	}

	return timeseries
}

func (e *Engine) eval(expr Expr) ([]Timeseries, error) {
	switch expr := expr.(type) {
	case *TimeseriesExpr:
		return expr.Timeseries, nil
	case *RefExpr:
		if ts, ok := e.vars[expr.Metric]; ok {
			clone := make([]Timeseries, len(ts))
			copy(clone, ts)
			return clone, nil
		}
		return nil, fmt.Errorf("can't find timeseries %q", expr.Metric)
	case *BinaryExpr:
		return e.binaryExpr(expr)
	case ParenExpr:
		return e.eval(expr.Expr)
	case *ast.Number:
		timeseries := e.storage.MakeTimeseries()

		ts := &timeseries[0]
		num := expr.Float64()
		for i := range ts.Value {
			ts.Value[i] = num
		}

		return timeseries, nil
	default:
		return nil, fmt.Errorf("unsupported expr: %T", expr)
	}
}

func (e *Engine) binaryExpr(expr *BinaryExpr) ([]Timeseries, error) {
	{
		lhsNum, lhsOK := expr.LHS.(*ast.Number)
		rhsNum, rhsOK := expr.RHS.(*ast.Number)

		if lhsOK && rhsOK {
			return e.binaryExprNum(lhsNum.Float64(), rhsNum.Float64(), expr.Op)
		}
		if lhsOK {
			rhs, err := e.eval(expr.RHS)
			if err != nil {
				return nil, err
			}
			return e.binaryExprNumLeft(lhsNum.Float64(), rhs, expr.Op)
		}
		if rhsOK {
			lhs, err := e.eval(expr.LHS)
			if err != nil {
				return nil, err
			}
			return e.binaryExprNumRight(lhs, rhsNum.Float64(), expr.Op)
		}
	}

	lhs, err := e.eval(expr.LHS)
	if err != nil {
		return nil, err
	}

	rhs, err := e.eval(expr.RHS)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "+":
		return e.join(lhs, rhs, addOp)
	case "-":
		return e.join(lhs, rhs, subtractOp)
	case "*":
		return e.join(lhs, rhs, multiplyOp)
	case "/":
		return e.join(lhs, rhs, divideOp)
	case "%":
		return e.join(lhs, rhs, remOp)
	case "==":
		return e.join(lhs, rhs, equalOp)
	case "!=":
		return e.join(lhs, rhs, notEqualOp)
	case ">":
		return e.join(lhs, rhs, gtOp)
	case ">=":
		return e.join(lhs, rhs, gteOp)
	case "<":
		return e.join(lhs, rhs, ltOp)
	case "<=":
		return e.join(lhs, rhs, lteOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", expr.Op)
	}
}

func (e *Engine) join(
	lhs, rhs []Timeseries, fn binaryOpFunc,
) ([]Timeseries, error) {
	if len(lhs) == 0 {
		return rhs, nil
	}
	if len(rhs) == 0 {
		return lhs, nil
	}

	if !slices.Equal(lhs[0].Grouping, rhs[0].Grouping) {
		return nil, fmt.Errorf("can't join timeseries with different grouping")
	}
	for i := range lhs {
		if lhs[i].GroupByAll {
			return nil, fmt.Errorf("joining timeseries with `group by all` is forbidden")
		}
	}
	for i := range rhs {
		if rhs[i].GroupByAll {
			return nil, fmt.Errorf("joining timeseries with `group by all` is forbidden")
		}
	}

	joined := make([]Timeseries, 0, max(len(lhs), len(rhs)))
	joined = append(joined, lhs...)

	m := e.makeTimeseriesMap(joined)
	for i := range rhs {
		ts2 := &rhs[i]

		e.buf = ts2.Attrs.Bytes(e.buf[:0])
		hash := xxhash.Sum64(e.buf)

		ts1, ok := m[hash]
		if !ok {
			joined = append(joined, *ts2)
			continue
		}

		value := make([]float64, len(ts1.Value))
		for i, v1 := range ts1.Value {
			v2 := ts2.Value[i]
			value[i] = fn(v1, v2)
		}
		ts1.Value = value
	}

	return joined, nil
}

func (e *Engine) makeTimeseriesMap(timeseries []Timeseries) map[uint64]*Timeseries {
	m := make(map[uint64]*Timeseries, len(timeseries))
	for i := range timeseries {
		ts := &timeseries[i]
		e.buf = ts.Attrs.Bytes(e.buf[:0])
		hash := xxhash.Sum64(e.buf)
		m[hash] = ts
	}
	return m
}

func (e *Engine) binaryExprNum(lhs, rhs float64, op ast.BinaryOp) ([]Timeseries, error) {
	switch op {
	case "+":
		return e.evalBinaryExprNum(lhs, rhs, addOp)
	case "-":
		return e.evalBinaryExprNum(lhs, rhs, subtractOp)
	case "*":
		return e.evalBinaryExprNum(lhs, rhs, multiplyOp)
	case "/":
		return e.evalBinaryExprNum(lhs, rhs, divideOp)
	case "%":
		return e.evalBinaryExprNum(lhs, rhs, remOp)
	case "==":
		return e.evalBinaryExprNum(lhs, rhs, equalOp)
	case "!=":
		return e.evalBinaryExprNum(lhs, rhs, notEqualOp)
	case ">":
		return e.evalBinaryExprNum(lhs, rhs, gtOp)
	case ">=":
		return e.evalBinaryExprNum(lhs, rhs, gteOp)
	case "<":
		return e.evalBinaryExprNum(lhs, rhs, ltOp)
	case "<=":
		return e.evalBinaryExprNum(lhs, rhs, lteOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", op)
	}
}

func (e *Engine) evalBinaryExprNum(lhs, rhs float64, fn binaryOpFunc) ([]Timeseries, error) {
	timeseries := e.storage.MakeTimeseries()

	ts := &timeseries[0]
	result := fn(lhs, rhs)
	for i := range ts.Value {
		ts.Value[i] = result
	}

	return timeseries, nil
}

func (e *Engine) binaryExprNumLeft(
	lhs float64, rhs []Timeseries, op ast.BinaryOp,
) ([]Timeseries, error) {
	switch op {
	case "+":
		return e.evalBinaryExprNumLeft(lhs, rhs, addOp)
	case "-":
		return e.evalBinaryExprNumLeft(lhs, rhs, subtractOp)
	case "*":
		return e.evalBinaryExprNumLeft(lhs, rhs, multiplyOp)
	case "/":
		return e.evalBinaryExprNumLeft(lhs, rhs, divideOp)
	case "%":
		return e.evalBinaryExprNumLeft(lhs, rhs, remOp)
	case "==":
		return e.evalBinaryExprNumLeft(lhs, rhs, equalOp)
	case "!=":
		return e.evalBinaryExprNumLeft(lhs, rhs, notEqualOp)
	case ">":
		return e.evalBinaryExprNumLeft(lhs, rhs, gtOp)
	case ">=":
		return e.evalBinaryExprNumLeft(lhs, rhs, gteOp)
	case "<":
		return e.evalBinaryExprNumLeft(lhs, rhs, ltOp)
	case "<=":
		return e.evalBinaryExprNumLeft(lhs, rhs, lteOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", op)
	}
}

func (e *Engine) evalBinaryExprNumLeft(
	lhs float64, rhs []Timeseries, fn binaryOpFunc,
) ([]Timeseries, error) {
	joined := make([]Timeseries, 0, len(rhs))

	for i := range rhs {
		ts2 := &rhs[i]

		joined = append(joined, Timeseries{
			Metric: "",
			Attrs:  ts2.Attrs,
			Value:  make([]float64, len(ts2.Value)),
			Time:   ts2.Time,
		})
		ts := &joined[len(joined)-1]

		for i, v2 := range ts2.Value {
			ts.Value[i] = fn(lhs, v2)
		}
	}

	return joined, nil
}

func (e *Engine) binaryExprNumRight(
	lhs []Timeseries, rhs float64, op ast.BinaryOp,
) ([]Timeseries, error) {
	switch op {
	case "+":
		return e.evalBinaryExprNumRight(lhs, rhs, addOp)
	case "-":
		return e.evalBinaryExprNumRight(lhs, rhs, subtractOp)
	case "*":
		return e.evalBinaryExprNumRight(lhs, rhs, multiplyOp)
	case "/":
		return e.evalBinaryExprNumRight(lhs, rhs, divideOp)
	case "%":
		return e.evalBinaryExprNumRight(lhs, rhs, remOp)
	case "==":
		return e.evalBinaryExprNumRight(lhs, rhs, equalOp)
	case "!=":
		return e.evalBinaryExprNumRight(lhs, rhs, notEqualOp)
	case ">":
		return e.evalBinaryExprNumRight(lhs, rhs, gtOp)
	case ">=":
		return e.evalBinaryExprNumRight(lhs, rhs, gteOp)
	case "<":
		return e.evalBinaryExprNumRight(lhs, rhs, ltOp)
	case "<=":
		return e.evalBinaryExprNumRight(lhs, rhs, lteOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", op)
	}
}

func (e *Engine) evalBinaryExprNumRight(
	lhs []Timeseries, rhs float64, fn binaryOpFunc,
) ([]Timeseries, error) {
	joined := make([]Timeseries, 0, len(lhs))

	for i := range lhs {
		ts1 := &lhs[i]

		joined = append(joined, Timeseries{
			Metric: "",
			Attrs:  ts1.Attrs,
			Value:  make([]float64, len(ts1.Value)),
			Time:   ts1.Time,
		})
		ts := &joined[len(joined)-1]

		for i, v1 := range ts1.Value {
			ts.Value[i] = fn(v1, rhs)
		}
	}

	return joined, nil
}
