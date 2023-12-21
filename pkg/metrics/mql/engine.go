package mql

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/zeebo/xxh3"
	"go4.org/syncutil"
)

type Engine struct {
	storage Storage

	consts map[string]float64
	vars   map[string][]*Timeseries
	buf    []byte
}

type Storage interface {
	Consts() map[string]float64
	MakeTimeseries(f *TimeseriesFilter) *Timeseries
	SelectTimeseries(f *TimeseriesFilter) ([]*Timeseries, error)
}

func NewEngine(storage Storage) *Engine {
	return &Engine{
		storage: storage,
		consts:  storage.Consts(),
		vars:    make(map[string][]*Timeseries),
	}
}

type Result struct {
	Metrics    []MetricInfo
	Timeseries []*Timeseries
}

type MetricInfo struct {
	Name      string
	TableFunc string
}

func (e *Engine) Run(parts []*QueryPart) *Result {
	exprs, timeseriesExprs := compile(parts)
	var wg sync.WaitGroup

	gate := syncutil.NewGate(2)
	for _, expr := range timeseriesExprs {
		expr := expr

		wg.Add(1)
		go func() {
			defer wg.Done()

			f := &TimeseriesFilter{
				Metric: expr.Metric,

				CHFunc: expr.CHFunc,
				Attr:   expr.Attr,

				Uniq:     expr.Uniq,
				Filters:  expr.Filters,
				Where:    expr.Where,
				Grouping: expr.Grouping,
			}

			gate.Start()
			defer gate.Done()

			timeseries, err := e.storage.SelectTimeseries(f)
			if err != nil {
				expr.Part.Error.Wrapped = err
			} else {
				expr.Timeseries = timeseries
			}
		}()
	}

	wg.Wait()
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
			result.Metrics = append(result.Metrics, MetricInfo{
				Name:      metricName,
				TableFunc: TableFuncName(expr.AST),
			})
			result.Timeseries = append(result.Timeseries, tmp...)
		}
	}

	return result
}

func updateTimeseries(timeseries []*Timeseries, metricName, nameTemplate string, isAlias bool) {
	for _, ts := range timeseries {
		ts.MetricName = metricName
		ts.NameTemplate = nameTemplate
		if isAlias {
			ts.Filters = nil
		}
	}
}

func (e *Engine) eval(expr Expr) ([]*Timeseries, error) {
	switch expr := expr.(type) {
	case *TimeseriesExpr:
		return expr.Timeseries, nil
	case *RefExpr:
		if timeseries, ok := e.vars[expr.Expr.Name]; ok {
			return timeseries, nil
		}

		if num, ok := e.consts[expr.Expr.Name]; ok {
			ts := e.storage.MakeTimeseries(nil)
			for i := range ts.Value {
				ts.Value[i] = num
			}
			return []*Timeseries{ts}, nil
		}

		return nil, fmt.Errorf("can't resolve name %q", ast.String(expr.Expr))
	case *BinaryExpr:
		return e.binaryExpr(expr)
	case ParenExpr:
		return e.eval(expr.Expr)
	case *ast.Number:
		ts := e.storage.MakeTimeseries(nil)

		num, err := expr.ConvertValue(ts.Unit)
		if err != nil {
			return nil, err
		}

		for i := range ts.Value {
			ts.Value[i] = num
		}

		return []*Timeseries{ts}, nil
	case *FuncCall:
		return e.callFunc(expr)
	default:
		return nil, fmt.Errorf("unsupported expr: %T", expr)
	}
}

func (e *Engine) resolveNumber(expr Expr) Expr {
	if ref, ok := expr.(*RefExpr); ok {
		if num, ok := e.consts[ref.Expr.Name]; ok {
			return ast.Number{
				Text: strconv.FormatFloat(num, 'f', -1, 64),
			}
		}
	}
	return expr
}

