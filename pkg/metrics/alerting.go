package metrics

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/metrics/alerting"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"go.uber.org/zap"
)

type AlertingEngine struct {
	app *bunapp.App
}

var _ alerting.Engine = (*AlertingEngine)(nil)

func NewAlertingEngine(app *bunapp.App) *AlertingEngine {
	return &AlertingEngine{
		app: app,
	}
}

func (e *AlertingEngine) Eval(
	ctx context.Context,
	projects []uint32,
	metrics []upql.Metric,
	expr string,
	gte, lt time.Time,
) ([]upql.Timeseries, map[string][]upql.Timeseries, error) {
	metricMap := make(map[string]*Metric, len(metrics))

	var projectID uint32
	if len(projects) == 1 {
		projectID = projects[0]
	}

	for _, m := range metrics {
		metric, err := SelectMetricByName(ctx, e.app, projectID, m.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil, fmt.Errorf("metric %q not found", m.Name)
			}
			return nil, nil, err
		}
		metricMap[m.Alias] = metric
	}

	storage := NewCHStorage(ctx, e.app.CH, &CHStorageConfig{
		Projects: projects,
		TimeFilter: org.TimeFilter{
			TimeGTE: gte,
			TimeLT:  lt,
		},
		MetricMap: metricMap,

		TableName:      e.app.DistTable("measure_minutes_buffer"),
		GroupingPeriod: time.Minute,

		GroupByTime: true,
		FillHoles:   true,
	})
	engine := upql.NewEngine(storage)

	parts := upql.Parse(expr)
	result := engine.Run(parts)

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			return nil, nil, part.Error.Wrapped
		}
	}

	return result.Timeseries, result.Vars, nil
}

//------------------------------------------------------------------------------

type AlertManager struct {
	db       *bun.DB
	notifier *bunapp.Notifier
	logger   *otelzap.Logger
}

var _ alerting.AlertManager = (*AlertManager)(nil)

func NewAlertManager(
	db *bun.DB,
	notifier *bunapp.Notifier,
	logger *otelzap.Logger,
) *AlertManager {
	return &AlertManager{
		db:       db,
		notifier: notifier,
		logger:   logger,
	}
}

func (m *AlertManager) SendAlerts(
	ctx context.Context, rule *alerting.RuleConfig, alerts []alerting.Alert,
) error {
	postableAlerts := make(models.PostableAlerts, 0, len(alerts))

	for i := range alerts {
		alert := &alerts[i]

		postableAlert, err := m.convert(rule, alert)
		if err != nil {
			m.logger.Error("can't create a postable alert", zap.Error(err))
			continue
		}

		postableAlerts = append(postableAlerts, postableAlert)

		fields := []zap.Field{
			zap.String("name", rule.Name),
			zap.Uint32("project_id", alert.ProjectID),
		}
		for _, attr := range alert.Attrs {
			fields = append(fields, zap.String(attr.Key, attr.Value))
		}
		for key, value := range alert.Annotations {
			fields = append(fields, zap.String(key, value))
		}
		for metric, ts := range alert.Metrics {
			fields = append(fields, zap.Float64(metric, ts.Value[len(ts.Value)-1]))
		}

		if alert.State == alerting.StateFiring {
			m.logger.Info("alerting rule is firing", fields...)
		} else {
			m.logger.Info("alerting rule is resolved", fields...)
		}

	}

	return m.notifier.Send(ctx, postableAlerts)
}

type TemplateData struct {
	Labels      map[string]string
	Annotations map[string]string
	Values      map[string]float64
	Query       string
}

var templateDefs = []string{
	"{{$labels := $.Labels}}",
	"{{$annotations := $.Annotations}}",
	"{{$values := $.Values}}",
}

func (m *AlertManager) convert(
	rule *alerting.RuleConfig, alert *alerting.Alert,
) (*models.PostableAlert, error) {
	labels := make(models.LabelSet)
	for _, kv := range alert.Attrs {
		labels[cleanLabelName(kv.Key)] = kv.Value
	}
	for k, v := range rule.Labels {
		labels[cleanLabelName(k)] = v
	}

	values := make(map[string]float64, len(alert.Metrics))
	for metric, ts := range alert.Metrics {
		lastValue := ts.Value[len(ts.Value)-1]
		values[cleanLabelName(metric)] = lastValue
	}

	labels["alertname"] = rule.Name
	labels["uptrace_project_id"] = fmt.Sprint(alert.ProjectID)

	annotations := make(models.LabelSet)
	for k, v := range alert.Annotations {
		annotations[cleanLabelName(k)] = v
	}

	tplData := &TemplateData{
		Labels:      labels,
		Annotations: annotations,
		Values:      values,
		Query:       rule.Query,
	}
	for k, v := range rule.Annotations {
		tpl := append(templateDefs, v)
		t, err := template.New("").Funcs(sprig.FuncMap()).Parse(strings.Join(tpl, ""))
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, tplData); err != nil {
			return nil, err
		}

		annotations[cleanLabelName(k)] = buf.String()
	}

	return &models.PostableAlert{
		Alert: models.Alert{
			Labels: labels,
		},
		Annotations: annotations,
		StartsAt:    strfmt.DateTime(alert.FiredAt),
		EndsAt:      strfmt.DateTime(alert.ResolvedAt),
	}, nil
}

type RuleAlerts struct {
	bun.BaseModel `bun:"rule_alerts,alias:m"`

	RuleID int64            `bun:",pk"`
	Alerts []alerting.Alert `bun:"type:json"`
}

func (m *AlertManager) SaveAlerts(
	ctx context.Context, rule *alerting.RuleConfig, alerts []alerting.Alert,
) error {
	model := &RuleAlerts{
		RuleID: rule.ID(),
		Alerts: alerts,
	}
	_, err := m.db.NewInsert().
		Model(model).
		On("CONFLICT (rule_id) DO UPDATE").
		Set("alerts = EXCLUDED.alerts").
		Exec(ctx)
	return err
}

func (m *AlertManager) LoadAlerts(
	ctx context.Context, rule *alerting.RuleConfig,
) ([]alerting.Alert, error) {
	model := new(RuleAlerts)
	if err := m.db.NewSelect().
		Model(model).
		Where("rule_id = ?", rule.ID()).
		Limit(1).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return model.Alerts, nil
}

func cleanLabelName(s string) string {
	if isValidLabelName(s) {
		return s
	}

	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowedLabelNameChar(c) {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return unsafeconv.String(r)
}

func isValidLabelName(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowedLabelNameChar(c) {
			return false
		}
	}
	return true
}

func isAllowedLabelNameChar(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_'
}
