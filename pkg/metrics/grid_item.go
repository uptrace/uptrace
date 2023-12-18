package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
)

type GridItem interface {
	Base() *BaseGridItem
	Validate() error
	Metrics() []string
}

var (
	_ GridItem = (*ChartGridItem)(nil)
	_ GridItem = (*TableGridItem)(nil)
	_ GridItem = (*HeatmapGridItem)(nil)
	_ GridItem = (*GaugeGridItem)(nil)
)

type BaseGridItem struct {
	bun.BaseModel `bun:"grid_items,alias:g"`

	ID       uint64   `json:"id" bun:",pk,autoincrement"`
	DashID   uint64   `json:"dashId"`
	DashKind DashKind `json:"dashKind"`
	RowID    uint64   `json:"rowId" bun:",nullzero"`

	Title       string `json:"title"`
	Description string `json:"description" bun:",nullzero"`

	Width  int `json:"width"`
	Height int `json:"height"`
	XAxis  int `json:"xAxis"`
	YAxis  int `json:"yAxis"`

	Type   GridItemType   `json:"type"`
	Params bunutil.Params `json:"params"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

type GridItemType string

const (
	GridItemChart   GridItemType = "chart"
	GridItemTable   GridItemType = "table"
	GridItemHeatmap GridItemType = "heatmap"
	GridItemGauge   GridItemType = "gauge"
)

func (item *BaseGridItem) Validate() error {
	if item.Title == "" {
		return errors.New("grid item title can't be empty")
	}
	if item.DashKind == "" {
		return errors.New("grid item dash kind can't be empty")
	}
	if item.Type == "" {
		return errors.New("grid item type can't be empty")
	}

	if item.DashKind == DashKindGrid {
		if item.RowID == 0 {
			return errors.New("grid item row id can't be zero")
		}
	}

	switch item.Type {
	case GridItemHeatmap:
		item.Width = 12
		item.Height = 40
	case GridItemGauge:
		if item.Width == 0 {
			item.Width = 2
		}
		if item.Height == 0 {
			item.Height = 10
		}
	default:
		if item.Width == 0 {
			item.Width = 6
		}
		if item.Height == 0 {
			item.Height = 28
		}
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = time.Now().Add(time.Second)
	}

	return nil
}

//------------------------------------------------------------------------------

type ChartGridItem struct {
	*BaseGridItem `bun:",inherit"`

	Params ChartGridItemParams `json:"params"`
}

func NewChartGridItem() *ChartGridItem {
	gridItem := &ChartGridItem{
		BaseGridItem: &BaseGridItem{
			Type: GridItemChart,
		},
	}
	gridItem.BaseGridItem.Params.Any = &gridItem.Params
	return gridItem
}

type ChartGridItemParams struct {
	ChartKind ChartKind         `json:"chartKind"`
	Metrics   []mql.MetricAlias `json:"metrics"`
	Query     string            `json:"query"`
	// Use ColumnMap instead of MetricMap for compatibility with TableGridItem.
	ColumnMap     map[string]*MetricColumn    `json:"columnMap"`
	TimeseriesMap map[string]*TimeseriesStyle `json:"timeseriesMap"`
	Legend        *ChartLegend                `json:"legend"`
}

type TimeseriesStyle struct {
	Color      string  `json:"color" yaml:"color"`
	Opacity    int32   `json:"opacity" yaml:"opacity"`
	LineWidth  float32 `json:"lineWidth" yaml:"line_width"`
	Symbol     string  `json:"symbol" yaml:"symbol"`
	SymbolSize int32   `json:"symbolSize" yaml:"symbol_size"`
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

func (c *ChartGridItem) Base() *BaseGridItem {
	return c.BaseGridItem
}

func (c *ChartGridItem) Validate() error {
	if err := c.BaseGridItem.Validate(); err != nil {
		return err
	}
	if len(c.Params.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(c.Params.Metrics) > 6 {
		return errors.New("at most 6 metrics are allowed")
	}
	for _, metric := range c.Params.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if c.Params.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseQueryError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	for _, col := range c.Params.ColumnMap {
		if err := col.Validate(); err != nil {
			return err
		}
	}
	if c.Params.ColumnMap == nil {
		c.Params.ColumnMap = make(map[string]*MetricColumn)
	}

	if c.Params.TimeseriesMap == nil {
		c.Params.TimeseriesMap = make(map[string]*TimeseriesStyle)
	}

	return nil
}

func (c *ChartGridItem) Metrics() []string {
	var metrics []string
	for _, metric := range c.Params.Metrics {
		metrics = append(metrics, metric.Name)
	}
	return metrics
}

//------------------------------------------------------------------------------

type TableGridItem struct {
	*BaseGridItem `bun:",inherit"`

	Params TableGridItemParams `json:"params"`
}

func NewTableGridItem() *TableGridItem {
	gridItem := &TableGridItem{
		BaseGridItem: &BaseGridItem{
			Type: GridItemTable,
		},
	}
	gridItem.BaseGridItem.Params.Any = &gridItem.Params
	return gridItem
}

type TableGridItemParams struct {
	Metrics      []mql.MetricAlias        `json:"metrics"`
	Query        string                   `json:"query"`
	ColumnMap    map[string]*MetricColumn `json:"columnMap"`
	ItemsPerPage int                      `json:"itemsPerPage"`
	DenseTable   bool                     `json:"denseTable"`
}

func (c *TableGridItem) Base() *BaseGridItem {
	return c.BaseGridItem
}

func (item *TableGridItem) Validate() error {
	if err := item.BaseGridItem.Validate(); err != nil {
		return err
	}

	if len(item.Params.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(item.Params.Metrics) > 6 {
		return errors.New("at most 6 metrics are allowed")
	}
	for _, metric := range item.Params.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if item.Params.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseQueryError(item.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	for _, col := range item.Params.ColumnMap {
		if err := col.Validate(); err != nil {
			return err
		}
	}
	if item.Params.ColumnMap == nil {
		item.Params.ColumnMap = make(map[string]*MetricColumn)
	}
	if item.Params.ItemsPerPage == 0 {
		item.Params.ItemsPerPage = 5
	}

	return nil
}

func (item *TableGridItem) Metrics() []string {
	var metrics []string
	for _, metric := range item.Params.Metrics {
		metrics = append(metrics, metric.Name)
	}
	return metrics
}

//------------------------------------------------------------------------------

type HeatmapGridItem struct {
	*BaseGridItem `bun:",inherit"`

	Params HeatmapGridItemParams `json:"params"`
}

func NewHeatmapGridItem() *HeatmapGridItem {
	gridItem := &HeatmapGridItem{
		BaseGridItem: &BaseGridItem{
			Type: GridItemHeatmap,
		},
	}
	gridItem.BaseGridItem.Params.Any = &gridItem.Params
	return gridItem
}

type HeatmapGridItemParams struct {
	Metric string `json:"metric"`
	Unit   string `json:"unit"`
	Query  string `json:"query"`
}

func (c *HeatmapGridItem) Base() *BaseGridItem {
	return c.BaseGridItem
}

func (c *HeatmapGridItem) Validate() error {
	if err := c.BaseGridItem.Validate(); err != nil {
		return err
	}
	if c.Params.Metric == "" {
		return errors.New("metric can't be empty")
	}
	if _, err := mql.ParseQueryError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	if c.Width == 0 && c.Height == 0 {
		c.Width = 12
		c.Height = 18
	}
	c.Params.Unit = bunconv.NormUnit(c.Params.Unit)

	return nil
}

func (c *HeatmapGridItem) Metrics() []string {
	return []string{c.Params.Metric}
}

//------------------------------------------------------------------------------

type GaugeGridItem struct {
	*BaseGridItem `bun:",inherit"`

	Params GaugeGridItemParams `json:"params"`
}

func NewGaugeGridItem() *GaugeGridItem {
	gridItem := &GaugeGridItem{
		BaseGridItem: &BaseGridItem{
			Type: GridItemGauge,
		},
	}
	gridItem.BaseGridItem.Params.Any = &gridItem.Params
	return gridItem
}

func (c *GaugeGridItem) Base() *BaseGridItem {
	return c.BaseGridItem
}

func (c *GaugeGridItem) Validate() error {
	if err := c.BaseGridItem.Validate(); err != nil {
		return err
	}

	if len(c.Params.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(c.Params.Metrics) > 6 {
		return errors.New("at most 6 metrics are allowed")
	}
	for _, metric := range c.Params.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if c.Params.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseQueryError(c.Params.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	for _, col := range c.Params.ColumnMap {
		if err := col.Validate(); err != nil {
			return err
		}
	}
	if c.Params.ColumnMap == nil {
		c.Params.ColumnMap = make(map[string]*MetricColumn)
	}

	return nil
}

func (c *GaugeGridItem) Metrics() []string {
	var metrics []string
	for _, metric := range c.Params.Metrics {
		metrics = append(metrics, metric.Name)
	}
	return metrics
}

type GaugeGridItemParams struct {
	Metrics   []mql.MetricAlias        `json:"metrics"`
	Query     string                   `json:"query"`
	ColumnMap map[string]*MetricColumn `json:"columnMap" bun:",nullzero"`

	Template      string         `json:"template" bun:",nullzero"`
	ValueMappings []ValueMapping `json:"valueMappings" bun:",nullzero"`
}

type ValueMapping struct {
	Op    MappingOp   `json:"op" yaml:"op"`
	Value json.Number `json:"value" yaml:"value"`
	Text  string      `json:"text" yaml:"text"`
	Color string      `json:"color" yaml:"color"`
}

type MappingOp string

const (
	MappingAny   = "any"
	MappingEqual = "eq"
	MappingLT    = "lt"
	MappingLTE   = "lte"
	MappingGT    = "gt"
	MappingGTE   = "gte"
)

func (m *ValueMapping) Validate() error {
	switch m.Op {
	case "":
		return fmt.Errorf("mapping op can't be empty")
	case MappingAny, MappingEqual, MappingLT, MappingLTE, MappingGT, MappingGTE:
		// okay
	default:
		return fmt.Errorf("invalid mapping op: %q", m.Op)
	}

	return nil
}

//------------------------------------------------------------------------------

func SelectGridItem(
	ctx context.Context, app *bunapp.App, itemID uint64,
) (GridItem, error) {
	baseItem := new(BaseGridItem)

	if err := app.PG.NewSelect().
		Model(baseItem).
		Where("id = ?", itemID).
		Scan(ctx); err != nil {
		return nil, err
	}

	return decodeBaseGridItem(baseItem)
}

func decodeBaseGridItem(baseItem *BaseGridItem) (GridItem, error) {
	switch baseItem.Type {
	case GridItemGauge:
		col := &GaugeGridItem{
			BaseGridItem: baseItem,
		}
		if err := baseItem.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil

	case GridItemChart:
		col := &ChartGridItem{
			BaseGridItem: baseItem,
		}
		if err := baseItem.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil

	case GridItemTable:
		col := &TableGridItem{
			BaseGridItem: baseItem,
		}
		if err := baseItem.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil

	case GridItemHeatmap:
		col := &HeatmapGridItem{
			BaseGridItem: baseItem,
		}
		if err := baseItem.Params.Decode(&col.Params); err != nil {
			return nil, err
		}
		return col, nil

	default:
		return nil, fmt.Errorf("unsupported grid item type: %s", baseItem.Type)
	}
}

func SelectGridItems(
	ctx context.Context, app *bunapp.App, dashID uint64,
) ([]GridItem, error) {
	baseItems, err := SelectBaseGridItems(ctx, app, dashID)
	if err != nil {
		return nil, err
	}

	gridItems := make([]GridItem, len(baseItems))
	for i, baseItem := range baseItems {
		gridItems[i], err = decodeBaseGridItem(baseItem)
		if err != nil {
			return nil, err
		}
	}
	return gridItems, nil
}

func SelectBaseGridItems(
	ctx context.Context, app *bunapp.App, dashID uint64,
) ([]*BaseGridItem, error) {
	var gridItems []*BaseGridItem
	if err := app.PG.NewSelect().
		Model(&gridItems).
		Where("dash_id = ?", dashID).
		OrderExpr("row_id ASC, y_axis ASC, x_axis ASC, id ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return gridItems, nil
}

func InsertGridItems(
	ctx context.Context, db bun.IDB, gridItems []GridItem,
) error {
	if _, err := db.NewInsert().
		Model(&gridItems).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
