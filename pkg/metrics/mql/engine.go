package mql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"golang.org/x/exp/slices"
)

type Engine struct {
	storage Storage

	consts map[string]float64
	vars   map[string][]Timeseries
	buf    []byte
}

type Storage interface {
	Consts() map[string]float64
	MakeTimeseries(f *TimeseriesFilter) []Timeseries
	SelectTimeseries(f *TimeseriesFilter) ([]Timeseries, error)
}

func NewEngine(storage Storage) *Engine {
	return &Engine{
		storage: storage,
		consts:  storage.Consts(),
		vars:    make(map[string][]Timeseries),
	}
}

type Result struct {
	Columns    []string
	Timeseries []Timeseries
}

func (e *Engine) Run(parts []*QueryPart) *Result {
	exprs, timeseriesExprs := compile(parts)

	for _, expr := range timeseriesExprs {
		f := &TimeseriesFilter{
			Metric:     expr.Metric,
			AggFunc:    expr.AggFunc,
			TableFunc:  expr.TableFunc,
			Uniq:       expr.Uniq,
			Filters:    expr.Filters,
			Where:      expr.Where,
			Grouping:   expr.Grouping,
			GroupByAll: expr.GroupByAll,
		}

		timeseries, err := e.storage.SelectTimeseries(f)
		if err != nil {
			expr.Part.Error.Wrapped = err
			continue
		}

		expr.Timeseries = timeseries
	}

	result := new(Result)

	for _, expr := range exprs {
		if expr.Part.Error.Wrapped != nil {
			continue
		}

		tmp, err := e.eval(expr.Expr)
		if err != nil {
			expr.Part.Error.Wrapped = err
			continue
		}

		metricName := expr.String()
		updateTimeseries(tmp, metricName, expr.NameTemplate(), expr.Alias != "")

		if _, ok := e.vars[metricName]; ok {
			expr.Part.Error.Wrapped = fmt.Errorf("column %q already exists", metricName)
			continue
		}
		e.vars[metricName] = tmp

		if !strings.HasPrefix(metricName, "_") {
			result.Columns = append(result.Columns, metricName)
			result.Timeseries = append(result.Timeseries, tmp...)
		}
	}

	return result
}

func updateTimeseries(timeseries []Timeseries, metricName, nameTemplate string, isAlias bool) {
	for i := range timeseries {
		ts := &timeseries[i]
		ts.MetricName = metricName
		ts.NameTemplate = nameTemplate
		if isAlias {
			ts.Filters = nil
		}
	}
}

func (e *Engine) eval(expr Expr) ([]Timeseries, error) {
	switch expr := expr.(type) {
	case *TimeseriesExpr:
		return expr.Timeseries, nil
	case *RefExpr:
		if ts, ok := e.vars[expr.Name.Name]; ok {
			clone := make([]Timeseries, len(ts))
			copy(clone, ts)
			return clone, nil
		}

		if num, ok := e.consts[expr.Name.Name]; ok {
			timeseries := e.storage.MakeTimeseries(nil)
			ts := &timeseries[0]
			for i := range ts.Value {
				ts.Value[i] = num
			}
			return timeseries, nil
		}

		return nil, fmt.Errorf("can't resolve name %q", expr.Name)
	case *BinaryExpr:
		return e.binaryExpr(expr)
	case ParenExpr:
		return e.eval(expr.Expr)
	case *ast.Number:
		timeseries := e.storage.MakeTimeseries(nil)
		ts := &timeseries[0]

		num, err := expr.ConvertValue(ts.Unit)
		if err != nil {
			return nil, err
		}

		for i := range ts.Value {
			ts.Value[i] = num
		}

		return timeseries, nil
	case *FuncCall:
		return e.callFunc(expr)
	default:
		return nil, fmt.Errorf("unsupported expr: %T", expr)
	}
}

func (e *Engine) resolveNumber(expr Expr) Expr {
	if ref, ok := expr.(*RefExpr); ok {
		if num, ok := e.consts[ref.Name.Name]; ok {
			return &ast.Number{
				Text: strconv.FormatFloat(num, 'f', -1, 64),
			}
		}
	}
	return expr
}

