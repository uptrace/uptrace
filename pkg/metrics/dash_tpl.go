package metrics

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unixtime"
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

		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, err
		}

		id := strings.TrimSuffix(entry.Name(), ".yml")
		tpl := new(DashboardTpl)

		dec := yaml.NewDecoder(bytes.NewReader(data))
		//dec.KnownFields(false)
		if err := dec.Decode(tpl); err != nil {
			return nil, fmt.Errorf("invalid %q dashboard: %w", id, err)
		}

		tpl.ID = id
		dashboards = append(dashboards, tpl)
	}

	return dashboards, nil
}

type DashboardTpl struct {
	Schema string `yaml:"schema"`

	ID         string          `yaml:"id"`
	Name       string          `yaml:"name"`
	TimeOffset unixtime.Millis `yaml:"time_offset,omitempty"`

	Table struct {
		GridItems []*GridItemTpl          `yaml:"grid_items,omitempty"`
		Metrics   []string                `yaml:"metrics"`
		Query     []string                `yaml:"query"`
		Columns   map[string]*TableColumn `yaml:"columns,omitempty"`
	} `yaml:"table"`

	GridRows []*GridRowTpl       `yaml:"grid_rows"`
	Monitors []*MetricMonitorTpl `yaml:"monitors,omitempty"`
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

	tpl.Table.GridItems = make([]*GridItemTpl, len(tableItems))
	for i, item := range tableItems {
		tpl.Table.GridItems[i] = NewGridItemTpl(item)
	}

	tpl.Table.Metrics = make([]string, len(dash.TableMetrics))
	for i, metric := range dash.TableMetrics {
		tpl.Table.Metrics[i] = metric.String()
	}
	tpl.Table.Query = mql.SplitQuery(dash.TableQuery)
	tpl.Table.Columns = dash.TableColumnMap

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

	metrics, err := parseMetrics(tpl.Table.Metrics)
	if err != nil {
		return err
	}

	dash.TemplateID = tpl.ID
	dash.Name = tpl.Name
	dash.TimeOffset = tpl.TimeOffset
	dash.TableMetrics = metrics
	dash.TableQuery = mql.JoinQuery(tpl.Table.Query)
	dash.TableColumnMap = tpl.Table.Columns

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
	row.Expanded = true
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

	return nil
}

type MonitorTpl struct {
	Value any // *MetricMonitorTpl | *ErrorMonitorTpl
}

func NewMonitorTpl(monitor org.Monitor) *MonitorTpl {
	var tpl any
	switch monitor := monitor.(type) {
	case *org.MetricMonitor:
		tpl = NewMetricMonitorTpl(monitor)
	case *org.ErrorMonitor:
		tpl = NewErrorMonitorTpl(monitor)
	default:
		panic(fmt.Errorf("unsupported grid monitor type: %T", monitor))
	}
	return &MonitorTpl{
		Value: tpl,
	}
}

var _ yaml.Marshaler = (*MonitorTpl)(nil)

func (tpl *MonitorTpl) MarshalYAML() (interface{}, error) {
	return tpl.Value, nil
}

var _ yaml.Unmarshaler = (*MonitorTpl)(nil)

func (tpl *MonitorTpl) UnmarshalYAML(node *yaml.Node) error {
	var in struct {
		Type org.MonitorType `yaml:"type"`
	}

	if err := node.Decode(&in); err != nil {
		return err
	}

	switch in.Type {
	case "", org.MonitorMetric:
		tpl.Value = new(MetricMonitorTpl)
		if err := node.Decode(tpl.Value); err != nil {
			return err
		}
		return nil
	case org.MonitorError:
		tpl.Value = new(ErrorMonitorTpl)
		if err := node.Decode(tpl.Value); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported monitor type: %q", in.Type)
	}
}

type BaseMonitorTpl struct {
	Name string `yaml:"name"`

	Type org.MonitorType `yaml:"type,omitempty"`

	NotifyEveryoneByEmail bool `yaml:"notify_everyone_by_email"`
}

func (tpl *BaseMonitorTpl) initFrom(monitor *org.BaseMonitor) {
	tpl.Name = monitor.Name
	tpl.NotifyEveryoneByEmail = monitor.NotifyEveryoneByEmail
}

