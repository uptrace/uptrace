package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	"gopkg.in/yaml.v3"
)

type MonitorFilter struct {
	*bunapp.App
	urlstruct.Pager
	org.OrderByMixin

	Q string

	ProjectID uint32
	MonitorID uint64
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

	if err := f.extractParamsFromQuery(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *MonitorFilter) whereClause(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.Where("project_id = ?", f.ProjectID)

	if f.MonitorID != 0 {
		q = q.Where("id = ?", f.MonitorID)
	}
	if f.Q != "" {
		q = q.Where("word_similarity(?, name) >= 0.3", f.Q)
	}
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
	if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

func (f *MonitorFilter) PGOrder(q *bun.SelectQuery) *bun.SelectQuery {
	if f.SortBy == "" {
		return q
	}

	var sortBy string

	switch f.SortBy {
	case "name", "type", "state":
		sortBy = f.SortBy
	default:
		sortBy = "updated_at"
	}

	return q.OrderExpr("? ? NULLS LAST", bun.Ident(sortBy), bun.Safe(f.SortDir()))
}

func (f *MonitorFilter) extractParamsFromQuery() error {
	parts := strings.Split(f.Q, " ")

	for i, part := range parts {
		ss := strings.Split(part, ":")

		if len(ss) != 2 {
			continue
		}

		switch ss[0] {
		case "monitor":
			monitorID, err := strconv.ParseUint(ss[1], 10, 64)
			if err != nil {
				return err
			}

			f.MonitorID = monitorID

			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	f.Q = strings.Join(parts, " ")

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

type MonitorOut struct {
	org.BaseMonitor  `bun:",extend"`
	AlertOpenCount   int `json:"alertOpenCount" bun:",scanonly"`
	AlertClosedCount int `json:"alertClosedCount" bun:",scanonly"`
}

func (h *MonitorHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeMonitorFilter(h.App, req)
	if err != nil {
		return err
	}

	var monitors []*MonitorOut

	count, err := h.PG.NewSelect().
		Model(&monitors).
		Apply(f.whereClause).
		Apply(f.PGOrder).
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		ScanAndCount(ctx)
	if err != nil {
		return err
	}

	states, err := SelectMonitorStatesCount(ctx, f)
	if err != nil {
		return err
	}

	if err := h.countMonitorAlerts(ctx, monitors); err != nil {
		h.Zap(ctx).Error("countMonitorAlerts failed", zap.Error(err))
	}

	return httputil.JSON(w, bunrouter.H{
		"monitors": monitors,
		"states":   states,
		"count":    count,
	})
}

func (h *MonitorHandler) countMonitorAlerts(
	ctx context.Context, monitors []*MonitorOut,
) error {
	var group syncutil.Group

	for _, monitor := range monitors {
		monitor := monitor
		group.Go(func() error {
			if err := h.PG.NewSelect().
				ColumnExpr("count(*) filter (where event.status = ?)", org.AlertStatusOpen).
				ColumnExpr("count(*) filter (where event.status = ?)", org.AlertStatusClosed).
				Model((*org.BaseAlert)(nil)).
				Join("JOIN alert_events AS event ON event.id = a.event_id").
				Where("a.monitor_id = ?", monitor.ID).
				Scan(ctx, &monitor.AlertOpenCount, &monitor.AlertClosedCount); err != nil {
				return err
			}
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
		Model((*org.BaseMonitor)(nil)).
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

func (h *MonitorHandler) ShowYAML(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	monitor := monitorFromContext(ctx)
	base := monitor.Base()

	var out map[string]any
	switch monitor := monitor.(type) {
	case *org.MetricMonitor:
		tpl := metrics.NewMetricMonitorTpl(monitor)
		out = map[string]any{
			"monitors": []*metrics.MetricMonitorTpl{tpl},
		}
	case *org.ErrorMonitor:
		tpl := metrics.NewErrorMonitorTpl(monitor)
		out = map[string]any{
			"monitors": []*metrics.ErrorMonitorTpl{tpl},
		}
	default:
		panic(fmt.Errorf("unsupported monitor type: %T", monitor))
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(out); err != nil {
		return err
	}

	header := w.Header()
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=monitor_%d.yml", base.ID))
	header.Set("Content-Type", "text/yaml")
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (h *MonitorHandler) selectChannelIDs(ctx context.Context, monitorID uint64) ([]uint64, error) {
	ids := make([]uint64, 0)
	if err := h.PG.NewSelect().
		ColumnExpr("channel_id").
		Model((*org.MonitorChannel)(nil)).
		Where("monitor_id = ?", monitorID).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

type MetricMonitorIn struct {
	Name string `json:"name"`

	NotifyEveryoneByEmail bool                    `json:"notifyEveryoneByEmail"`
	Params                org.MetricMonitorParams `json:"params"`

	ChannelIDs []uint64 `json:"channelIds"`
}

func (in *MetricMonitorIn) Validate(
	ctx context.Context, app *bunapp.App, monitor *org.MetricMonitor,
) error {
	monitor.Name = in.Name
	monitor.NotifyEveryoneByEmail = in.NotifyEveryoneByEmail
	monitor.Params = in.Params

	if err := monitor.Validate(); err != nil {
		return err
	}

	for _, ma := range in.Params.Metrics {
		if _, err := metrics.SelectMetricByName(
			ctx, app, monitor.ProjectID, ma.Name,
		); err == sql.ErrNoRows {
			return fmt.Errorf("metric %s does not exist", ma.Name)
		}
	}

	options, err := monitor.MadalarmOptions()
	if err != nil {
		return err
	}

	if _, err := madalarm.Check([]float64{}, options...); err != nil {
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

	monitor := org.NewMetricMonitor()
	monitor.ProjectID = project.ID
	monitor.State = org.MonitorActive
	monitor.Type = org.MonitorMetric

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

	org.CreateAchievementOnce(ctx, h.App, &org.Achievement{
		ProjectID: project.ID,
		Name:      org.AchievCreateMetricMonitor,
	})

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
		Model((*org.MonitorChannel)(nil)).
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

	NotifyEveryoneByEmail bool                   `json:"notifyEveryoneByEmail"`
	Params                org.ErrorMonitorParams `json:"params"`

	ChannelIDs []uint64 `json:"channelIds"`
}

func (in *ErrorMonitorIn) Validate(
	ctx context.Context, app *bunapp.App, monitor *org.ErrorMonitor,
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

	monitor := org.NewErrorMonitor()
	monitor.ProjectID = project.ID
	monitor.State = org.MonitorActive
	monitor.Type = org.MonitorError

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
		Model((*org.MonitorChannel)(nil)).
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
	monitor *org.BaseMonitor,
	channelIDs []uint64,
) error {
	for _, channelID := range channelIDs {
		mc := &org.MonitorChannel{
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

func (h *MonitorHandler) CreateMonitorFromYAML(
	w http.ResponseWriter, req bunrouter.Request,
) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	user := org.UserFromContext(ctx)

	var in struct {
		Monitors []*metrics.MonitorTpl `yaml:"monitors"`
	}
	dec := yaml.NewDecoder(req.Body)
	if err := dec.Decode(&in); err != nil {
		return err
	}

	if len(in.Monitors) == 0 {
		return errors.New("YAML does not contain any monitors")
	}

	monitors := make([]org.Monitor, 0, len(in.Monitors))
	var hasMetricMonitor bool

	for _, monitorTpl := range in.Monitors {
		switch tpl := monitorTpl.Value.(type) {
		case *metrics.MetricMonitorTpl:
			monitor := org.NewMetricMonitor()

			if err := tpl.Populate(monitor); err != nil {
				return err
			}
			monitor.ProjectID = project.ID

			if err := monitor.Validate(); err != nil {
				return err
			}

			monitors = append(monitors, monitor)
			hasMetricMonitor = true
		case *metrics.ErrorMonitorTpl:
			monitor := org.NewErrorMonitor()

			if err := tpl.Populate(monitor); err != nil {
				return err
			}
			monitor.ProjectID = project.ID

			if err := monitor.Validate(); err != nil {
				return err
			}

			monitors = append(monitors, monitor)
		default:
			panic(fmt.Errorf("unsupported monitor type: %T", tpl))
		}
	}

	for _, monitor := range monitors {
		if err := org.InsertMonitor(ctx, h.App.PG, monitor); err != nil {
			return err
		}
	}

	if hasMetricMonitor {
		org.CreateAchievementOnce(ctx, h.App, &org.Achievement{
			UserID:    user.ID,
			ProjectID: project.ID,
			Name:      org.AchievCreateMetricMonitor,
		})
	}

	return httputil.JSON(w, bunrouter.H{
		"monitors": monitors,
	})
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
	return h.updateState(w, req, org.MonitorActive)
}

func (h *MonitorHandler) Pause(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateState(w, req, org.MonitorPaused)
}

func (h *MonitorHandler) updateState(
	w http.ResponseWriter, req bunrouter.Request, state org.MonitorState,
) error {
	ctx := req.Context()

	monitor := monitorFromContext(ctx).Base()
	monitor.State = state

	if _, err := h.PG.NewUpdate().
		Model((*org.BaseMonitor)(nil)).
		Set("state = ?", state).
		Where("id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