func (e *Engine) binaryExpr(expr *BinaryExpr) ([]Timeseries, error) {
	lhs := e.resolveNumber(expr.LHS)
	rhs := e.resolveNumber(expr.RHS)

	{
		lhsNum, lhsOK := lhs.(*ast.Number)
		rhsNum, rhsOK := rhs.(*ast.Number)

		if lhsOK && rhsOK {
			return e.binaryExprNum(lhsNum.Float64(), rhsNum.Float64(), expr.Op)
		}
		if lhsOK {
			rhs, err := e.eval(rhs)
			if err != nil {
				return nil, err
			}
			if len(rhs) == 0 {
				return nil, nil
			}

			lhs, err := lhsNum.ConvertValue(rhs[0].Unit)
			if err != nil {
				return nil, err
			}

			return e.binaryExprNumLeft(lhs, rhs, expr.Op)
		}
		if rhsOK {
			lhs, err := e.eval(lhs)
			if err != nil {
				return nil, err
			}
			if len(lhs) == 0 {
				return nil, nil
			}

			rhs, err := rhsNum.ConvertValue(lhs[0].Unit)
			if err != nil {
				return nil, err
			}

			return e.binaryExprNumRight(lhs, rhs, expr.Op)
		}
	}

	lhsTimeseries, err := e.eval(lhs)
	if err != nil {
		return nil, err
	}

	rhsTimeseries, err := e.eval(rhs)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "+":
		return e.join(lhsTimeseries, rhsTimeseries, addOp)
	case "-":
		return e.join(lhsTimeseries, rhsTimeseries, subtractOp)
	case "*":
		return e.join(lhsTimeseries, rhsTimeseries, multiplyOp)
	case "/":
		return e.join(lhsTimeseries, rhsTimeseries, divideOp)
	case "%":
		return e.join(lhsTimeseries, rhsTimeseries, remOp)
	case "==":
		return e.join(lhsTimeseries, rhsTimeseries, equalOp)
	case "!=":
		return e.join(lhsTimeseries, rhsTimeseries, notEqualOp)
	case ">":
		return e.join(lhsTimeseries, rhsTimeseries, gtOp)
	case ">=":
		return e.join(lhsTimeseries, rhsTimeseries, gteOp)
	case "<":
		return e.join(lhsTimeseries, rhsTimeseries, ltOp)
	case "<=":
		return e.join(lhsTimeseries, rhsTimeseries, lteOp)
	case "and":
		return e.join(lhsTimeseries, rhsTimeseries, andOp)
	case "or":
		return e.join(lhsTimeseries, rhsTimeseries, orOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", expr.Op)
	}
}

func (e *Engine) join(
	lhs, rhs []Timeseries, op binaryOpFunc,
) ([]Timeseries, error) {
	if len(lhs) == 0 && len(rhs) == 0 {
		return nil, nil
	}
	if len(lhs) == 0 {
		return e.evalBinaryExprNumLeft(0, rhs, op)
	}
	if len(rhs) == 0 {
		return e.evalBinaryExprNumRight(lhs, 0, op)
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
			joined = append(joined, newTimeseriesFrom(ts2))
			ts1 = &joined[len(joined)-1]
		}

		joinedValue := make([]float64, len(ts1.Value))
		for i, v1 := range ts1.Value {
			v2 := ts2.Value[i]
			joinedValue[i] = op(v1, v2)
		}
		ts1.Value = joinedValue
		ts1.Unit = bununit.None
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
	case "and":
		return e.evalBinaryExprNum(lhs, rhs, andOp)
	case "or":
		return e.evalBinaryExprNum(lhs, rhs, orOp)
	default:
		return nil, fmt.Errorf("unsupported binary op: %q", op)
	}
}

func (e *Engine) evalBinaryExprNum(lhs, rhs float64, fn binaryOpFunc) ([]Timeseries, error) {
	timeseries := e.storage.MakeTimeseries(new(TimeseriesFilter))

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
	case "and":
		return e.evalBinaryExprNumLeft(lhs, rhs, andOp)
	case "or":
		return e.evalBinaryExprNumLeft(lhs, rhs, orOp)
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

		joined = append(joined, newTimeseriesFrom(ts2))
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
	case "and":
		return e.evalBinaryExprNumRight(lhs, rhs, andOp)
	case "or":
		return e.evalBinaryExprNumRight(lhs, rhs, orOp)
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

		joined = append(joined, newTimeseriesFrom(ts1))
		ts := &joined[len(joined)-1]

		for i, v1 := range ts1.Value {
			ts.Value[i] = fn(v1, rhs)
		}
	}

	return joined, nil
}

func (e *Engine) callFunc(fn *FuncCall) ([]Timeseries, error) {
	switch fn.Func {
	case FuncDelta:
		return e.callSingleArgFunc(fn.Func, fn, deltaFunc)
	case FuncPerMin:
		return e.callSingleArgFunc(fn.Func, fn, perMinFunc)
	case FuncPerSec:
		return e.callSingleArgFunc(fn.Func, fn, perSecFunc)
	default:
		return nil, fmt.Errorf("unsupported func: %s", fn.Func)
	}
}

func (e *Engine) callSingleArgFunc(
	funcName string,
	fn *FuncCall,
	op FuncOp,
) ([]Timeseries, error) {
	if len(fn.Args) != 1 {
		return nil, fmt.Errorf("%s func expects a single arg", funcName)
	}

	timeseries, err := e.eval(fn.Args[0])
	if err != nil {
		return nil, err
	}

	for i := range timeseries {
		op(timeseries[i].Value, e.consts)
	}

	return timeseries, nil
}
