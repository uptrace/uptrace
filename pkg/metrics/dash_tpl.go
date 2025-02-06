package metrics

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"gopkg.in/yaml.v3"
)

func readDashboardTemplates() ([]*DashboardTpl, error) {
	fsys := uptrace.DashTemplatesFS()

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	var dashboards []*DashboardTpl

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		dash, err := readDashboardTemplate(fsys, entry.Name())
		if err != nil {
			return nil, err
		}

		dashboards = append(dashboards, dash)
	}

	return dashboards, nil
}

func readDashboardTemplate(fsys fs.FS, name string) (*DashboardTpl, error) {
	const ext = ".yml"

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}

	dashID := strings.TrimSuffix(name, ext)
	tpl := new(DashboardTpl)

	dec := yaml.NewDecoder(bytes.NewReader(data))
	//dec.KnownFields(false)
	if err := dec.Decode(tpl); err != nil {
		return nil, fmt.Errorf("invalid %q dashboard: %w", dashID, err)
	}

	for _, fileName := range tpl.IncludeGridRows {
		temp, err := readDashboardTemplate(fsys, fileName+ext)
		if err != nil {
			return nil, err
		}
		tpl.GridRows = append(tpl.GridRows, temp.GridRows...)
	}

	tpl.ID = dashID
	return tpl, nil
}

type DashboardTpl struct {
	Schema string `yaml:"schema"`

	ID   string          `yaml:"id"`
	If   []MetricMatcher `yaml:"if,omitempty"`
	Name string          `yaml:"name"`

	MinInterval time.Duration `yaml:"min_interval,omitempty"`
	TimeOffset  time.Duration `yaml:"time_offset,omitempty"`
	GridQuery   string        `yaml:"grid_query,omitempty"`

	TableGridItems []*GridItemTpl `yaml:"table_grid_items,omitempty"`
	Table          []DashTableTpl `yaml:"table"`

	IncludeGridRows []string      `yaml:"include_grid_rows,omitempty"`
	GridRows        []*GridRowTpl `yaml:"grid_rows"`
}

type DashTableTpl struct {
	If      []MetricMatcher         `yaml:"if,omitempty"`
	Metrics []string                `yaml:"metrics"`
	Query   []string                `yaml:"query"`
	Columns map[string]*TableColumn `yaml:"columns,omitempty"`
}

func (tpl *DashTableTpl) Populate(dash *Dashboard) error {
	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	dash.TableMetrics = metrics
	dash.TableQuery = mql.JoinQuery(tpl.Query)
	dash.TableColumnMap = tpl.Columns

	return nil
}

type MetricMatcher struct {
	Metric          string `yaml:"metric"`
	Instrumentation string `yaml:"instrumentation"`
	State           string `yaml:"state"`
}

func NewDashboardTpl(
	dash *Dashboard, tableItems []GridItem, gridRows []*GridRow,
) (*DashboardTpl, error) {
	tpl := new(DashboardTpl)
	tpl.Schema = "v2"
	tpl.ID = dash.TemplateID
	tpl.Name = dash.Name
	tpl.TimeOffset = dash.TimeOffset

	if tpl.ID == "" {
		tpl.ID = fmt.Sprintf("project_%d.dashboard_%d", dash.ProjectID, dash.ID)
	}

	tpl.TableGridItems = make([]*GridItemTpl, len(tableItems))
	for i, item := range tableItems {
		tpl.TableGridItems[i] = NewGridItemTpl(item)
	}

	tpl.Table = make([]DashTableTpl, 1)
	table := &tpl.Table[0]
	table.Metrics = make([]string, len(dash.TableMetrics))
	for i, metric := range dash.TableMetrics {
		table.Metrics[i] = metric.String()
	}
	table.Query = mql.SplitQuery(dash.TableQuery)
	table.Columns = dash.TableColumnMap

	tpl.GridRows = make([]*GridRowTpl, len(gridRows))
	for i, row := range gridRows {
		tpl.GridRows[i] = NewGridRowTpl(row)
	}

	return tpl, nil
}

func (tpl *DashboardTpl) Populate(dash *Dashboard) error {
	if tpl.Schema != "v2" {
		return fmt.Errorf("unsupported template schema: %q", tpl.Schema)
	}
	if dash.TemplateID != "" && dash.TemplateID != tpl.ID {
		return fmt.Errorf("template id does not match: got %q, has %q", tpl.ID, dash.TemplateID)
	}

	dash.TemplateID = tpl.ID
	dash.Name = tpl.Name

	dash.MinInterval = tpl.MinInterval
	dash.TimeOffset = tpl.TimeOffset
	dash.GridQuery = tpl.GridQuery

	return nil
}

