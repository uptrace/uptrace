package pgmigrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
)

type MetricMonitor struct {
	org.BaseMonitor `bun:",inherit"`
	Params          MetricMonitorParams `json:"params"`
}

type MetricMonitorParams struct {
	Metrics    []mql.MetricAlias `json:"metrics"`
	Query      string            `json:"query"`
	Column     string            `json:"column"`
	ColumnUnit string            `json:"columnUnit"`

	// Common params

	CheckNumPoint int           `json:"checkNumPoint"`
	TimeOffset    time.Duration `json:"timeOffset"`

	MinValue bunutil.NullFloat64 `json:"minValue"`
	MaxValue bunutil.NullFloat64 `json:"maxValue"`

	MinAllowedValue bunutil.NullFloat64 `json:"minAllowedValue"`
	MaxAllowedValue bunutil.NullFloat64 `json:"maxAllowedValue"`
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var monitors []*MetricMonitor

		if err := db.NewSelect().
			Model(&monitors).
			Where("type = ?", org.MonitorMetric).
			Scan(ctx); err != nil {
			return err
		}

		for _, monitor := range monitors {
			monitor.Params.MinAllowedValue = monitor.Params.MinValue
			monitor.Params.MaxAllowedValue = monitor.Params.MaxValue

			if _, err := db.NewUpdate().
				Model(monitor).
				Set("params = ?", monitor.Params).
				Where("id = ?", monitor.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
