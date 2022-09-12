package bunconf

import (
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/metrics/alerting"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
)

type AlertRule struct {
	Name        string            `yaml:"name"`
	Metrics     []string          `yaml:"metrics"`
	Query       []string          `yaml:"query"`
	For         time.Duration     `yaml:"for"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Projects    []uint32          `yaml:"projects"`

	metrics []upql.Metric
}

func (r *AlertRule) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if err := r.validate(); err != nil {
		return fmt.Errorf("invalid rule %q: %w", r.Name, err)
	}
	return nil
}

func (r *AlertRule) validate() error {
	if len(r.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}

	metrics, err := upql.ParseMetrics(r.Metrics)
	if err != nil {
		return err
	}
	r.metrics = metrics

	if len(r.Query) == 0 {
		return fmt.Errorf("rule query is required")
	}
	if r.For == 0 {
		return fmt.Errorf("rule duration is required")
	}
	if r.For%time.Minute != 0 {
		return fmt.Errorf("rule duration must be in minutes, got %s", r.For)
	}
	if len(r.Projects) == 0 {
		return fmt.Errorf("at least on project is required")
	}
	return nil
}

func (r *AlertRule) RuleConfig() alerting.RuleConfig {
	return alerting.RuleConfig{
		Name:        r.Name,
		Metrics:     r.metrics,
		Query:       strings.Join(r.Query, " | "),
		For:         r.For,
		Labels:      r.Labels,
		Annotations: r.Annotations,
	}
}
