package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	app       *bunapp.App
	projectID uint32
}

var _ alerting.Engine = (*AlertingEngine)(nil)

func NewAlertingEngine(app *bunapp.App, projectID uint32) *AlertingEngine {
	return &AlertingEngine{
		app:       app,
		projectID: projectID,
	}
}

func (e *AlertingEngine) Eval(
	ctx context.Context, metrics []upql.Metric, expr string, gte, lt time.Time,
) ([]upql.Timeseries, error) {
	metricMap := make(map[string]*Metric, len(metrics))
	for _, m := range metrics {
		metric, err := SelectMetricByName(ctx, e.app, e.projectID, m.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("metric %q not found", m.Name)
			}
			return nil, err
		}
		metricMap[m.Alias] = metric
	}

	storage := NewCHStorage(ctx, e.app.CH, &CHStorageConfig{
		ProjectID: e.projectID,
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
	timeseries := engine.Run(parts)

	for _, part := range parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}

	return timeseries, nil
}

//------------------------------------------------------------------------------

type AlertManager struct {
	db        *bun.DB
	notifier  *bunapp.Notifier
	projectID uint32
	logger    *otelzap.Logger
}

var _ alerting.AlertManager = (*AlertManager)(nil)

func NewAlertManager(
	db *bun.DB,
	notifier *bunapp.Notifier,
	projectID uint32,
	logger *otelzap.Logger,
) *AlertManager {
	return &AlertManager{
		db:        db,
		notifier:  notifier,
		projectID: projectID,
		logger:    logger,
	}
}

func (m *AlertManager) SendAlerts(
	ctx context.Context, rule *alerting.RuleConfig, alerts []alerting.Alert,
) error {
	postableAlerts := make(models.PostableAlerts, 0, len(alerts))

	for i := range alerts {
		alert := &alerts[i]

		if alert.State == alerting.StateFiring {
			m.logger.Info("alerting rule is firing",
				zap.String("name", rule.Name),
				zap.String("attrs", alert.Attrs.String()),
				zap.Uint32("project_id", m.projectID))
		} else {
			m.logger.Info("alerting rule is resolved",
				zap.String("name", rule.Name),
				zap.String("attrs", alert.Attrs.String()),
				zap.Uint32("project_id", m.projectID))
		}

		postableAlert := m.convert(rule, alert)
		postableAlerts = append(postableAlerts, postableAlert)
	}

	return m.notifier.Send(ctx, postableAlerts)
}

func (m *AlertManager) convert(
	rule *alerting.RuleConfig, alert *alerting.Alert,
) *models.PostableAlert {
	labels := make(models.LabelSet)
	for _, kv := range alert.Attrs {
		labels[cleanLabelName(kv.Key)] = kv.Value
	}
	for k, v := range rule.Labels {
		labels[cleanLabelName(k)] = v
	}

	labels["alertname"] = rule.Name
	labels["project_id"] = fmt.Sprint(m.projectID)

	annotations := make(models.LabelSet)
	annotations["rule_query"] = rule.Query
	for k, v := range rule.Annotations {
		annotations[cleanLabelName(k)] = v
	}

	return &models.PostableAlert{
		Alert: models.Alert{
			Labels: labels,
		},
		Annotations: annotations,
		StartsAt:    strfmt.DateTime(alert.FiredAt),
		EndsAt:      strfmt.DateTime(alert.ResolvedAt),
	}
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
