package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
)

type GridColumn interface {
	Base() *BaseGridColumn
	Validate() error
}

type BaseGridColumn struct {
	bun.BaseModel `bun:"dash_grid_columns,alias:g"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`
	DashID    uint64 `json:"dashId"`

	Name        string `json:"name"`
	Description string `json:"description" bun:",nullzero"`

	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	XAxis  int32 `json:"xAxis"`
	YAxis  int32 `json:"yAxis"`

	GridQueryTemplate string `json:"gridQueryTemplate" bun:",nullzero"`

	Type   GridColumnType `json:"type"`
	Params bunutil.Params `json:"params"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

type GridColumnType string

const (
	GridColumnChart   GridColumnType = "chart"
	GridColumnTable   GridColumnType = "table"
	GridColumnHeatmap GridColumnType = "heatmap"
)

var _ GridColumn = (*BaseGridColumn)(nil)

func (c *BaseGridColumn) Base() *BaseGridColumn {
	return c
}

func (c *BaseGridColumn) FromTemplate(tpl *BaseGridColumnTpl) error {
	c.Name = tpl.Name
	c.Description = tpl.Description

	c.Width = tpl.Width
	c.Height = tpl.Height
	c.XAxis = tpl.XAxis
	c.YAxis = tpl.YAxis

	c.GridQueryTemplate = tpl.GridQueryTemplate

	return nil
}

func (c *BaseGridColumn) Validate() error {
	if c.Name == "" {
		return errors.New("grid column name can't be empty")
	}

	if false {
		if _, err := mql.ParseError(c.GridQueryTemplate); err != nil {
			return fmt.Errorf("can't parse grid query template: %w", err)
		}
	}

	switch c.Type {
	case "":
		return errors.New("grid column type can't be empty")
	case GridColumnChart, GridColumnTable, GridColumnHeatmap:
	default:
		return fmt.Errorf("unsupported grid column type: %q", c.Type)
	}

	if c.CreatedAt.IsZero() {
		now := time.Now()
		c.CreatedAt = now
		c.UpdatedAt = now
	}

	return nil
}

//------------------------------------------------------------------------------

type ChartGridColumn struct {
	*BaseGridColumn `bun:",inherit"`

	Params ChartColumnParams `json:"params"`
}

type ChartColumnParams struct {
	ChartKind     ChartKind                   `json:"chartKind"`
	Metrics       []mql.MetricAlias           `json:"metrics"`
	Query         string                      `json:"query"`
	ColumnMap     map[string]*MetricColumn    `json:"columnMap"`
	TimeseriesMap map[string]*TimeseriesStyle `json:"timeseriesMap"`
	Legend        *ChartLegend                `json:"legend"`
}

type TimeseriesStyle struct {
	Color      string `json:"color" yaml:"color"`
	Opacity    int32  `json:"opacity" yaml:"opacity"`
	LineWidth  int32  `json:"lineWidth" yaml:"line_width"`
	Symbol     string `json:"symbol" yaml:"symbol"`
	SymbolSize int32  `json:"symbolSize" yaml:"symbol_size"`
}

type ChartKind string

const (
	ChartLine        ChartKind = "line"
	ChartArea        ChartKind = "area"
	ChartBar         ChartKind = "bar"
	ChartStackedArea ChartKind = "stacked-area"
	ChartStackedBar  ChartKind = "stacked-bar"
)

type ChartLegend struct {
	Type      LegendType      `json:"type" yaml:"type"`
	Placement LegendPlacement `json:"placement" yaml:"placement"`
	Values    []LegendValue   `json:"values" yaml:"values"`
}

type LegendType string

const (
	LegendNone  LegendType = "none"
	LegendList  LegendType = "list"
	LegendTable LegendType = "table"
)

type LegendPlacement string

const (
	LegendBottom LegendPlacement = "bottom"
	LegendRight  LegendPlacement = "right"
)

type LegendValue string

const (
	LegendAvg LegendValue = "avg"
	LegendMin LegendValue = "min"
	LegendMax LegendValue = "max"
	LegenLast LegendValue = "last"
)

var _ GridColumn = (*ChartGridColumn)(nil)

func (c *ChartGridColumn) Base() *BaseGridColumn {
	return c.BaseGridColumn
}

func (c *ChartGridColumn) FromTemplate(tpl *ChartGridColumnTpl) error {
	if err := c.BaseGridColumn.FromTemplate(&tpl.BaseGridColumnTpl); err != nil {
		return err
	}

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	c.Params.ChartKind = tpl.ChartKind
	c.Params.Metrics = metrics
	c.Params.Query = strings.Join(tpl.Query, " | ")
	c.Params.ColumnMap = tpl.Columns
	c.Params.Legend = tpl.Legend

	if c.Params.ChartKind == "" {
		if len(c.Params.Metrics) == 1 && !strings.Contains(c.Params.Query, "group by") {
			c.Params.ChartKind = ChartArea
		} else {
			c.Params.ChartKind = ChartLine
		}
	}

	return nil
}

func (c *ChartGridColumn) Validate() error {
	if err := c.BaseGridColumn.Validate(); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("grid column %q is invalid: %w", c.Name, err)
	}
	return nil
}

func (c *ChartGridColumn) validate() error {
	if len(c.Params.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(c.Params.Metrics) > 5 {
		return errors.New("you can't use more than 5 metrics in a single visualization")
	}
	for _, metric := range c.Params.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if c.Params.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	if c.Params.ColumnMap == nil {
		c.Params.ColumnMap = make(map[string]*MetricColumn)
	}
	if c.Params.TimeseriesMap == nil {
		c.Params.TimeseriesMap = make(map[string]*TimeseriesStyle)
	}
	return nil
}

//------------------------------------------------------------------------------

type TableGridColumn struct {
	*BaseGridColumn `bun:",inherit"`

	Params TableColumnParams `json:"params"`
}

type TableColumnParams struct {
	Metrics   []mql.MetricAlias        `json:"metrics"`
	Query     string                   `json:"query"`
	ColumnMap map[string]*MetricColumn `json:"columnMap"`
}

var _ GridColumn = (*TableGridColumn)(nil)

func (c *TableGridColumn) Base() *BaseGridColumn {
	return c.BaseGridColumn
}

func (c *TableGridColumn) FromTemplate(tpl *TableGridColumnTpl) error {
	if err := c.BaseGridColumn.FromTemplate(&tpl.BaseGridColumnTpl); err != nil {
		return err
	}

	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	c.Params.Metrics = metrics
	c.Params.Query = strings.Join(tpl.Query, " | ")
	c.Params.ColumnMap = tpl.Columns

	return nil
}

func (c *TableGridColumn) Validate() error {
	if err := c.BaseGridColumn.Validate(); err != nil {
		return err
	}

	if len(c.Params.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(c.Params.Metrics) > 5 {
		return errors.New("you can't use more than 5 metrics in a single visualization")
	}
	for _, metric := range c.Params.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if c.Params.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	if c.Params.ColumnMap == nil {
		c.Params.ColumnMap = make(map[string]*MetricColumn)
	}

	if c.Width == 0 && c.Height == 0 {
		c.Width = 12
		c.Height = 21
	}

	return nil
}

//------------------------------------------------------------------------------

type HeatmapGridColumn struct {
	*BaseGridColumn `bun:",inherit"`

	Params HeatmapColumnParams `json:"params"`
}

type HeatmapColumnParams struct {
	Metric string `json:"metric"`
	Unit   string `json:"unit"`
	Query  string `json:"query"`
}

var _ GridColumn = (*HeatmapGridColumn)(nil)

func (c *HeatmapGridColumn) Base() *BaseGridColumn {
	return c.BaseGridColumn
}

func (c *HeatmapGridColumn) FromTemplate(tpl *HeatmapGridColumnTpl) error {
	if err := c.BaseGridColumn.FromTemplate(&tpl.BaseGridColumnTpl); err != nil {
		return err
	}

	c.Params.Metric = tpl.Metric
	c.Params.Unit = tpl.Unit
	c.Params.Query = strings.Join(tpl.Query, " | ")

	return nil
}

func (c *HeatmapGridColumn) Validate() error {
	if err := c.BaseGridColumn.Validate(); err != nil {
		return err
	}
	if c.Params.Metric == "" {
		return errors.New("metric can't be empty")
	}
	if _, err := mql.ParseError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	if c.Width == 0 && c.Height == 0 {
		c.Width = 12
		c.Height = 18
	}

	return nil
}

//------------------------------------------------------------------------------

func SelectGridColumn(
	ctx context.Context, app *bunapp.App, colID uint64,
) (GridColumn, error) {
	baseCol := new(BaseGridColumn)

	if err := app.PG.NewSelect().
		Model(baseCol).
		Where("id = ?", colID).
		Scan(ctx); err != nil {
		return nil, err
	}

	return decodeBaseGridColumn(baseCol)
}

func decodeBaseGridColumn(baseCol *BaseGridColumn) (GridColumn, error) {
	switch baseCol.Type {
	case GridColumnChart:
		col := &ChartGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := baseCol.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil
	case GridColumnTable:
		col := &TableGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := baseCol.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil
	case GridColumnHeatmap:
		col := &HeatmapGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := baseCol.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil
	default:
		return nil, fmt.Errorf("unknown grid column type: %s", baseCol.Type)
	}
}

func SelectGridColumns(
	ctx context.Context, app *bunapp.App, dashID uint64,
) ([]GridColumn, error) {
	baseColumns, err := SelectBaseGridColumns(ctx, app, dashID)
	if err != nil {
		return nil, err
	}

	columns := make([]GridColumn, len(baseColumns))
	for i, baseCol := range baseColumns {
		columns[i], err = decodeBaseGridColumn(baseCol)
		if err != nil {
			return nil, err
		}
	}
	return columns, nil
}

func SelectBaseGridColumns(
	ctx context.Context, app *bunapp.App, dashID uint64,
) ([]*BaseGridColumn, error) {
	var columns []*BaseGridColumn
	if err := app.PG.NewSelect().
		Model(&columns).
		Where("dash_id = ?", dashID).
		OrderExpr("id ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return columns, nil
}

func InsertGridColumns(
	ctx context.Context, app *bunapp.App, columns []*BaseGridColumn,
) error {
	if _, err := app.PG.NewInsert().
		Model(&columns).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
