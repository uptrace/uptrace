package upql

import (
	"fmt"
	"strings"
)

type Metric struct {
	Name  string `yaml:"name" json:"name"`
	Alias string `yaml:"alias" json:"alias"`
}

func ParseMetrics(ss []string) ([]Metric, error) {
	metrics := make([]Metric, len(ss))
	for i, s := range ss {
		metric, err := parseMetric(s)
		if err != nil {
			return nil, err
		}
		metrics[i] = metric
	}
	return metrics, validateMetrics(metrics)
}

func parseMetric(s string) (Metric, error) {
	for _, sep := range []string{" as ", " AS "} {
		if ss := strings.Split(s, sep); len(ss) == 2 {
			name := strings.TrimSpace(ss[0])
			alias := strings.TrimSpace(ss[1])
			return Metric{
				Name:  name,
				Alias: strings.TrimPrefix(alias, "$"),
			}, nil
		}
	}
	return Metric{}, fmt.Errorf("can't parse metric alias %q", s)
}

func validateMetrics(metrics []Metric) error {
	seen := make(map[string]struct{}, len(metrics))
	for _, metric := range metrics {
		if metric.Name == "" {
			return fmt.Errorf("metric name is empty")
		}
		if metric.Alias == "" {
			return fmt.Errorf("metric alias is empty")
		}
		if _, ok := seen[metric.Alias]; ok {
			return fmt.Errorf("duplicated metric alias %q", metric.Alias)
		}
		seen[metric.Alias] = struct{}{}
	}
	return nil
}
