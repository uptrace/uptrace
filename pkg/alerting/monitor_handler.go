package alerting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type MonitorFilter struct {
	*bunapp.App
	urlstruct.Pager

	ProjectID uint32
	State     string
}

func decodeMonitorFilter(app *bunapp.App, req bunrouter.Request) (*MonitorFilter, error) {
	ctx := req.Context()

	f := new(MonitorFilter)
	f.App = app
	f.ProjectID = org.ProjectFromContext(ctx).ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *MonitorFilter) whereClause(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.Where("project_id = ?", f.ProjectID)

	if f.State != "" {
		q = q.Where("state = ?", f.State)
	}

	return q
}

var _ urlstruct.ValuesUnmarshaler = (*MonitorFilter)(nil)

func (f *MonitorFilter) UnmarshalValues(ctx context.Context, values url.Values) (err error) {
	if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type MonitorHandler struct {
	*bunapp.App
}

func NewMonitorHandler(app *bunapp.App) *MonitorHandler {
	return &MonitorHandler{
		App: app,
	}
}

func (h *MonitorHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeMonitorFilter(h.App, req)
	if err != nil {
		return err
	}

	var baseMonitors []*BaseMonitor

	if err := h.PG.NewSelect().
		Model(&baseMonitors).
		Apply(f.whereClause).
		OrderExpr("created_at DESC").
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		Scan(ctx); err != nil {
		return err
	}

	monitors := make([]Monitor, len(baseMonitors))

	for i, baseMonitor := range baseMonitors {
		monitor, err := decodeMonitor(baseMonitor)
		if err != nil {
			return err
		}
		monitors[i] = monitor
	}

	states, err := SelectMonitorStatesCount(ctx, f)
	if err != nil {
		return err
	}

	if err := CountMonitorAlerts(ctx, h.App, baseMonitors); err != nil {
		h.App.Zap(ctx).Error("CountMonitorAlerts failed", zap.Error(err))
	}

	return httputil.JSON(w, bunrouter.H{
		"monitors": monitors,
		"states":   states,
	})
}

func CountMonitorAlerts(ctx context.Context, app *bunapp.App, monitors []*BaseMonitor) error {
	var group syncutil.Group

	for _, monitor := range monitors {
		monitor := monitor
		group.Go(func() error {
			count, err := app.PG.NewSelect().
				Model((*org.BaseAlert)(nil)).
				Where("monitor_id = ?", monitor.ID).
				Count(ctx)
			if err != nil {
				return err
			}

			monitor.AlertCount = count
			return nil
		})
	}

	return group.Err()
}

type StateCount struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

func SelectMonitorStatesCount(ctx context.Context, f *MonitorFilter) ([]StateCount, error) {
	var states []StateCount
	if err := f.App.PG.NewSelect().
		Model((*BaseMonitor)(nil)).
		ColumnExpr("state").
		ColumnExpr("count(*)").
		Where("project_id = ?", f.ProjectID).
		Group("state").
		Scan(ctx, &states); err != nil {
		return nil, err
	}
	return states, nil
}

func (h *MonitorHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	monitor := monitorFromContext(ctx)
	base := monitor.Base()

	var err error

	base.ChannelIDs, err = h.selectChannelIDs(ctx, base.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"monitor": monitor,
	})
}

func (h *MonitorHandler) selectChannelIDs(ctx context.Context, monitorID uint64) ([]uint64, error) {
	ids := make([]uint64, 0)
	if err := h.PG.NewSelect().
		ColumnExpr("channel_id").
		Model((*MonitorChannel)(nil)).
		Where("monitor_id = ?", monitorID).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

type MetricMonitorIn struct {
	Name string `json:"name"`

	NotifyEveryoneByEmail bool                `json:"notifyEveryoneByEmail"`
	Params                MetricMonitorParams `json:"params"`

	ChannelIDs []uint64 `json:"channelIds"`
}

func (in *MetricMonitorIn) Validate(
	ctx context.Context, app *bunapp.App, monitor *MetricMonitor,
) error {
	if in.Name == "" {
		return errors.New("name can't be empty")
	}

	if len(in.Params.Metrics) == 0 {
		return errors.New("at least one metric is required")
	}
	if in.Params.Query == "" {
		return errors.New("query can't be empty")
	}
	if in.Params.Column == "" {
		return errors.New("column can't be empty")
	}

	if in.Params.ForDuration == 0 {
		return errors.New("forDuration can't be zero")
	}
	switch in.Params.ForDurationUnit {
	case MonitorUnitMinutes, MonitorUnitHours:
	default:
		return fmt.Errorf("unsupported duration unit: %q", in.Params.ForDurationUnit)
	}

	for _, ma := range in.Params.Metrics {
		if _, err := metrics.SelectMetricByName(
			ctx, app, monitor.ProjectID, ma.Name,
		); err == sql.ErrNoRows {
			return fmt.Errorf("metric %s does not exist", ma.Name)
		}
	}

	monitor.Name = in.Name
	monitor.NotifyEveryoneByEmail = in.NotifyEveryoneByEmail
	monitor.Params = in.Params

	options, err := monitor.MadalarmOptions()
	if err != nil {
		return err
	}

	if _, err := madalarm.Check(make([]float64, 100), options...); err != nil {
		return err
	}

	return nil
}

func (h *MonitorHandler) CreateMetricMonitor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in MetricMonitorIn

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	monitor := NewMetricMonitor()
	monitor.ProjectID = project.ID
	monitor.State = MonitorActive
	monitor.Type = MonitorMetric

	if err := in.Validate(ctx, h.App, monitor); err != nil {
		return err
	}

	if _, err := h.PG.NewInsert().
		Model(monitor).
		Exec(ctx); err != nil {
		return err
	}

	if err := h.insertMonitorChannels(ctx, monitor.BaseMonitor, in.ChannelIDs); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"monitor": monitor,
	})
}

