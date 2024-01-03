package pgmigrations

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var monitors []*MetricMonitor

		if err := db.NewSelect().
			Model(&monitors).
			Where("type = ?", org.MonitorMetric).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range monitors {
			for i := range item.Params.Metrics {
				metric := &item.Params.Metrics[i]
				metric.Name = updateMetricName(metric.Name)
			}
			item.Params.Query = updateMetricQuery(item.Params.Query)

			params := &org.MetricMonitorParams{
				Metrics:    item.Params.Metrics,
				Query:      item.Params.Query,
				Column:     item.Params.Column,
				ColumnUnit: item.Params.ColumnUnit,

				CheckNumPoint: int(item.Params.ForDuration),

				MinAllowedValue: item.Params.MinValue,
				MaxAllowedValue: item.Params.MaxValue,
			}

			if _, err := db.NewUpdate().
				Model(item).
				Set("params = ?", params).
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}

type BaseMonitor struct {
	bun.BaseModel `bun:"monitors,alias:m"`

	ID        uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name  string           `json:"name"`
	State org.MonitorState `json:"state"`

	NotifyEveryoneByEmail bool `json:"notifyEveryoneByEmail"`

	Type   org.MonitorType `json:"type"`
	Params bunutil.Params  `json:"params"`

	CreatedAt time.Time    `json:"createdAt" bun:",nullzero"`
	UpdatedAt bun.NullTime `json:"updatedAt"`

	ChannelIDs []uint64 `json:"channelIds" bun:"-"`
	AlertCount int      `json:"alertCount" bun:"-"`
}

type MetricMonitor struct {
	*BaseMonitor `bun:",inherit"`
	Params       MetricMonitorParams `json:"params"`
}

type MetricMonitorParams struct {
	Metrics    []mql.MetricAlias `json:"metrics"`
	Query      string            `json:"query"`
	Column     string            `json:"column"`
	ColumnUnit string            `json:"columnUnit"`

	ForDuration int32 `json:"forDuration"`

	MinValue bunutil.NullFloat64 `json:"minValue"`
	MaxValue bunutil.NullFloat64 `json:"maxValue"`
}
