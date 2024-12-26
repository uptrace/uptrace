package tracing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

type QueryBuilder struct {
	Filter *SpanFilter
	Table  *Table
}

func NewQueryBuilder(f *SpanFilter) *QueryBuilder {
	table := TableSpansIndex
	if isLogSystem(f.System...) {
		table = TableLogsIndex
	} else if isEventSystem(f.System...) {
		table = TableEventsIndex
	}

	return &QueryBuilder{
		Filter: f,
		Table:  table,
	}
}

func (qb *QueryBuilder) AppendCHColumn(b []byte, expr *tql.Column, dur time.Duration) ([]byte, error) {
	return qb.appendCHColumn(b, expr, dur)
}

func (qb *QueryBuilder) AppendCHExpr(b []byte, expr tql.Expr, dur time.Duration) ([]byte, error) {
	return qb.appendCHExpr(b, expr, dur)
}

func (qb *QueryBuilder) AppendCHAttr(b []byte, attr tql.Attr) ([]byte, error) {
	return qb.appendCHAttr(b, attr)
}

func (qb *QueryBuilder) appendCHColumn(b []byte, expr *tql.Column, dur time.Duration) ([]byte, error) {
	b, err := qb.appendCHExpr(b, expr.Value, dur)
	if err != nil {
		return nil, err
	}
	b = append(b, " AS "...)
	if expr.Alias != "" {
		b = chschema.AppendName(b, expr.Alias)
	} else {
		b = chschema.AppendName(b, tql.String(expr.Value))
	}
	return b, nil
}

func (qb *QueryBuilder) appendCHExpr(b []byte, expr tql.Expr, dur time.Duration) ([]byte, error) {
	switch expr := expr.(type) {
	case tql.Attr:
		return qb.appendCHAttr(b, expr)
	case *tql.FuncCall:
		return qb.appendCHFuncCall(b, expr, dur)
	case *tql.BinaryExpr:
		b, err := qb.appendCHExpr(b, expr.LHS, dur)
		if err != nil {
			return nil, err
		}

		b = append(b, ' ')
		b = append(b, expr.Op...)
		b = append(b, ' ')

		b, err = qb.appendCHExpr(b, expr.RHS, dur)
		if err != nil {
			return nil, err
		}

		return b, nil
	case tql.ParenExpr:
		b = append(b, '(')
		b, err := qb.appendCHExpr(b, expr, dur)
		if err != nil {
			return nil, err
		}
		b = append(b, ')')
		return b, nil
	case tql.NumberValue:
		b = append(b, expr.Text...)
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported expr: %T", expr)
	}
}

func (qb *QueryBuilder) appendCHAttr(b []byte, attr tql.Attr) ([]byte, error) {
	switch attr.Name {
	case attrkey.SpanErrorCount:
		return chschema.AppendQuery(b, "if(s.status_code = 'error', s.count, 0)"), nil
	case attrkey.SpanErrorRate:
		return chschema.AppendQuery(b, "sumIf(s.count, s.status_code = 'error') / sum(s.count)"), nil
	case attrkey.SpanIsEvent:
		return chschema.AppendQuery(b, "s.type IN ?", ch.In(EventTypes)), nil
	default:
		if strings.HasPrefix(attr.Name, "_") {
			if !qb.Table.IsIndexedAttr(attr.Name) {
				return nil, fmt.Errorf("unsupported attr: %s", attr.Name)
			}

			ident := strings.TrimPrefix(attr.Name, "_")
			b = append(b, "s."...)
			return chschema.AppendIdent(b, ident), nil
		}

		if qb.Table.IsIndexedAttr(attr.Name) {
			b = append(b, "s."...)
			return chschema.AppendIdent(b, attr.Name), nil
		}

		return chschema.AppendQuery(b, "s.string_values[indexOf(s.string_keys, ?)]", attr.Name), nil
	}
}

