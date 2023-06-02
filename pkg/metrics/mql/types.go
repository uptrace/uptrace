package mql

import "errors"

const (
	AggMin   = "min"
	AggMax   = "max"
	AggSum   = "sum"
	AggCount = "count"
	AggAvg   = "avg"

	AggP50 = "p50"
	AggP75 = "p75"
	AggP90 = "p90"
	AggP95 = "p95"
	AggP99 = "p99"

	AggUniq = "uniq"
)

const (
	FuncDelta  = "delta"
	FuncPerMin = "per_min"
	FuncPerSec = "per_sec"
)

const (
	TableMin  = "min"
	TableMax  = "max"
	TableSum  = "sum"
	TableAvg  = "avg"
	TableLast = "last"
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

func isAggFunc(name string) bool {
	switch name {
	case AggMin, AggMax, AggSum, AggCount, AggAvg,
		AggP50, AggP75, AggP90, AggP95, AggP99,
		AggUniq:
		return true
	default:
		return false
	}
}

func isTableFunc(name string) bool {
	switch name {
	case TableMin, TableMax, TableSum,
		TableAvg, TableLast:
		return true
	default:
		return false
	}
}

func isOpFunc(name string) bool {
	switch name {
	case FuncDelta, FuncPerMin, FuncPerSec:
		return true
	default:
		return false
	}
}