func (h *MonitorHandler) UpdateMetricMonitor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	monitor, err := metricMonitorFromContext(ctx)
	if err != nil {
		return err
	}

	var in MetricMonitorIn

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}
	if err := in.Validate(ctx, h.App, monitor); err != nil {
		return err
	}

	if err := h.PG.NewUpdate().
		Model(monitor).
		Column("name").
		Column("notify_everyone_by_email").
		Column("params").
		Column("updated_at").
		Where("id = ?", monitor.ID).
		Returning("*").
		Scan(ctx); err != nil {
		return err
	}

	if _, err := h.PG.NewDelete().
		Model((*MonitorChannel)(nil)).
		Where("monitor_id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}
	if err := h.insertMonitorChannels(ctx, monitor.Base(), in.ChannelIDs); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"monitor": monitor,
	})
}

type ErrorMonitorIn struct {
	Name string `json:"name"`

	NotifyEveryoneByEmail bool               `json:"notifyEveryoneByEmail"`
	Params                ErrorMonitorParams `json:"params"`

	ChannelIDs []uint64 `json:"channelIds"`
}

func (in *ErrorMonitorIn) Validate(
	ctx context.Context, app *bunapp.App, monitor *ErrorMonitor,
) error {
	if in.Name == "" {
		return errors.New("name can't be empty")
	}

	if in.Params.Matchers == nil {
		in.Params.Matchers = make([]org.AttrMatcher, 0)
	}

	monitor.Name = in.Name
	monitor.NotifyEveryoneByEmail = in.NotifyEveryoneByEmail
	monitor.Params = in.Params

	return nil
}

func (h *MonitorHandler) CreateErrorMonitor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in ErrorMonitorIn

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	monitor := NewErrorMonitor()
	monitor.ProjectID = project.ID
	monitor.State = MonitorActive
	monitor.Type = MonitorError

	if err := in.Validate(ctx, h.App, monitor); err != nil {
		return err
	}

	if _, err := h.PG.NewInsert().
		Model(monitor).
		Exec(ctx); err != nil {
		return err
	}

	if err := h.insertMonitorChannels(ctx, monitor.BaseMonitor, in.ChannelIDs); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"monitor": monitor,
	})
}

func (h *MonitorHandler) UpdateErrorMonitor(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	monitor, err := errorMonitorFromContext(ctx)
	if err != nil {
		return err
	}

	var in ErrorMonitorIn

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}
	if err := in.Validate(ctx, h.App, monitor); err != nil {
		return err
	}
	monitor.UpdatedAt = bun.NullTime{Time: time.Now()}

	if err := h.PG.NewUpdate().
		Model(monitor).
		Column("name").
		Column("notify_everyone_by_email").
		Column("params").
		Column("updated_at").
		Where("id = ?", monitor.ID).
		Returning("*").
		Scan(ctx); err != nil {
		return err
	}

	if _, err := h.PG.NewDelete().
		Model((*MonitorChannel)(nil)).
		Where("monitor_id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}
	if err := h.insertMonitorChannels(ctx, monitor.Base(), in.ChannelIDs); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"monitor": monitor,
	})
}

func (h *MonitorHandler) insertMonitorChannels(
	ctx context.Context,
	monitor *BaseMonitor,
	channelIDs []uint64,
) error {
	for _, channelID := range channelIDs {
		mc := &MonitorChannel{
			MonitorID: monitor.ID,
			ChannelID: channelID,
		}
		if _, err := h.PG.NewInsert().
			Model(mc).
			On("CONFLICT DO NOTHING").
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (h *MonitorHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	monitor := monitorFromContext(ctx).Base()

	if _, err := h.PG.NewDelete().
		Model(monitor).
		Where("id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *MonitorHandler) Activate(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateState(w, req, MonitorActive)
}

func (h *MonitorHandler) Pause(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateState(w, req, MonitorPaused)
}

func (h *MonitorHandler) updateState(
	w http.ResponseWriter, req bunrouter.Request, state MonitorState,
) error {
	ctx := req.Context()

	monitor := monitorFromContext(ctx).Base()
	monitor.State = state

	if _, err := h.PG.NewUpdate().
		Model((*BaseMonitor)(nil)).
		Set("state = ?", state).
		Where("id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