type GridRowTpl struct {
	Title string         `yaml:"title"`
	Items []*GridItemTpl `yaml:"items"`
}

func NewGridRowTpl(row *GridRow) *GridRowTpl {
	tpl := &GridRowTpl{
		Title: row.Title,
	}
	tpl.Items = make([]*GridItemTpl, len(row.Items))
	for i, item := range row.Items {
		tpl.Items[i] = NewGridItemTpl(item)
	}
	return tpl
}

func (tpl *GridRowTpl) Populate(row *GridRow) error {
	row.Title = tpl.Title
	return nil
}

type GridItemTpl struct {
	Value any
}

func NewGridItemTpl(item GridItem) *GridItemTpl {
	var tpl any
	switch item := item.(type) {
	case *ChartGridItem:
		tpl = NewChartGridItemTpl(item)
	case *TableGridItem:
		tpl = NewTableGridItemTpl(item)
	case *HeatmapGridItem:
		tpl = NewHeatmapGridItemTpl(item)
	case *GaugeGridItem:
		tpl = NewGaugeGridItemTpl(item)
	default:
		panic(fmt.Errorf("unsupported grid item type: %T", item))
	}
	return &GridItemTpl{
		Value: tpl,
	}
}

var _ yaml.Marshaler = (*GridItemTpl)(nil)

func (tpl *GridItemTpl) MarshalYAML() (interface{}, error) {
	return tpl.Value, nil
}

var _ yaml.Unmarshaler = (*GridItemTpl)(nil)

func (item *GridItemTpl) UnmarshalYAML(node *yaml.Node) error {
	var in struct {
		Type GridItemType `yaml:"type"`
	}

	if err := node.Decode(&in); err != nil {
		return err
	}

	switch in.Type {
	case "", GridItemChart:
		tpl := new(ChartGridItemTpl)
		if err := node.Decode(tpl); err != nil {
			return err
		}
		item.Value = tpl
		return nil
	case GridItemTable:
		tpl := new(TableGridItemTpl)
		if err := node.Decode(tpl); err != nil {
			return err
		}
		item.Value = tpl
		return nil
	case GridItemHeatmap:
		tpl := new(HeatmapGridItemTpl)
		if err := node.Decode(tpl); err != nil {
			return err
		}
		item.Value = tpl
		return nil
	case GridItemGauge:
		tpl := new(GaugeGridItemTpl)
		if err := node.Decode(tpl); err != nil {
			return err
		}
		item.Value = tpl
		return nil
	default:
		return fmt.Errorf("unsupported grid item type: %q", in.Type)
	}
}

