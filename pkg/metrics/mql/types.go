package mql

import (
	"errors"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
)

const (
	CHAggMin   = "min"
	CHAggMax   = "max"
	CHAggSum   = "sum"
	CHAggCount = "count"
	CHAggAvg   = "avg"

	CHAggP50 = "p50"
	CHAggP75 = "p75"
	CHAggP90 = "p90"
	CHAggP95 = "p95"
	CHAggP99 = "p99"

	CHAggUniq = "uniq"
)

const (
	GoMapDelta  = "delta"
	GoMapPerMin = "per_min"
	GoMapPerSec = "per_sec"
	GoMapRate   = "rate"
	GoMapIrate  = "irate"
)

const (
	GoAggMin = "min"
	GoAggMax = "max"
	GoAggAvg = "avg"
	GoAggSum = "sum"
)

const (
	TableFuncMin  = "min"
	TableFuncMax  = "max"
	TableFuncAvg  = "avg"
	TableFuncSum  = "sum"
	TableFuncLast = "last"
)

type MetricAlias struct {
	Name  string `yaml:"name" json:"name"`
	Alias string `yaml:"alias" json:"alias"`
}

func (m *MetricAlias) String() string {
	return m.Name + " as $" + m.Alias
}

func (m *MetricAlias) Validate() error {
	if m.Name == "" {
		return errors.New("metric name can't be empty")
	}
	if m.Alias == "" {
		return errors.New("metric alias can't be empty")
	}
	return nil
}

func isCHFunc(name string) bool {
	switch name {
	case CHAggMin, CHAggMax, CHAggSum, CHAggCount, CHAggAvg,
		CHAggP50, CHAggP75, CHAggP90, CHAggP95, CHAggP99,
		CHAggUniq:
		return true
	default:
		return false
	}
}

func isGoMapFunc(name string) bool {
	switch name {
	case GoMapDelta, GoMapPerMin, GoMapPerSec:
		return true
	default:
		return false
	}
}

func TableFuncName(expr ast.Expr) string {
	fn, ok := expr.(*ast.FuncCall)
	if !ok {
		return TableFuncLast
	}

	switch fn.Func {
	case CHAggMin, CHAggMax, CHAggAvg:
		return fn.Func
	case CHAggSum, CHAggCount:
		return TableFuncSum
	case CHAggP50, CHAggP75, CHAggP90, CHAggP95, CHAggP99:
		return TableFuncAvg
	case CHAggUniq:
		return TableFuncLast
	case GoMapDelta:
		return TableFuncSum
	case GoMapPerMin, GoMapPerSec:
		return TableFuncAvg
	default:
		return TableFuncLast
	}
}