func (qb *QueryBuilder) appendCHFuncCall(b []byte, fn *tql.FuncCall, dur time.Duration) ([]byte, error) {
	tmp, err := qb.appendCHFuncArg(nil, fn, dur)
	if err != nil {
		return nil, err
	}
	arg := ch.Safe(tmp)

	funcName := fn.Func
	switch funcName {
	case "any_last":
		funcName = "anyLast"
	case "uniq":
		funcName = "uniqCombined64"
	}

	switch funcName {
	case "per_min":
		return chschema.AppendQuery(b, "? / ?", arg, dur.Minutes()), nil
	case "per_sec":
		return chschema.AppendQuery(b, "? / ?", arg, dur.Seconds()), nil
	case "p50", "p75", "p90", "p99":
		return chschema.AppendQuery(b, "quantileTDigest(?)(?)",
			quantileLevel(funcName), arg), nil
	case "top3":
		return chschema.AppendQuery(b, "topK(3)(?)", arg), nil
	case "top5":
		return chschema.AppendQuery(b, "topK(5)(?)", arg), nil
	case "top10":
		return chschema.AppendQuery(b, "topK(10)(?)", arg), nil
	case "sum", "avg", "min", "max",
		"any", "anyLast", "uniqCombined64":
		b = append(b, funcName...)
		b = append(b, '(')
		b = append(b, arg...)
		b = append(b, ')')
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported func: %s", fn.Func)
	}
}

func (qb *QueryBuilder) appendCHFuncArg(b []byte, fn *tql.FuncCall, dur time.Duration) ([]byte, error) {
	convNum := isNumFunc(fn.Func) && !isNumExpr(fn.Arg)
	if convNum {
		b = append(b, "toFloat64OrDefault("...)
	}

	b, err := qb.appendCHExpr(b, fn.Arg, dur)
	if err != nil {
		return nil, err
	}

	if convNum {
		b = append(b, ')')
	}

	return b, nil
}

//------------------------------------------------------------------------------

func (qb *QueryBuilder) AppendWhereHaving(ast *tql.Where, dur time.Duration) ([]byte, []byte, error) {
	var where []byte
	var having []byte
	var firstErr error

	for _, filter := range ast.Filters {
		bb, err := qb.AppendFilter(filter, dur)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		if isAggExpr(filter.LHS) {
			having = appendFilter(having, filter, bb)
		} else {
			where = appendFilter(where, filter, bb)
		}
	}

	return where, having, firstErr
}

func (qb *QueryBuilder) AppendFilter(filter tql.Filter, dur time.Duration) ([]byte, error) {
	var b []byte

	switch filter.Op {
	case tql.FilterExists, tql.FilterNotExists:
		attrKey := tql.String(filter.LHS)

		if strings.HasPrefix(attrKey, ".") {
			if filter.Op == tql.FilterNotExists {
				b = append(b, '0')
			} else {
				b = append(b, '1')
			}
			return b, nil
		}

		if filter.Op == tql.FilterNotExists {
			b = append(b, "NOT "...)
		}
		b = chschema.AppendQuery(b, "has(s.all_keys, ?)", attrKey)
		return b, nil
	case tql.FilterIn, tql.FilterNotIn:
		if filter.Op == tql.FilterNotIn {
			b = append(b, "NOT "...)
		}

		b, err := qb.appendCHExpr(b, filter.LHS, dur)
		if err != nil {
			return nil, err
		}

		b = append(b, " IN "...)
		b = chschema.AppendQuery(b, "?", ch.In(filter.RHS.Values()))
		return b, nil
	case tql.FilterContains, tql.FilterNotContains:
		if filter.Op == tql.FilterNotContains {
			b = append(b, "NOT "...)
		}

		values := strings.Split(filter.RHS.String(), "|")
		b = append(b, "multiSearchAnyCaseInsensitiveUTF8("...)
		b, err := qb.appendCHExpr(b, filter.LHS, dur)
		if err != nil {
			return nil, err
		}
		b = append(b, ", "...)
		b = chschema.AppendQuery(b, "?", ch.Array(values))
		b = append(b, ")"...)

		return b, nil
	}

	var convToNum bool
	if _, ok := filter.RHS.(*tql.NumberValue); ok {
		convToNum = !isNumExpr(filter.LHS)
	}

	if convToNum {
		b = append(b, "toFloat64OrDefault("...)
	}
	b, err := qb.appendCHExpr(b, filter.LHS, dur)
	if err != nil {
		return nil, err
	}
	if convToNum {
		b = append(b, ')')
	}

	b = append(b, ' ')
	b = append(b, filter.Op...)
	b = append(b, ' ')

	switch value := filter.RHS.(type) {
	case tql.NumberValue:
		if convToNum {
			b = append(b, "toFloat64OrDefault("...)
		}

		switch value.Kind {
		case tql.NumberDuration:
			dur, err := time.ParseDuration(value.Text)
			if err != nil {
				panic(err)
			}
			b = strconv.AppendInt(b, int64(dur), 10)
		case tql.NumberBytes:
			n, err := bunconv.ParseBytes(value.Text)
			if err != nil {
				panic(err)
			}
			b = strconv.AppendInt(b, n, 10)
		default:
			b = append(b, value.Text...)
		}

		if convToNum {
			b = append(b, ')')
		}
	default:
		b = chschema.AppendString(b, value.String())
	}

	return b, nil
}