type BaseGridItemTpl struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`

	Width  int `yaml:"width,omitempty"`
	Height int `yaml:"height,omitempty"`
	XAxis  int `yaml:"x_axis,omitempty"`
	YAxis  int `yaml:"y_axis,omitempty"`

	Type GridItemType `yaml:"type"`
}

func (c *BaseGridItemTpl) From(gridItemType GridItemType, baseItem *BaseGridItem) {
	c.Title = baseItem.Title
	c.Description = baseItem.Description

	c.Width = baseItem.Width
	c.Height = baseItem.Height
	c.XAxis = baseItem.XAxis
	c.YAxis = baseItem.YAxis

	c.Type = gridItemType
}

func (tpl *BaseGridItemTpl) Populate(item *BaseGridItem) {
	item.Title = tpl.Title
	item.Description = tpl.Description

	item.Width = tpl.Width
	item.Height = tpl.Height
	item.XAxis = tpl.XAxis
	item.YAxis = tpl.YAxis
}

type ChartGridItemTpl struct {
	BaseGridItemTpl `yaml:",inline"`

	ChartKind ChartKind                   `yaml:"chart"`
	Metrics   []string                    `yaml:"metrics"`
	Query     []string                    `yaml:"query"`
	Columns   map[string]*MetricColumn    `yaml:"columns,omitempty"`
	Styles    map[string]*TimeseriesStyle `yaml:"styles,omitempty"`
	Legend    *ChartLegend                `yaml:"legend,omitempty"`
}

func NewChartGridItemTpl(item *ChartGridItem) *ChartGridItemTpl {
	tpl := new(ChartGridItemTpl)
	tpl.BaseGridItemTpl.From(GridItemChart, item.BaseGridItem)

	tpl.Metrics = make([]string, len(item.Params.Metrics))
	for i, metric := range item.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.ChartKind = item.Params.ChartKind
	tpl.Query = mql.SplitQuery(item.Params.Query)
	tpl.Columns = item.Params.ColumnMap
	tpl.Styles = item.Params.TimeseriesMap
	tpl.Legend = item.Params.Legend
	if tpl.Legend != nil && tpl.Legend.Type == "" {
		tpl.Legend = nil
	}

	return tpl
}

func (tpl *ChartGridItemTpl) Populate(item *ChartGridItem) error {
	tpl.BaseGridItemTpl.Populate(item.BaseGridItem)

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	item.Params.ChartKind = tpl.ChartKind
	item.Params.Metrics = metrics
	item.Params.Query = mql.JoinQuery(tpl.Query)
	item.Params.ColumnMap = tpl.Columns
	item.Params.Legend = tpl.Legend

	if item.Params.ChartKind == "" {
		if len(item.Params.Metrics) == 1 &&
			!strings.Contains(item.Params.Query, "group by") &&
			!strings.Contains(item.Params.Query, " | ") {
			item.Params.ChartKind = ChartArea
		} else {
			item.Params.ChartKind = ChartLine
		}
	}

	return nil
}

type TableGridItemTpl struct {
	BaseGridItemTpl `yaml:",inline"`

	Metrics []string                 `yaml:"metrics"`
	Query   []string                 `yaml:"query"`
	Columns map[string]*MetricColumn `yaml:"columns,omitempty"`
}

func NewTableGridItemTpl(item *TableGridItem) *TableGridItemTpl {
	tpl := new(TableGridItemTpl)
	tpl.BaseGridItemTpl.From(GridItemTable, item.BaseGridItem)

	tpl.Metrics = make([]string, len(item.Params.Metrics))
	for i, metric := range item.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.Query = mql.SplitQuery(item.Params.Query)
	tpl.Columns = item.Params.ColumnMap

	return tpl
}

func (tpl *TableGridItemTpl) Populate(item *TableGridItem) error {
	tpl.BaseGridItemTpl.Populate(item.BaseGridItem)

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	item.Params.Metrics = metrics
	item.Params.Query = mql.JoinQuery(tpl.Query)
	item.Params.ColumnMap = tpl.Columns

	return nil
}

type HeatmapGridItemTpl struct {
	BaseGridItemTpl `yaml:",inline"`

	Metric string   `yaml:"metric"`
	Unit   string   `yaml:"unit,omitempty"`
	Query  []string `yaml:"query"`
}

func NewHeatmapGridItemTpl(col *HeatmapGridItem) *HeatmapGridItemTpl {
	tpl := new(HeatmapGridItemTpl)
	tpl.BaseGridItemTpl.From(GridItemHeatmap, col.BaseGridItem)

	tpl.Metric = col.Params.Metric
	tpl.Unit = col.Params.Unit
	tpl.Query = mql.SplitQuery(col.Params.Query)

	return tpl
}

func (tpl *HeatmapGridItemTpl) Populate(item *HeatmapGridItem) error {
	tpl.BaseGridItemTpl.Populate(item.BaseGridItem)

	item.Params.Metric = tpl.Metric
	item.Params.Unit = tpl.Unit
	item.Params.Query = mql.JoinQuery(tpl.Query)

	return nil
}

type GaugeGridItemTpl struct {
	BaseGridItemTpl `yaml:",inline"`

	Metrics []string                `yaml:"metrics"`
	Query   []string                `yaml:"query"`
	Columns map[string]*GaugeColumn `yaml:"columns,omitempty"`

	Template      string         `yaml:"template,omitempty"`
	ValueMappings []ValueMapping `yaml:"value_mappings,omitempty"`
}

func NewGaugeGridItemTpl(item *GaugeGridItem) *GaugeGridItemTpl {
	tpl := new(GaugeGridItemTpl)
	tpl.BaseGridItemTpl.From(GridItemGauge, item.BaseGridItem)

	tpl.Metrics = make([]string, len(item.Params.Metrics))
	for i, metric := range item.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.Query = mql.SplitQuery(item.Params.Query)
	tpl.Columns = item.Params.ColumnMap

	tpl.Template = item.Params.Template
	tpl.ValueMappings = item.Params.ValueMappings

	return tpl
}

func (tpl *GaugeGridItemTpl) Populate(item *GaugeGridItem) error {
	tpl.BaseGridItemTpl.Populate(item.BaseGridItem)

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	item.Params.Metrics = metrics
	item.Params.Query = mql.JoinQuery(tpl.Query)
	item.Params.ColumnMap = tpl.Columns

	item.Params.Template = tpl.Template
	item.Params.ValueMappings = tpl.ValueMappings

	return nil
}
