package org

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/unixtime"
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

func (m *BaseMonitor) Validate() error {
	if m.Name == "" {
		return errors.New("monitor name can't be empty")
	}
	return nil
}

type MetricMonitor struct {
	*BaseMonitor `bun:",inherit"`
	Params       MetricMonitorParams `json:"params"`
}

func NewMetricMonitor() *MetricMonitor {
	return &MetricMonitor{
		BaseMonitor: &BaseMonitor{
			Type:  MonitorMetric,
			State: MonitorActive,
		},
	}
}

type MetricMonitorParams struct {
	Metrics    []mql.MetricAlias `json:"metrics"`
	Query      string            `json:"query"`
	Column     string            `json:"column"`
	ColumnUnit string            `json:"columnUnit"`

	CheckNumPoint int             `json:"checkNumPoint"`
	TimeOffset    unixtime.Millis `json:"timeOffset"`

	MinAllowedValue bunutil.NullFloat64 `json:"minAllowedValue"`
	MaxAllowedValue bunutil.NullFloat64 `json:"maxAllowedValue"`
}

func (m *MetricMonitor) Base() *BaseMonitor {
	return m.BaseMonitor
}

func (m *MetricMonitor) Validate() error {
	if err := m.BaseMonitor.Validate(); err != nil {
		return err
	}

	if len(m.Params.Metrics) == 0 {
		return errors.New("at least one metric is required")
	}
	if m.Params.Query == "" {
		return errors.New("query can't be empty")
	}
	if m.Params.Column == "" {
		return errors.New("column can't be empty")
	}
	if m.Params.TimeOffset > 300*unixtime.MillisOf(time.Minute) {
		return errors.New("time offset can't can't be larger 300 minutes")
	}

	if m.Params.CheckNumPoint == 0 {
		m.Params.CheckNumPoint = 5
	}

	return nil
}

func (m *MetricMonitor) MadalarmOptions() ([]madalarm.Option, error) {
	if !m.Params.MinAllowedValue.Valid && !m.Params.MaxAllowedValue.Valid {
		return nil, errors.New("at least min or max value is required")
	}

	var options []madalarm.Option
	options = append(options, madalarm.WithDuration(m.Params.CheckNumPoint))
	if m.Params.MinAllowedValue.Valid {
		options = append(options, madalarm.WithMinValue(m.Params.MinAllowedValue.Float64))
	}
	if m.Params.MaxAllowedValue.Valid {
		options = append(options, madalarm.WithMaxValue(m.Params.MaxAllowedValue.Float64))
	}

	return options, nil
}

type ErrorMonitor struct {
	*BaseMonitor `bun:",inherit"`
	Params       ErrorMonitorParams `json:"params"`
}

func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		BaseMonitor: &BaseMonitor{
			Type:  MonitorError,
			State: MonitorActive,
		},
	}
}

type ErrorMonitorParams struct {
	NotifyOnNewErrors       bool          `json:"notifyOnNewErrors"`
	NotifyOnRecurringErrors bool          `json:"notifyOnRecurringErrors"`
	Matchers                []AttrMatcher `json:"matchers"`
}

func (m *ErrorMonitor) Base() *BaseMonitor {
	return m.BaseMonitor
}

func (m *ErrorMonitor) Validate() error {
	if err := m.BaseMonitor.Validate(); err != nil {
		return err
	}
	if m.Params.Matchers == nil {
		m.Params.Matchers = make([]AttrMatcher, 0)
	}
	return nil
}

type MonitorChannel struct {
	MonitorID uint64
	ChannelID uint64
}

func InsertMonitor(ctx context.Context, db bun.IDB, monitor Monitor) error {
	if _, err := db.NewInsert().
		Model(monitor).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func SelectMonitor(ctx context.Context, pg *bun.DB, id uint64) (Monitor, error) {
	monitor, err := SelectBaseMonitor(ctx, pg, id)
	if err != nil {
		return nil, err
	}
	return DecodeMonitor(monitor)
}

func SelectBaseMonitor(ctx context.Context, pg *bun.DB, id uint64) (*BaseMonitor, error) {
	monitor := new(BaseMonitor)
	if err := pg.NewSelect().
		Model(monitor).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return monitor, nil
}

func DecodeMonitor(base *BaseMonitor) (Monitor, error) {
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

func PauseMonitor(ctx context.Context, pg *bun.DB, monitorID uint64) error {
	return UpdateMonitorState(ctx, pg, monitorID, MonitorActive, MonitorPaused)
}

func UpdateMonitorState(
	ctx context.Context, pg *bun.DB, monitorID uint64, fromState, toState MonitorState,
) error {
	if _, err := pg.NewUpdate().
		Model((*BaseMonitor)(nil)).
		Set("state = ?", toState).
		Where("id = ?", monitorID).
		Where("state = ?", fromState).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
