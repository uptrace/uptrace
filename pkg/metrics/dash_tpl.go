package metrics

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"

	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"gopkg.in/yaml.v3"
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
		tpl := new(DashboardTpl)
		if err := yamlUnmarshalDashboardTpl(dec, tpl); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		dashboards = append(dashboards, tpl)
	}

	return dashboards, nil
}

func yamlUnmarshalDashboardTpl(dec *yaml.Decoder, tpl *DashboardTpl) error {
	if err := dec.Decode(&tpl); err != nil {
		return err
	}

	tpl.Grid = make([]GridColumnTpl, len(tpl.GridNodes))
	for i := range tpl.GridNodes {
		var err error
		tpl.Grid[i], err = yamlDecodeGridColumnTpl(&tpl.GridNodes[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func yamlDecodeGridColumnTpl(node *yaml.Node) (any, error) {
	type GridColumnTpl struct {
		Type GridColumnType `yaml:"type"`
	}

	tpl := new(GridColumnTpl)
	if err := node.Decode(tpl); err != nil {
		return nil, err
	}

	switch tpl.Type {
	case "", GridColumnChart:
		tpl := new(ChartGridColumnTpl)
		if err := node.Decode(tpl); err != nil {
			return nil, err
		}
		return tpl, nil
	case GridColumnTable:
		tpl := new(TableGridColumnTpl)
		if err := node.Decode(tpl); err != nil {
			return nil, err
		}
		return tpl, nil
	case GridColumnHeatmap:
		tpl := new(HeatmapGridColumnTpl)
		if err := node.Decode(tpl); err != nil {
			return nil, err
		}
		return tpl, nil
	default:
		return nil, fmt.Errorf("unsupported grid column type: %q", tpl.Type)
	}
}

type DashboardTpl struct {
	Schema string `yaml:"schema"`
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`

	Table struct {
		Gauges  []*DashGaugeTpl          `yaml:"gauges,omitempty"`
		Metrics []string                 `yaml:"metrics"`
		Query   []string                 `yaml:"query"`
		Columns map[string]*MetricColumn `yaml:"columns,omitempty"`
	} `yaml:"table"`

	GridGauges []*DashGaugeTpl `yaml:"grid_gauges,omitempty"`
	GridNodes  []yaml.Node     `yaml:"grid"`
	Grid       []GridColumnTpl `yaml:"-"`
}

func NewDashboardTpl(
	dash *Dashboard, grid []GridColumn, tableGauges, gridGauges []*DashGauge,
) (*DashboardTpl, error) {
	tpl := new(DashboardTpl)
	tpl.Schema = "v1"
	tpl.ID = dash.TemplateID
	tpl.Name = dash.Name

	if tpl.ID == "" {
		tpl.ID = fmt.Sprintf("project_%d.dashboard_%d", dash.ProjectID, dash.ID)
	}

	tpl.Table.Metrics = make([]string, len(dash.TableMetrics))
	for i, metric := range dash.TableMetrics {
		tpl.Table.Metrics[i] = metric.String()
	}
	tpl.Table.Query = upql.SplitQuery(dash.TableQuery)
	tpl.Table.Columns = dash.TableColumnMap

	tpl.GridNodes = make([]yaml.Node, len(grid))
	for i, col := range grid {
		var colTpl any
		switch col := col.(type) {
		case *ChartGridColumn:
			colTpl = NewChartGridColumnTpl(col)
		case *TableGridColumn:
			colTpl = NewTableGridColumnTpl(col)
		case *HeatmapGridColumn:
			colTpl = NewHeatmapGridColumnTpl(col)
		default:
			return nil, fmt.Errorf("unsupported grid column type: %T", col)
		}

		node := &tpl.GridNodes[i]
		if err := node.Encode(colTpl); err != nil {
			return nil, err
		}
	}

	if len(tableGauges) > 0 {
		tpl.Table.Gauges = make([]*DashGaugeTpl, len(tableGauges))
		for i, gauge := range tableGauges {
			tpl.Table.Gauges[i] = NewDashGaugeTpl(gauge)
		}
	}

	if len(gridGauges) > 0 {
		tpl.GridGauges = make([]*DashGaugeTpl, len(gridGauges))
		for i, gauge := range gridGauges {
			tpl.GridGauges[i] = NewDashGaugeTpl(gauge)
		}
	}

	return tpl, nil
}

type DashGaugeTpl struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Template    string `yaml:"template,omitempty"`

	Metrics []string                 `yaml:"metrics"`
	Query   []string                 `yaml:"query"`
	Columns map[string]*MetricColumn `yaml:"columns,omitempty"`

	GridQueryTemplate string `yaml:"grid_query_template,omitempty"`
}

func NewDashGaugeTpl(gauge *DashGauge) *DashGaugeTpl {
	tpl := new(DashGaugeTpl)

	tpl.Name = gauge.Name
	tpl.Description = gauge.Description
	tpl.Template = gauge.Template

	tpl.Metrics = make([]string, len(gauge.Metrics))
	for i, metric := range gauge.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.Query = upql.SplitQuery(gauge.Query)
	tpl.Columns = gauge.ColumnMap

	tpl.GridQueryTemplate = gauge.GridQueryTemplate

	return tpl
}

type GridColumnTpl interface{}

type BaseGridColumnTpl struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`

	Width  int32 `yaml:"width,omitempty"`
	Height int32 `yaml:"height,omitempty"`
	XAxis  int32 `yaml:"xAxis,omitempty"`
	YAxis  int32 `yaml:"yAxis,omitempty"`

	GridQueryTemplate string `yaml:"grid_query_template,omitempty"`

	Type GridColumnType `yaml:"type"`
}

func (c *BaseGridColumnTpl) From(gridColType GridColumnType, baseCol *BaseGridColumn) {
	c.Name = baseCol.Name
	c.Description = baseCol.Description

	c.Width = baseCol.Width
	c.Height = baseCol.Height
	c.XAxis = baseCol.XAxis
	c.YAxis = baseCol.YAxis

	c.GridQueryTemplate = baseCol.GridQueryTemplate

	c.Type = gridColType
}

type ChartGridColumnTpl struct {
	BaseGridColumnTpl `yaml:",inline"`

	ChartKind ChartKind                   `yaml:"chart"`
	Metrics   []string                    `yaml:"metrics"`
	Query     []string                    `yaml:"query"`
	Columns   map[string]*MetricColumn    `yaml:"columns,omitempty"`
	Styles    map[string]*TimeseriesStyle `yaml:"styles,omitempty"`
	Legend    *ChartLegend                `yaml:"legend,omitempty"`
}

func NewChartGridColumnTpl(col *ChartGridColumn) *ChartGridColumnTpl {
	tpl := new(ChartGridColumnTpl)
	tpl.BaseGridColumnTpl.From(GridColumnChart, col.BaseGridColumn)

	tpl.Metrics = make([]string, len(col.Params.Metrics))
	for i, metric := range col.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.ChartKind = col.Params.ChartKind
	tpl.Query = upql.SplitQuery(col.Params.Query)
	tpl.Columns = col.Params.ColumnMap
	tpl.Styles = col.Params.TimeseriesMap
	tpl.Legend = col.Params.Legend
	if tpl.Legend.Type == "" {
		tpl.Legend = nil
	}

	return tpl
}

type TableGridColumnTpl struct {
	BaseGridColumnTpl `yaml:",inline"`

	Metrics []string                 `yaml:"metrics"`
	Query   []string                 `yaml:"query"`
	Columns map[string]*MetricColumn `yaml:"columns,omitempty"`
}

func NewTableGridColumnTpl(col *TableGridColumn) *TableGridColumnTpl {
	tpl := new(TableGridColumnTpl)
	tpl.BaseGridColumnTpl.From(GridColumnTable, col.BaseGridColumn)

	tpl.Metrics = make([]string, len(col.Params.Metrics))
	for i, metric := range col.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.Query = upql.SplitQuery(col.Params.Query)
	tpl.Columns = col.Params.ColumnMap

	return tpl
}

type HeatmapGridColumnTpl struct {
	BaseGridColumnTpl `yaml:",inline"`

	Metric string   `yaml:"metric"`
	Unit   string   `yaml:"unit,omitempty"`
	Query  []string `yaml:"query"`
}

func NewHeatmapGridColumnTpl(col *HeatmapGridColumn) *HeatmapGridColumnTpl {
	tpl := new(HeatmapGridColumnTpl)
	tpl.BaseGridColumnTpl.From(GridColumnHeatmap, col.BaseGridColumn)

	tpl.Metric = col.Params.Metric
	tpl.Unit = col.Params.Unit
	tpl.Query = upql.SplitQuery(col.Params.Query)

	return tpl
}
