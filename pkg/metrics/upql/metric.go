package upql

import (
	"fmt"
	"regexp"
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

var aliasRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func parseMetric(s string) (Metric, error) {
	for _, sep := range []string{" as ", " AS "} {
		if ss := strings.Split(s, sep); len(ss) == 2 {
			name := strings.TrimSpace(ss[0])
			alias := strings.TrimSpace(ss[1])

			if !strings.HasPrefix(alias, "$") {
				return Metric{}, fmt.Errorf("alias %q must start with a dollar sign", alias)
			}
			alias = strings.TrimPrefix(alias, "$")

			if !aliasRE.MatchString(alias) {
				return Metric{}, fmt.Errorf("invalid alias: %q", alias)
			}

			return Metric{
				Name:  name,
				Alias: alias,
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