func (tpl *BaseMonitorTpl) Populate(monitor *org.BaseMonitor) error {
	monitor.Name = tpl.Name
	monitor.NotifyEveryoneByEmail = tpl.NotifyEveryoneByEmail
	return nil
}

type MetricMonitorTpl struct {
	BaseMonitorTpl `yaml:",inline"`

	Metrics    []string `yaml:"metrics"`
	Query      []string `yaml:"query"`
	Column     string   `yaml:"column"`
	ColumnUnit string   `yaml:"column_unit,omitempty"`

	MinAllowedValue bunutil.NullFloat64 `yaml:"min_allowed_value,omitempty"`
	MaxAllowedValue bunutil.NullFloat64 `yaml:"max_allowed_value,omitempty"`

	CheckNumPoint int             `yaml:"check_num_point"`
	TimeOffset    unixtime.Millis `yaml:"time_offset,omitempty"`
}

func NewMetricMonitorTpl(monitor *org.MetricMonitor) *MetricMonitorTpl {
	tpl := new(MetricMonitorTpl)
	tpl.BaseMonitorTpl.initFrom(monitor.BaseMonitor)
	tpl.BaseMonitorTpl.Type = org.MonitorMetric

	tpl.Metrics = make([]string, len(monitor.Params.Metrics))
	for i, metric := range monitor.Params.Metrics {
		tpl.Metrics[i] = metric.String()
	}

	tpl.Query = mql.SplitQuery(monitor.Params.Query)
	tpl.Column = monitor.Params.Column
	tpl.ColumnUnit = monitor.Params.ColumnUnit

	tpl.MinAllowedValue = monitor.Params.MinAllowedValue
	tpl.MaxAllowedValue = monitor.Params.MaxAllowedValue

	tpl.CheckNumPoint = monitor.Params.CheckNumPoint
	tpl.TimeOffset = monitor.Params.TimeOffset

	return tpl
}

func (tpl *MetricMonitorTpl) Populate(monitor *org.MetricMonitor) error {
	if err := tpl.BaseMonitorTpl.Populate(monitor.BaseMonitor); err != nil {
		return err
	}

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	monitor.State = org.MonitorActive

	monitor.Params.Metrics = metrics
	monitor.Params.Query = mql.JoinQuery(tpl.Query)
	monitor.Params.Column = tpl.Column
	monitor.Params.ColumnUnit = tpl.ColumnUnit

	monitor.Params.CheckNumPoint = tpl.CheckNumPoint
	monitor.Params.TimeOffset = tpl.TimeOffset

	monitor.Params.MinAllowedValue = tpl.MinAllowedValue
	monitor.Params.MaxAllowedValue = tpl.MaxAllowedValue

	return nil
}

type ErrorMonitorTpl struct {
	BaseMonitorTpl `yaml:",inline"`

	NotifyOnNewErrors       bool              `yaml:"notify_on_new_errors"`
	NotifyOnRecurringErrors bool              `yaml:"notify_on_recurring_errors"`
	Matchers                []org.AttrMatcher `yaml:"matchers"`
}

func NewErrorMonitorTpl(monitor *org.ErrorMonitor) *ErrorMonitorTpl {
	tpl := new(ErrorMonitorTpl)
	tpl.BaseMonitorTpl.initFrom(monitor.BaseMonitor)

	tpl.BaseMonitorTpl.Type = org.MonitorError

	tpl.NotifyOnNewErrors = monitor.Params.NotifyOnNewErrors
	tpl.NotifyOnRecurringErrors = monitor.Params.NotifyOnRecurringErrors
	tpl.Matchers = monitor.Params.Matchers

	return tpl
}

func (tpl *ErrorMonitorTpl) Populate(monitor *org.ErrorMonitor) error {
	if err := tpl.BaseMonitorTpl.Populate(monitor.BaseMonitor); err != nil {
		return err
	}

	monitor.Params.NotifyOnNewErrors = tpl.NotifyOnNewErrors
	monitor.Params.NotifyOnRecurringErrors = tpl.NotifyOnRecurringErrors
	monitor.Params.Matchers = tpl.Matchers

	return nil
}
