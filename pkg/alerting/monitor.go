package alerting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
)

type Monitor interface {
	Base() *BaseMonitor
}

var (
	_ Monitor = (*MetricMonitor)(nil)
	_ Monitor = (*ErrorMonitor)(nil)
)

type MonitorType string

const (
	MonitorMetric MonitorType = "metric"
	MonitorError  MonitorType = "error"
)

type MonitorState string

const (
	MonitorActive MonitorState = "active"
	MonitorPaused MonitorState = "paused"
	MonitorFiring MonitorState = "firing"
	MonitorNoData MonitorState = "no-data"
	MonitorFailed MonitorState = "failed"
)

type BaseMonitor struct {
	bun.BaseModel `bun:"monitors,alias:m"`

	ID        uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name  string       `json:"name"`
	State MonitorState `json:"state"`

	NotifyEveryoneByEmail bool `json:"notifyEveryoneByEmail"`

	Type   MonitorType    `json:"type"`
	Params bunutil.Params `json:"params"`

	CreatedAt time.Time    `json:"createdAt" bun:",nullzero"`
	UpdatedAt bun.NullTime `json:"updatedAt"`

	ChannelIDs []uint64 `json:"channelIds" bun:"-"`
	AlertCount int      `json:"alertCount" bun:"-"`
}

func (m *BaseMonitor) Base() *BaseMonitor {
	return m
}

type MetricMonitor struct {
	*BaseMonitor `bun:",inherit"`
	Params       MetricMonitorParams `json:"params"`
}

func NewMetricMonitor() *MetricMonitor {
	return &MetricMonitor{
		BaseMonitor: new(BaseMonitor),
	}
}

type MonitorUnit string

const (
	MonitorUnitMinutes = "minutes"
	MonitorUnitHours   = "hours"
)

type MetricMonitorParams struct {
	Metrics    []mql.MetricAlias `json:"metrics"`
	Query      string            `json:"query"`
	Column     string            `json:"column"`
	ColumnUnit string            `json:"columnUnit"`

	ForDuration     int32       `json:"forDuration"`
	ForDurationUnit MonitorUnit `json:"forDurationUnit"`

	MinValue bunutil.NullFloat64 `json:"minValue"`
	MaxValue bunutil.NullFloat64 `json:"maxValue"`
}

func (m *MetricMonitor) Base() *BaseMonitor {
	return m.BaseMonitor
}

func (m *MetricMonitor) MadalarmOptions() ([]madalarm.Option, error) {
	if !m.Params.MinValue.Valid && !m.Params.MaxValue.Valid {
		return nil, errors.New("at least min or max value is required")
	}

	var options []madalarm.Option
	options = append(options, madalarm.WithDuration(int(m.Params.ForDuration)))
	if m.Params.MinValue.Valid {
		options = append(options, madalarm.WithMinValue(m.Params.MinValue.Float64))
	}
	if m.Params.MaxValue.Valid {
		options = append(options, madalarm.WithMaxValue(m.Params.MaxValue.Float64))
	}

	return options, nil
}

func (m *MetricMonitor) ForDuration() time.Duration {
	return time.Duration(m.Params.ForDuration) * time.Minute
}

type ErrorMonitor struct {
	*BaseMonitor `bun:",inherit"`
	Params       ErrorMonitorParams `json:"params"`
}

func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		BaseMonitor: new(BaseMonitor),
	}
}

type ErrorMonitorParams struct {
	NotifyOnNewErrors       bool              `json:"notifyOnNewErrors"`
	NotifyOnRecurringErrors bool              `json:"notifyOnRecurringErrors"`
	Matchers                []org.AttrMatcher `json:"matchers"`
}

func (m *ErrorMonitor) Base() *BaseMonitor {
	return m.BaseMonitor
}

func (m *ErrorMonitor) Matches(span *tracing.Span) bool {
	for i := range m.Params.Matchers {
		if !m.Params.Matchers[i].Matches(span.Attrs) {
			return false
		}
	}
	return true
}

type MonitorChannel struct {
	MonitorID uint64
	ChannelID uint64
}

func SelectMonitor(ctx context.Context, app *bunapp.App, id uint64) (Monitor, error) {
	monitor, err := SelectBaseMonitor(ctx, app, id)
	if err != nil {
		return nil, err
	}
	return decodeMonitor(monitor)
}

func SelectBaseMonitor(ctx context.Context, app *bunapp.App, id uint64) (*BaseMonitor, error) {
	monitor := new(BaseMonitor)
	if err := app.PG.NewSelect().
		Model(monitor).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return monitor, nil
}

func decodeMonitor(base *BaseMonitor) (Monitor, error) {
	switch base.Type {
	case MonitorMetric:
		monitor := &MetricMonitor{
			BaseMonitor: base,
		}
		if err := base.Params.Decode(&monitor.Params); err != nil {
			return nil, err
		}
		return monitor, nil
	case MonitorError:
		monitor := &ErrorMonitor{
			BaseMonitor: base,
		}
		if err := base.Params.Decode(&monitor.Params); err != nil {
			return nil, err
		}
		return monitor, nil
	default:
		return nil, fmt.Errorf("unknown monitor type: %s", base.Type)
	}
}

func PauseMonitor(ctx context.Context, app *bunapp.App, monitorID uint64) error {
	return UpdateMonitorState(ctx, app, monitorID, MonitorActive, MonitorPaused)
}

func UpdateMonitorState(
	ctx context.Context, app *bunapp.App, monitorID uint64, fromState, toState MonitorState,
) error {
	if _, err := app.PG.NewUpdate().
		Model((*BaseMonitor)(nil)).
		Set("state = ?", toState).
		Where("id = ?", monitorID).
		Where("state = ?", fromState).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func SelectMetricMonitors(ctx context.Context, app *bunapp.App) ([]*MetricMonitor, error) {
	monitors := make([]*MetricMonitor, 0)
	if err := app.PG.NewSelect().
		Model(&monitors).
		Where("type = ?", MonitorMetric).
		Where("state NOT IN (?)", MonitorPaused, MonitorFailed).
		Scan(ctx); err != nil {
		return nil, err
	}
	return monitors, nil
}
