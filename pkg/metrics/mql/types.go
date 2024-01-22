package mql

import (
	"errors"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
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

func TableFuncName(expr ast.Expr) string {
	fn, ok := expr.(*ast.FuncCall)
	if !ok {
		return TableFuncMedian
	}

	switch fn.Func {
	case CHAggMin, CHAggMax:
		return fn.Func
	case CHAggCount:
		return TableFuncSum
	case RollupIncrease, RollupDelta:
		return TableFuncSum
	default:
		return TableFuncMedian
	}
}
