package metrics

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/uptrace/pkg/org"
)

const (
	TableDatapointMinutes = "datapoint_minutes"
	TableDatapointHours   = "datapoint_hours"
)

type Datapoint struct {
	ch.CHModel `ch:"datapoint_minutes,alias:m"`

	ProjectID   uint32
	Metric      string     `ch:"metric,lc"`
	Description string     `ch:"-"`
	Unit        string     `ch:"-"`
	Instrument  Instrument `ch:",lc"`

	Time      time.Time `ch:"type:DateTime"`
	AttrsHash uint64

	Min       float64
	Max       float64
	Sum       float64
	Count     uint64
	Gauge     float64
	Histogram map[bfloat16.T]uint64 `ch:"type:AggregateFunction(quantilesBFloat16(0.5), Float32)"`

	Attrs        AttrMap  `ch:"-"`
	StringKeys   []string `ch:"type:Array(LowCardinality(String))"`
	StringValues []string

	OtelLibraryName    string `ch:",lc"`
	OtelLibraryVersion string `ch:",lc"`

	StartTimeUnixNano uint64 `ch:"-"`
	CumPoint          any    `ch:"-"`
}

type AttrMap map[string]string

func (m AttrMap) Merge(other AttrMap) {
	for k, v := range other {
		m[k] = v
	}
}

func InsertDatapoints(ctx context.Context, ch *ch.DB, datapoints []*Datapoint) error {
	_, err := ch.NewInsert().
		Model(&datapoints).
		Exec(ctx)
	return err
}

func DatapointTableForWhere(f *org.TimeFilter) string {
	return datapointTable(org.TableWhereResolution(f))
}

func DatapointTableForGrouping(
	f *org.TimeFilter, groupingIntervalFn func(time.Time, time.Time) time.Duration,
) (string, time.Duration) {
	tableResolution, groupingInterval := org.GroupingInterval(f)
	tableName := datapointTable(tableResolution)
	return tableName, groupingInterval
}

func datapointTable(tableResolution time.Duration) string {
	switch tableResolution {
	case time.Minute:
		return TableDatapointMinutes
	case time.Hour:
		return TableDatapointHours
	}
	panic("not reached")
}
