package alerting

import (
	"context"
	"time"

	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Manager struct {
	conf  *ManagerConfig
	rules []*Rule

	stop func()
	exit <-chan struct{}
	done <-chan struct{}
}

type ManagerConfig struct {
	Engine   Engine
	Rules    []RuleConfig
	AlertMan AlertManager
	Logger   *zap.Logger
}

type AlertManager interface {
	SendAlerts(ctx context.Context, rule *RuleConfig, alerts []Alert) error
	SaveAlerts(ctx context.Context, rule *RuleConfig, alerts []Alert) error
	LoadAlerts(ctx context.Context, rule *RuleConfig) ([]Alert, error)
}

type Engine interface {
	Eval(
		ctx context.Context,
		projectIDs []uint32,
		metrics []upql.Metric,
		expr string,
		gte, lt time.Time,
	) ([]upql.Timeseries, map[string][]upql.Timeseries, error)
}

func NewManager(conf *ManagerConfig) *Manager {
	rules := make([]*Rule, len(conf.Rules))

	for i := range rules {
		ruleConf := &conf.Rules[i]

		alerts, err := conf.AlertMan.LoadAlerts(context.TODO(), ruleConf)
		if err != nil {
			conf.Logger.Error("LoadAlerts failed", zap.Error(err))
		}

		rules[i] = NewRule(ruleConf, alerts)
	}

	return &Manager{
		conf:  conf,
		rules: rules,
	}
}

func (m *Manager) Stop() {
	m.stop()
	<-m.done
}

func (m *Manager) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	m.stop = cancel
	m.exit = ctx.Done()

	done := make(chan struct{})
	m.done = done
	defer func() {
		close(done)
	}()

	nextCheck := time.Now().
		Add(time.Minute).
		Truncate(time.Minute).
		Add(30 * time.Second)
	for {
		select {
		case <-time.After(time.Until(nextCheck)):
		case <-m.exit:
			return
		}

		m.tick(nextCheck.Truncate(time.Minute))
		nextCheck = nextCheck.Add(time.Minute)
	}
}

func (m *Manager) tick(tm time.Time) {
	ctx := context.Background()
	for _, rule := range m.rules {
		ruleConf := rule.Config()
		if err := bunotel.RunWithNewRoot(ctx, "eval-rule", func(ctx context.Context) error {
			span := trace.SpanFromContext(ctx)

			span.SetAttributes(
				attribute.String("rule.name", ruleConf.Name),
				attribute.String("rule.query", ruleConf.Query),
			)

			return m.evalRule(ctx, rule, tm)
		}); err != nil {
			m.conf.Logger.Error("eval-rule failed",
				zap.Error(err),
				zap.String("rule.name", ruleConf.Name),
				zap.String("rule.query", ruleConf.Query))
		}
	}
}

func (m *Manager) evalRule(ctx context.Context, rule *Rule, tm time.Time) error {
	alerts, err := rule.Eval(ctx, m.conf.Engine, tm)
	if err != nil {
		return err
	}

	if len(alerts) > 0 {
		if err := m.conf.AlertMan.SendAlerts(ctx, rule.Config(), alerts); err != nil {
			return err
		}

		for i := range alerts {
			alerts[i].LastSentAt = tm
		}
	}

	if err := m.conf.AlertMan.SaveAlerts(ctx, rule.Config(), rule.Alerts()); err != nil {
		return err
	}

	return nil
}
