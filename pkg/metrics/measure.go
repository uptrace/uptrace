package metrics

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

type Measure struct {
	ch.CHModel `ch:"measure_minutes_stub,insert:measure_minutes_buffer,alias:m"`

	ProjectID   uint32
	Metric      string `ch:"metric,lc"`
	Description string `ch:"-"`
	Unit        string `ch:"-"`
	Instrument  string `ch:",lc"`

	Time      time.Time `ch:"type:DateTime"`
	AttrsHash uint64

	Sum       float64
	Count     uint64
	Value     float64
	Histogram bfloat16.Map `ch:"type:AggregateFunction(quantilesBFloat16(0.5), Float32)"`

	Attrs      AttrMap  `ch:"-"`
	AttrKeys   []string `ch:"type:Array(LowCardinality(String))"`
	AttrValues []string `ch:"type:Array(LowCardinality(String))"`

	StartTimeUnixNano uint64 `ch:"-"`
	CumPoint          any    `ch:"-"`
}

type AttrMap map[string]string

func (m AttrMap) Merge(other AttrMap) {
	for k, v := range other {
		m[k] = v
	}
}

func InsertMeasures(ctx context.Context, app *bunapp.App, measures []*Measure) error {
	_, err := app.CH.NewInsert().
		Model(&measures).
		ModelTableExpr("?", app.DistTable("measure_minutes_buffer")).
		Exec(ctx)
	return err
}

func measureTableForWhere(app *bunapp.App, f *org.TimeFilter) ch.Ident {
	switch org.TablePeriod(f) {
	case time.Minute:
		return app.DistTable("measure_minutes_buffer")
	case time.Hour:
		return app.DistTable("measure_hours")
	}
	panic("not reached")
}

func measureTableForGroup(
	app *bunapp.App, f *org.TimeFilter, groupPeriodFn func(time.Time, time.Time) time.Duration,
) (ch.Ident, time.Duration) {
	tablePeriod, groupPeriod := org.TableGroupPeriod(f)
	switch tablePeriod {
	case time.Minute:
		return app.DistTable("measure_minutes_buffer"), groupPeriod
	case time.Hour:
		return app.DistTable("measure_hours"), groupPeriod
	}
	panic("not reached")
}