func (e *Engine) binaryExpr(expr *BinaryExpr) ([]*Timeseries, error) {
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
	lhs, rhs []*Timeseries, op binaryOpFunc,
) ([]*Timeseries, error) {
	if len(lhs) == 0 && len(rhs) == 0 {
		return nil, nil
	}
	if len(lhs) == 0 {
		return e.evalBinaryExprNumLeft(0, rhs, op)
	}
	if len(rhs) == 0 {
		return e.evalBinaryExprNumRight(lhs, 0, op)
	}

	lhsGrouping := lhs[0].Grouping
	rhsGrouping := rhs[0].Grouping

	if slices.Equal(lhsGrouping, rhsGrouping) {
		return e.fullJoin(lhs, rhs, op, nil)
	}
	if len(lhsGrouping) > len(rhsGrouping) {
		return e.oneToManyJoin(rhs, lhs, op, makeSet(rhsGrouping...), true)
	}
	return e.oneToManyJoin(lhs, rhs, op, makeSet(lhsGrouping...), false)
}

func (e *Engine) fullJoin(
	lhs, rhs []*Timeseries, op binaryOpFunc, grouping map[string]struct{},
) ([]*Timeseries, error) {
	joined := make([]*Timeseries, 0, max(len(lhs), len(rhs)))
	for _, ts := range lhs {
		joined = append(joined, ts.DeepClone())
	}

	m := e.makeTimeseriesMap(joined, grouping)
	for _, rhsTs := range rhs {
		e.buf = rhsTs.Attrs.Bytes(e.buf[:0], grouping)
		hash := xxh3.Hash(e.buf)

		lhsTs, ok := m[hash]
		if !ok {
			lhsTs = rhsTs.DeepClone()
			for i := range lhsTs.Value {
				lhsTs.Value[i] = math.NaN()
			}
			joined = append(joined, lhsTs)
		}

		for i, v1 := range lhsTs.Value {
			v2 := rhsTs.Value[i]
			lhsTs.Value[i] = op(v1, v2)
		}
		lhsTs.Unit = bunconv.UnitNone
	}

	return joined, nil
}

func (e *Engine) oneToManyJoin(
	lhs, rhs []*Timeseries, op binaryOpFunc, grouping map[string]struct{}, swapArgs bool,
) ([]*Timeseries, error) {
	joined := make([]*Timeseries, 0, len(rhs))

	m := e.makeTimeseriesMap(lhs, grouping)
	for _, rhsTs := range rhs {
		e.buf = rhsTs.Attrs.Bytes(e.buf[:0], grouping)
		hash := xxh3.Hash(e.buf)

		var lhsValue []float64
		if lhsTs, ok := m[hash]; ok {
			lhsValue = lhsTs.Value
		} else {
			lhsValue = make([]float64, len(rhsTs.Value))
			for i := range lhsValue {
				lhsValue[i] = math.NaN()
			}
		}

		joinedTs := rhsTs.DeepClone()
		joinedTs.Unit = bunconv.UnitNone
		joined = append(joined, joinedTs)

		if swapArgs {
			for i, v1 := range lhsValue {
				v2 := rhsTs.Value[i]
				joinedTs.Value[i] = op(v2, v1)
			}
		} else {
			for i, v1 := range lhsValue {
				v2 := rhsTs.Value[i]
				joinedTs.Value[i] = op(v1, v2)
			}
		}
	}

	return joined, nil
}

func (e *Engine) makeTimeseriesMap(
	timeseries []*Timeseries, grouping map[string]struct{},
) map[uint64]*Timeseries {
	m := make(map[uint64]*Timeseries, len(timeseries))
	for _, ts := range timeseries {
		e.buf = ts.Attrs.Bytes(e.buf[:0], grouping)
		hash := xxh3.Hash(e.buf)
		m[hash] = ts
	}
	return m
}