func appendFilter(b []byte, filter tql.Filter, bb []byte) []byte {
	if len(b) > 0 {
		b = append(b, ' ')
		b = append(b, filter.BoolOp...)
		b = append(b, ' ')
	}
	return append(b, bb...)
}

//------------------------------------------------------------------------------

func unitForExpr(expr tql.Expr) string {
	switch expr := expr.(type) {
	case tql.Attr:
		switch expr.Name {
		case attrkey.SpanTime:
			return bunconv.UnitUnixTime
		case attrkey.SpanErrorRate:
			return bunconv.UnitUtilization
		case attrkey.SpanDuration:
			return bunconv.UnitNanoseconds
		default:
			return bunconv.UnitNone
		}
	case *tql.FuncCall:
		unit := unitForExpr(expr.Arg)
		switch expr.Func {
		case "",
			"sum", "avg", "min", "max",
			"any", "anyLast", "any_last",
			"p50", "p75", "p90", "p95", "p99":
			return unit
		default:
			return bunconv.UnitNone
		}
	case *tql.BinaryExpr:
		return unitForExpr(expr.LHS)
	case tql.ParenExpr:
		return unitForExpr(expr.Expr)
	default:
		return bunconv.UnitNone
	}
}

func isAggExpr(expr tql.Expr) bool {
	switch expr := expr.(type) {
	case tql.Attr:
		switch expr.Name {
		case attrkey.SpanErrorRate:
			return true
		default:
			return false
		}
	case *tql.FuncCall:
		switch expr.Func {
		case "sum", "avg", "min", "max",
			"any", "anyLast", "any_last",
			"uniq",
			"p50", "p75", "p90", "p95", "p99":
			return true
		case "per_min", "per_sec":
			return isAggExpr(expr.Arg)
		default:
			return false
		}
	case *tql.BinaryExpr:
		return isAggExpr(expr.LHS) && isAggExpr(expr.RHS)
	case tql.ParenExpr:
		return isAggExpr(expr.Expr)
	default:
		return false
	}
}

func isNumExpr(expr tql.Expr) bool {
	switch expr := expr.(type) {
	case tql.Attr:
		switch expr.Name {
		case attrkey.SpanID,
			attrkey.SpanParentID,
			attrkey.SpanGroupID,
			attrkey.SpanDuration,

			attrkey.SpanLinkCount,
			attrkey.SpanEventCount,
			attrkey.SpanEventErrorCount,
			attrkey.SpanEventLogCount,

			attrkey.SpanCount,
			attrkey.SpanErrorCount,
			attrkey.SpanErrorRate:
			return true
		default:
			return false
		}
	case *tql.FuncCall:
		if !isNumFunc(expr.Func) {
			return false
		}
		return isNumExpr(expr.Arg)
	case *tql.BinaryExpr:
		return true
	case tql.ParenExpr:
		return isNumExpr(expr.Expr)
	default:
		return false
	}
}

func isNumFunc(name string) bool {
	switch name {
	case "sum", "avg", "min", "max",
		"uniq",
		"p50", "p75", "p90", "p95", "p99",
		"per_min", "per_sec":
		return true
	default:
		return false
	}
}
