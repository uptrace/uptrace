package metrics

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"

	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"gopkg.in/yaml.v2"
)

func readDashboardTemplates() ([]*DashboardTpl, error) {
	fsys := uptrace.DashTemplatesFS()

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	var dashboards []*DashboardTpl

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}

		got, err := parseDashboards(data)
		if err != nil {
			return nil, err
		}

		dashboards = append(dashboards, got...)
	}

	return dashboards, nil
}

func parseDashboards(data []byte) ([]*DashboardTpl, error) {
	var dashboards []*DashboardTpl

	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		dashboard := new(DashboardTpl)
		if err := dec.Decode(&dashboard); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if err := dashboard.Validate(); err != nil {
			return nil, err
		}

		dashboards = append(dashboards, dashboard)
	}

	return dashboards, nil
}

type DashboardTpl struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`

	Table struct {
		Metrics []string                 `yaml:"metrics"`
		Query   []string                 `yaml:"query"`
		Columns map[string]*MetricColumn `yaml:"columns"`
		Gauges  []*DashGaugeTpl          `yaml:"gauges"`
	} `yaml:"table"`

	Gauges  []*DashGaugeTpl `yaml:"gauges"`
	Entries []*DashEntryTpl `yaml:"entries"`
}

func (d *DashboardTpl) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("template id is required")
	}
	if err := d.validate(); err != nil {
		return fmt.Errorf("%s: %w", d.ID, err)
	}
	return nil
}

func (d *DashboardTpl) validate() error {
	if d.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}
	if len(d.Table.Query) == 0 && len(d.Entries) == 0 {
		return fmt.Errorf("either dashboard query or an entry is required")
	}

	if _, err := upql.ParseMetrics(d.Table.Metrics); err != nil {
		return err
	}

	for _, gauge := range d.Table.Gauges {
		if err := gauge.Validate(); err != nil {
			return err
		}
	}
	for _, gauge := range d.Gauges {
		if err := gauge.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range d.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type DashGaugeTpl struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Template    string                   `yaml:"template"`
	Metrics     []string                 `yaml:"metrics"`
	Query       []string                 `yaml:"query"`
	Columns     map[string]*MetricColumn `yaml:"columns"`
}

func (g *DashGaugeTpl) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("gauge name is required")
	}
	if g.Description == "" {
		return fmt.Errorf("gauge description is required")
	}
	if len(g.Metrics) == 0 {
		return fmt.Errorf("gauge requires at least one metric")
	}
	if len(g.Query) == 0 {
		return fmt.Errorf("gauge query is required")
	}

	if _, err := upql.ParseMetrics(g.Metrics); err != nil {
		return err
	}

	return nil
}

type DashEntryTpl struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	ChartType   string                   `yaml:"chart_type"`
	Metrics     []string                 `yaml:"metrics"`
	Query       []string                 `yaml:"query"`
	Columns     map[string]*MetricColumn `yaml:"columns"`
}

const (
	LineChartType        = "line"
	AreaChartType        = "area"
	BarChartType         = "bar"
	StackedAreaChartType = "stacked-area"
	StackedBarChartType  = "stacked-bar"
)

func (e *DashEntryTpl) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("entry name is required")
	}

	switch e.ChartType {
	case "", LineChartType, AreaChartType, BarChartType, StackedAreaChartType, StackedBarChartType:
	default:
		return fmt.Errorf("unknown chart type: %q", e.ChartType)
	}

	if len(e.Metrics) == 0 {
		return fmt.Errorf("entry requires at least one metric")
	}
	if len(e.Query) == 0 {
		return fmt.Errorf("entry query is required")
	}

	if _, err := upql.ParseMetrics(e.Metrics); err != nil {
		return err
	}

	return nil
}