func (e *Engine) binaryExprNum(lhs, rhs float64, op ast.BinaryOp) ([]*Timeseries, error) {
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

func (e *Engine) evalBinaryExprNum(lhs, rhs float64, fn binaryOpFunc) ([]*Timeseries, error) {
	ts := e.storage.MakeTimeseries(nil)

	result := fn(lhs, rhs)
	for i := range ts.Value {
		ts.Value[i] = result
	}

	return []*Timeseries{ts}, nil
}

func (e *Engine) binaryExprNumLeft(
	lhs float64, rhs []*Timeseries, op ast.BinaryOp,
) ([]*Timeseries, error) {
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
	lhs float64, rhs []*Timeseries, fn binaryOpFunc,
) ([]*Timeseries, error) {
	joined := make([]*Timeseries, 0, len(rhs))

	for _, rhsTs := range rhs {
		joinedTs := rhsTs.DeepClone()
		joined = append(joined, joinedTs)

		for i, v2 := range rhsTs.Value {
			joinedTs.Value[i] = fn(lhs, v2)
		}
	}

	return joined, nil
}

func (e *Engine) binaryExprNumRight(
	lhs []*Timeseries, rhs float64, op ast.BinaryOp,
) ([]*Timeseries, error) {
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
	lhs []*Timeseries, rhs float64, fn binaryOpFunc,
) ([]*Timeseries, error) {
	joined := make([]*Timeseries, 0, len(lhs))

	for _, lhsTs := range lhs {
		joinedTs := lhsTs.DeepClone()
		joined = append(joined, joinedTs)

		for i, v1 := range lhsTs.Value {
			joinedTs.Value[i] = fn(v1, rhs)
		}
	}

	return joined, nil
}

func (e *Engine) callFunc(fn *FuncCall) ([]*Timeseries, error) {
	switch fn.Func {
	case GoMapDelta:
		return e.callMappingFunc(fn, deltaFunc)
	case GoMapPerMin:
		return e.callMappingFunc(fn, perMinFunc)
	case GoMapPerSec, GoMapRate:
		return e.callMappingFunc(fn, perSecFunc)
	case GoMapIrate:
		return e.callMappingFunc(fn, irateFunc)
	case GoAggMin:
		return e.callAggFunc(fn, minAgg)
	case GoAggMax:
		return e.callAggFunc(fn, maxAgg)
	case GoAggAvg:
		return e.callAggFunc(fn, avgAgg)
	case GoAggSum:
		return e.callAggFunc(fn, sumAgg)
	default:
		if isCHFunc(fn.Func) {
			return nil, fmt.Errorf("func %q must be appied to a metric", fn.Func)
		}
		return nil, fmt.Errorf("unsupported func: %s", fn.Func)
	}
}

func (e *Engine) callMappingFunc(
	fn *FuncCall,
	op FuncOp,
) ([]*Timeseries, error) {
	timeseries, err := e.eval(fn.Arg)
	if err != nil {
		return nil, err
	}

	for _, ts := range timeseries {
		op(ts.Value, e.consts)
	}

	return timeseries, nil
}

func (e *Engine) callAggFunc(fn *FuncCall, op aggFunc) ([]*Timeseries, error) {
	timeseries, err := e.eval(fn.Arg)
	if err != nil {
		return nil, err
	}

	if len(timeseries) == 0 {
		return timeseries, nil
	}

	if len(fn.Grouping) == 0 {
		ts := e.aggTimeseriesAlloc(timeseries, op)
		ts.Attrs = nil
		ts.Grouping = nil
		return []*Timeseries{ts}, nil
	}

	m := make(map[uint64][]*Timeseries)
	var hashes []uint64

	grouping := makeSet(fn.Grouping...)
	for _, ts := range timeseries {
		e.buf = ts.Attrs.Bytes(e.buf[:0], grouping)
		hash := xxh3.Hash(e.buf)

		if _, ok := m[hash]; !ok {
			hashes = append(hashes, hash)
		}
		m[hash] = append(m[hash], ts)
	}

	result := make([]*Timeseries, 0, len(m))
	for _, hash := range hashes {
		ts := e.aggTimeseriesAlloc(m[hash], op)
		ts.Attrs = ts.Attrs.Pick(grouping)
		ts.Grouping = fn.Grouping
		result = append(result, ts)
	}
	return result, nil
}

func (e *Engine) aggTimeseriesAlloc(
	timeseries []*Timeseries,
	fn aggFunc,
) *Timeseries {
	if len(timeseries) == 1 {
		return timeseries[0].Clone()
	}

	ts := timeseries[0].DeepClone()

	tmp := make([]float64, len(timeseries))
	for i := 0; i < len(ts.Value); i++ {
		for j, ts := range timeseries {
			tmp[j] = ts.Value[i]
		}
		ts.Value[i] = fn(tmp)
	}

	return ts
}

func makeSet(slice ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(slice))
	for _, el := range slice {
		m[el] = struct{}{}
	}
	return m
}
