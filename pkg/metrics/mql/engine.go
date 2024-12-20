package mql

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unixtime"
	"go4.org/syncutil"
)

type Engine struct {
	storage Storage

	timeGTE  unixtime.Seconds
	timeLT   unixtime.Seconds
	duration time.Duration
	interval time.Duration

	consts map[string]float64
	vars   map[string][]*Timeseries
	buf    []byte
}

type Storage interface {
	SelectTimeseries(f *TimeseriesFilter) ([]*Timeseries, error)
}

func NewEngine(
	storage Storage,
	timeGTE, timeLT unixtime.Seconds,
	interval time.Duration,
) *Engine {
	return &Engine{
		storage: storage,

		timeGTE:  timeGTE,
		timeLT:   timeLT,
		duration: time.Duration(timeLT-timeGTE) * time.Second,
		interval: interval,

		consts: map[string]float64{
			"_seconds": interval.Seconds(),
			"_minutes": interval.Minutes(),
		},
		vars: make(map[string][]*Timeseries),
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

	rollupWindow := 10 * time.Minute
	for _, expr := range timeseriesExprs {
		if expr.RollupWindow > rollupWindow {
			rollupWindow = expr.RollupWindow
		}
	}
	rollupWindow = min(rollupWindow, 100*e.interval)

	gate := syncutil.NewGate(2)
	for _, expr := range timeseriesExprs {
		expr := expr

		wg.Add(1)
		go func() {
			defer wg.Done()

			f := &TimeseriesFilter{
				Metric: expr.Metric,

				TimeGTE:  e.timeGTE.Time().Add(-expr.Offset).Add(-rollupWindow),
				TimeLT:   e.timeLT.Time().Add(-expr.Offset),
				Interval: e.interval,

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
				return
			}

			expr.Timeseries = timeseries

			if expr.Offset == 0 {
				return
			}
			for _, ts := range timeseries {
				offset := int64(expr.Offset.Seconds())
				for i, t := range ts.Time {
					ts.Time[i] = unixtime.Seconds(int64(t) + offset)
				}
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

		metricName := expr.Alias
		setTimeseriesName(tmp, metricName, expr.NameTemplate(), expr.HasAlias)

		if _, ok := e.vars[metricName]; ok {
			expr.Part.Error.Wrapped = fmt.Errorf("column %q already exists", metricName)
			continue
		}
		e.vars[metricName] = tmp

		if strings.HasPrefix(metricName, "_") {
			continue
		}

		result.Metrics = append(result.Metrics, MetricInfo{
			Name:      metricName,
			TableFunc: TableFuncName(expr.AST),
		})
		for _, ts := range tmp {
			ts.Value, ts.Time = pruneTimeseries(ts.Value, ts.Time, e.timeGTE)
			result.Timeseries = append(result.Timeseries, ts)
		}
	}

	return result
}

func (e *Engine) eval(expr Expr) ([]*Timeseries, error) {
	switch expr := expr.(type) {
	case *TimeseriesExpr:
		return expr.Timeseries, nil
	case RefExpr:
		if timeseries, ok := e.vars[expr.Name]; ok {
			return timeseries, nil
		}

		if num, ok := e.consts[expr.Name]; ok {
			ts := e.makeTimeseries()
			for i := range ts.Value {
				ts.Value[i] = num
			}
			return []*Timeseries{ts}, nil
		}

		return nil, fmt.Errorf("can't resolve name %q", expr.Name)
	case *BinaryExpr:
		return e.binaryExpr(expr)
	case ParenExpr:
		return e.eval(expr.Expr)
	case ast.Number:
		ts := e.makeTimeseries()

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
	if ref, ok := expr.(RefExpr); ok {
		if num, ok := e.consts[ref.Name]; ok {
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
		lhsNum, lhsOK := lhs.(ast.Number)
		rhsNum, rhsOK := rhs.(ast.Number)

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
		hash := xxhash.Sum64(e.buf)

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
		hash := xxhash.Sum64(e.buf)

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
		hash := xxhash.Sum64(e.buf)
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
	ts := e.makeTimeseries()

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

const minRateWindow = 5 * time.Minute

func (e *Engine) callFunc(fn *FuncCall) ([]*Timeseries, error) {
	switch fn.Func {
	case TransformPerMin:
		return e.callTransformFunc(fn, e.perMinTransform)
	case TransformPerSec:
		return e.callTransformFunc(fn, e.perSecTransform)
	case TransformAbs:
		return e.callTransformFunc(fn, absTransform)
	case TransformCeil:
		return e.callTransformFunc(fn, ceilTransform)
	case TransformFloor:
		return e.callTransformFunc(fn, floorTransform)
	case TransformTrunc:
		return e.callTransformFunc(fn, truncTransform)
	case TransformCos:
		return e.callTransformFunc(fn, cosTransform)
	case TransformCosh:
		return e.callTransformFunc(fn, coshTransform)
	case TransformAcos:
		return e.callTransformFunc(fn, acosTransform)
	case TransformAcosh:
		return e.callTransformFunc(fn, acoshTransform)
	case TransformSin:
		return e.callTransformFunc(fn, sinTransform)
	case TransformSinh:
		return e.callTransformFunc(fn, sinhTransform)
	case TransformAsin:
		return e.callTransformFunc(fn, asinTransform)
	case TransformAsinh:
		return e.callTransformFunc(fn, asinhTransform)
	case TransformTan:
		return e.callTransformFunc(fn, tanTransform)
	case TransformTanh:
		return e.callTransformFunc(fn, tanhTransform)
	case TransformAtan:
		return e.callTransformFunc(fn, atanTransform)
	case TransformAtanh:
		return e.callTransformFunc(fn, atanhTransform)
	case TransformExp:
		return e.callTransformFunc(fn, expTransform)
	case TransformExp2:
		return e.callTransformFunc(fn, exp2Transform)
	case TransformLog, TransformLn:
		return e.callTransformFunc(fn, logTransform)
	case TransformLog2:
		return e.callTransformFunc(fn, log2Transform)
	case TransformLog10:
		return e.callTransformFunc(fn, log10Transform)
	case GoAggMin:
		return e.callAggFunc(fn, minAgg)
	case GoAggMax:
		return e.callAggFunc(fn, maxAgg)
	case GoAggAvg:
		return e.callAggFunc(fn, avgAgg)
	case GoAggSum:
		return e.callAggFunc(fn, sumAgg)
	case GoAggMedian:
		return e.callAggFunc(fn, Median)
	case RollupRate, RollupIRate:
		return e.callRollupFunc(fn, rateRollup, max(5*e.interval, minRateWindow))
	case RollupIncrease, RollupDelta:
		return e.callRollupFunc(fn, e.increaseRollup, 0)
	case RollupMinOverTime:
		return e.callRollupFunc(fn, minRollup, 0)
	case RollupMaxOverTime:
		return e.callRollupFunc(fn, maxRollup, 0)
	case RollupSumOverTime:
		return e.callRollupFunc(fn, sumRollup, 0)
	case RollupAvgOverTime:
		return e.callRollupFunc(fn, avgRollup, 0)
	case RollupMedianOverTime:
		return e.callRollupFunc(fn, medianRollup, 0)
	default:
		if isCHFunc(fn.Func) {
			return nil, fmt.Errorf("func %q must be appied to a metric", fn.Func)
		}
		return nil, fmt.Errorf("unsupported func: %s", fn.Func)
	}
}

func (e *Engine) callTransformFunc(
	fn *FuncCall,
	op transformFunc,
) ([]*Timeseries, error) {
	timeseries, err := e.eval(fn.Arg)
	if err != nil {
		return nil, err
	}

	timeseries = cloneDeepTimeseries(timeseries)
	for _, ts := range timeseries {
		op(ts.Value)
	}

	return timeseries, nil
}

func (e *Engine) callRollupFunc(
	fn *FuncCall,
	op rollupFunc,
	window time.Duration,
) ([]*Timeseries, error) {
	switch {
	case window == 0:
		window = e.interval
	case window > e.duration:
		window = e.duration
	}

	if expr, ok := fn.Arg.(*TimeseriesExpr); ok {
		window = max(window, expr.RollupWindow)
	}

	intermediate, err := e.eval(fn.Arg)
	if err != nil {
		return nil, err
	}

	result := cloneDeepTimeseries(intermediate)
	for i, ts := range result {
		op(ts.Value, intermediate[i].Value, ts.Time, window)
	}

	return result, nil
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
		hash := xxhash.Sum64(e.buf)

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

func (e *Engine) makeTimeseries() *Timeseries {
	var ts Timeseries

	period := float64(e.timeLT - e.timeGTE)
	size := int(period/e.interval.Seconds()) + 1
	ts.Value = make([]float64, size)
	ts.Time = make([]unixtime.Seconds, size)

	interval := unixtime.Seconds(e.interval.Seconds())
	for i := range ts.Time {
		ts.Time[i] = e.timeGTE + unixtime.Seconds(i)*interval
	}

	return &ts
}

func setTimeseriesName(timeseries []*Timeseries, metricName, nameTemplate string, hasAlias bool) {
	for _, ts := range timeseries {
		ts.MetricName = metricName
		ts.NameTemplate = nameTemplate
		if hasAlias {
			ts.Filters = nil
		}
	}
}

func pruneTimeseries(
	valueSlice []float64,
	timeSlice []unixtime.Seconds,
	gte unixtime.Seconds,
) ([]float64, []unixtime.Seconds) {
	for i, t := range timeSlice {
		if t >= gte {
			return valueSlice[i:], timeSlice[i:]
		}
	}
	return nil, nil
}

func makeSet(slice ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(slice))
	for _, el := range slice {
		m[el] = struct{}{}
	}
	return m
}

func cloneDeepTimeseries(timeseries []*Timeseries) []*Timeseries {
	mem := make([]Timeseries, len(timeseries))
	clone := make([]*Timeseries, len(timeseries))
	for i := range clone {
		ts := &mem[i]
		clone[i] = ts

		*ts = *timeseries[i]
		ts.Value = slices.Clone(ts.Value)
	}
	return clone
}
