package metrics

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/zeebo/xxh3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/histutil"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
)

type QueryHandler struct {
	*bunapp.App
}

func NewQueryHandler(app *bunapp.App) *QueryHandler {
	return &QueryHandler{
		App: app,
	}
}

func (h *QueryHandler) Table(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(QueryFilter)
	if err := DecodeQueryFilter(req, f); err != nil {
		return err
	}

	var grouping []string

	for _, part := range f.allParts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch value := part.AST.(type) {
		case *ast.Grouping:
			for _, elem := range value.Elems {
				if strings.HasPrefix(elem.Name, "$") {
					part.Error.Wrapped = errors.New("individual grouping is forbidden")
				} else {
					grouping = append(grouping, elem.Alias)
				}
			}
		}
	}

	metricMap, err := f.MetricMap(ctx, h.App)
	if err != nil {
		return err
	}

	tableName, groupingInterval := DatapointTableForGrouping(
		&f.TimeFilter, org.GroupingIntervalLarge)
	engine := mql.NewEngine(NewCHStorage(ctx, h.CH, &CHStorageConfig{
		ProjectID:  f.Project.ID,
		TimeFilter: f.TimeFilter,
		MetricMap:  metricMap,
		Search:     f.searchTokens,

		TableName:        tableName,
		GroupingInterval: groupingInterval,
	}))
	result := engine.Run(f.allParts)

	columns, table := convertToTable(result.Timeseries, result.Metrics, f.TableAgg)
	sortTable(ctx, h.App, columns, table, f)

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(
			attribute.Int64("num_timeseries", int64(len(result.Timeseries))),
			attribute.Int64("table_size", int64(len(table))),
		)
	}

	if len(table) == 0 {
		var firstErr error
		for _, part := range f.allParts {
			if part.Error.Wrapped != nil {
				firstErr = part.Error.Wrapped
				break
			}
		}
		if firstErr != nil {
			return firstErr
		}
	}

	var hasMore bool
	if len(table) > 1000 {
		table = table[:1000]
		hasMore = true
	}

	return httputil.JSON(w, bunrouter.H{
		"query":   f.parsedQuery,
		"columns": columns,
		"items":   table,
		"hasMore": hasMore,
		"order":   f.OrderByMixin,
	})
}

type ColumnInfo struct {
	Name      string `json:"name"`
	Unit      string `json:"unit"`
	IsGroup   bool   `json:"isGroup"`
	TableFunc string `json:"tableFunc"`
}

func convertToTable(
	timeseries []*mql.Timeseries, metrics []mql.MetricInfo, tableAgg map[string]string,
) ([]*ColumnInfo, []map[string]any) {
	columnMap := make(map[string]*ColumnInfo)
	var columns []*ColumnInfo

	for i := range metrics {
		metric := &metrics[i]
		if strings.HasPrefix(metric.Name, "_") {
			continue
		}
		col := &ColumnInfo{
			Name:      metric.Name,
			TableFunc: metric.TableFunc,
		}
		columnMap[metric.Name] = col
		columns = append(columns, col)
	}

	rowMap := make(map[uint64]map[string]any)

	var buf []byte
	for _, ts := range timeseries {
		col, ok := columnMap[ts.MetricName]
		if !ok {
			col := &ColumnInfo{
				Name: ts.MetricName,
			}
			columnMap[ts.MetricName] = col
			columns = append(columns, col)
		}
		col.Unit = ts.Unit

		for _, attrKey := range ts.Grouping {
			if _, ok := columnMap[attrKey]; !ok {
				col := &ColumnInfo{
					Name:    attrKey,
					IsGroup: true,
				}
				columnMap[attrKey] = col
				columns = append(columns, col)
			}
		}

		buf = ts.Attrs.Bytes(buf[:0], nil)
		hash := xxh3.Hash(buf)

		row, ok := rowMap[hash]
		if !ok {
			row = make(map[string]any)
			rowMap[hash] = row

			row["_name"] = ts.Attrs.String()
			row["_query"] = ts.WhereQuery()

			buf = ts.Attrs.Bytes(buf[:0], nil)
			row["_hash"] = strconv.FormatUint(xxh3.Hash(buf), 10)

			for _, kv := range ts.Attrs {
				if _, ok := columnMap[kv.Key]; !ok {
					col := &ColumnInfo{
						Name:    kv.Key,
						IsGroup: true,
					}
					columnMap[kv.Key] = col
					columns = append(columns, col)
				}

				row[kv.Key] = kv.Value
			}

			for k, v := range ts.Annotations {
				row[k] = v
			}
		}

		value := tableValue(ts.Value, tableAgg[col.Name])
		if col.Unit == bunconv.UnitTime {
			row[ts.MetricName] = value * 1000
		} else {
			row[ts.MetricName] = value
		}
	}

	table := make([]map[string]any, 0, len(rowMap))
	for _, row := range rowMap {
		for _, col := range columns {
			if _, ok := row[col.Name]; !ok {
				if col.IsGroup {
					row[col.Name] = ""
				} else {
					row[col.Name] = float64(0)
				}
			}
		}
		table = append(table, row)
	}

	return columns, table
}

func sortTable(
	ctx context.Context,
	app *bunapp.App,
	columns []*ColumnInfo,
	table []map[string]any,
	f *QueryFilter,
) {
	if len(table) == 0 {
		return
	}

	row := table[0]
	if _, ok := row[f.SortBy]; !ok {
		f.SortBy = columns[0].Name
		f.SortDesc = true
	}

	switch v := row[f.SortBy]; v.(type) {
	case nil:
		return
	case float64:
		slices.SortFunc(table, func(a, b map[string]any) int {
			v1, _ := a[f.SortBy].(float64)
			v2, _ := b[f.SortBy].(float64)
			if f.SortDesc {
				return cmp.Compare(v2, v1)
			}
			return cmp.Compare(v1, v2)
		})
	case string:
		slices.SortFunc(table, func(a, b map[string]any) int {
			v1, _ := a[f.SortBy].(string)
			v2, _ := b[f.SortBy].(string)
			if f.SortDesc {
				return strings.Compare(v2, v1)
			}
			return strings.Compare(v1, v2)
		})
	default:
		app.Zap(ctx).Error("unsupported table value type",
			zap.String("column", f.SortBy),
			zap.String("type", fmt.Sprintf("%T", v)))
	}
}

//------------------------------------------------------------------------------

type Timeseries struct {
	ID     uint64    `json:"id"`
	Name   string    `json:"name"`
	Metric string    `json:"metric"`
	Unit   string    `json:"unit"`
	Attrs  mql.Attrs `json:"attrs"`
	Value  []float64 `json:"value"`
}

func (h *QueryHandler) Timeseries(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(QueryFilter)
	if err := DecodeQueryFilter(req, f); err != nil {
		return err
	}

	if len(f.parsedQuery.Parts) == 0 {
		return httputil.JSON(w, bunrouter.H{
			"query":      f.parsedQuery,
			"timeseries": []any{},
			"time":       []any{},
			"columns":    []any{},
		})
	}

	metricMap, err := f.MetricMap(ctx, h.App)
	if err != nil {
		return err
	}

	timeseries, timeCol, metrics := h.selectTimeseries(ctx, f, metricMap)
	jsonTimeseries := make([]Timeseries, len(timeseries))

	columnMap := make(map[string]*ColumnInfo)
	var columns []*ColumnInfo

	for i, src := range timeseries {
		dest := &jsonTimeseries[i]

		name := src.Name()
		dest.ID = xxh3.HashString(name)
		dest.Name = name
		dest.Metric = src.MetricName
		dest.Unit = src.Unit
		dest.Attrs = src.Attrs
		dest.Value = src.Value

		if _, ok := columnMap[dest.Metric]; !ok {
			col := &ColumnInfo{
				Name: dest.Metric,
				Unit: dest.Unit,
			}
			columnMap[dest.Metric] = col
			columns = append(columns, col)
		}
	}

	for i := range metrics {
		metric := &metrics[i]
		if _, ok := columnMap[metric.Name]; ok {
			continue
		}
		columns = append(columns, &ColumnInfo{
			Name: metric.Name,
			// no unit
		})
	}

	return httputil.JSON(w, bunrouter.H{
		"query":      f.parsedQuery,
		"timeseries": jsonTimeseries,
		"time":       timeCol,
		"columns":    columns,
	})
}

func (h *QueryHandler) selectTimeseries(
	ctx context.Context, f *QueryFilter, metricMap map[string]*Metric,
) ([]*mql.Timeseries, []time.Time, []mql.MetricInfo) {
	tableName, groupingInterval := DatapointTableForGrouping(
		&f.TimeFilter, org.GroupingIntervalLarge)
	storage := NewCHStorage(ctx, h.CH, &CHStorageConfig{
		ProjectID:  f.Project.ID,
		TimeFilter: f.TimeFilter,
		MetricMap:  metricMap,

		TableName:        tableName,
		GroupingInterval: groupingInterval,
	})
	engine := mql.NewEngine(storage)
	result := engine.Run(f.allParts)
	timeCol := bunutil.FillTime(nil, f.TimeGTE, f.TimeLT, groupingInterval)
	return result.Timeseries, timeCol, result.Metrics
}

//------------------------------------------------------------------------------

func (h *QueryHandler) Gauge(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(QueryFilter)
	if err := DecodeQueryFilter(req, f); err != nil {
		return err
	}

	for _, part := range f.allParts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch part.AST.(type) {
		case *ast.Grouping:
			part.Error.Wrapped = errors.New("grouping is forbidden")
		}
	}

	metricMap, err := f.MetricMap(ctx, h.App)
	if err != nil {
		return err
	}

	tableName, groupingInterval := DatapointTableForGrouping(
		&f.TimeFilter, org.GroupingIntervalLarge)
	storage := NewCHStorage(ctx, h.CH, &CHStorageConfig{
		ProjectID:  f.Project.ID,
		TimeFilter: f.TimeFilter,
		MetricMap:  metricMap,

		TableName:        tableName,
		GroupingInterval: groupingInterval,
	})
	engine := mql.NewEngine(storage)
	result := engine.Run(f.allParts)

	columns, table := convertToTable(result.Timeseries, result.Metrics, f.TableAgg)

	var values map[string]any
	if len(table) > 0 {
		values = table[0]
		delete(values, "_query")
	} else {
		values = make(map[string]any)
	}

	return httputil.JSON(w, bunrouter.H{
		"query":   f.parsedQuery,
		"columns": columns,
		"values":  values,
	})
}

func (h *QueryHandler) Heatmap(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(QueryFilter)
	if err := DecodeQueryFilter(req, f); err != nil {
		return err
	}

	heatmap, err := h.selectMetricHeatmap(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"query":   f.parsedQuery,
		"heatmap": heatmap,
	})
}

func (h *QueryHandler) selectMetricHeatmap(
	ctx context.Context, f *QueryFilter,
) (*histutil.Heatmap, error) {
	tableName, groupingInterval := DatapointTableForGrouping(
		&f.TimeFilter, org.GroupingIntervalLarge)

	q := h.CH.NewSelect().
		ColumnExpr("quantilesBFloat16MergeState(0.5, 0.9, 0.99)(histogram) AS value").
		ColumnExpr("toStartOfInterval(time, INTERVAL ? minute) AS time_",
			groupingInterval.Minutes()).
		TableExpr("?", ch.Name(tableName)).
		Where("project_id = ?", f.Project.ID).
		Where("metric = ?", f.Metric[0]).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		GroupExpr("time_").
		OrderExpr("time_").
		Limit(10000)

	for _, part := range f.allParts {
		if part.Error.Wrapped != nil {
			continue
		}
		switch ast := part.AST.(type) {
		case *ast.Selector:
			part.Error.Wrapped = errors.New("not supported by heatmap")
		case *ast.Grouping:
			part.Error.Wrapped = errors.New("not supported by heatmap")
		case *ast.Where:
			if err := compileFilters(q, InstrumentHistogram, ast.Filters); err != nil {
				part.Error.Wrapped = err
			}
		default:
			return nil, fmt.Errorf("unexpected ast: %T", ast)
		}
	}

	var bfloat16Col []map[bfloat16.T]uint64
	var timeCol []time.Time

	if err := q.ScanColumns(ctx, &bfloat16Col, &timeCol); err != nil {
		return nil, err
	}

	tdigestCol := make([][]float32, len(bfloat16Col))
	for i, m := range bfloat16Col {
		tdigest := make([]float32, 0, 2*len(m))
		for value, count := range m {
			tdigest = append(tdigest, value.Float32(), float32(count))
		}
		tdigestCol[i] = tdigest
	}

	tdigestCol = bunutil.Fill(tdigestCol, timeCol, nil, f.TimeGTE, f.TimeLT, groupingInterval)
	timeCol = bunutil.FillTime(timeCol, f.TimeGTE, f.TimeLT, groupingInterval)
	heatmap := histutil.BuildHeatmap(tdigestCol, timeCol)

	return heatmap, nil
}

//------------------------------------------------------------------------------

func tableValue(value []float64, aggFunc string) float64 {
	switch aggFunc {
	case mql.TableFuncMin:
		return minTableValue(value)
	case mql.TableFuncMax:
		return maxTableValue(value)
	case mql.TableFuncAvg:
		return avgTableValue(value)
	case mql.TableFuncSum:
		return sumTableValue(value)
	case "", mql.TableFuncLast:
		return lastTableValue(value)
	default:
		return math.NaN()
	}
}

func minTableValue(ns []float64) float64 {
	min := math.MaxFloat64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n < min {
			min = n
		}
	}
	if min != math.MaxFloat64 {
		return min
	}
	return math.NaN()
}

func maxTableValue(ns []float64) float64 {
	max := -math.MaxFloat64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n > max {
			max = n
		}
	}
	if max != -math.MaxFloat64 {
		return max
	}
	return math.NaN()
}

func lastTableValue(ns []float64) float64 {
	for i := len(ns) - 1; i >= 0; i-- {
		n := ns[i]
		if !math.IsNaN(n) {
			return n
		}
	}
	return math.NaN()
}

func avgTableValue(ns []float64) float64 {
	sum, count := sumCount(ns)
	if count > 0 {
		return sum / float64(count)
	}
	return math.NaN()
}

func sumTableValue(ns []float64) float64 {
	sum, _ := sumCount(ns)
	return sum
}

func sumCount(ns []float64) (float64, int) {
	var sum float64
	var count int
	for _, n := range ns {
		if !math.IsNaN(n) {
			sum += n
			count++
		}
	}
	return sum, count
}

func deltaTableValue(value []float64) float64 {
	for i, num := range value {
		if !math.IsNaN(num) {
			value = value[i:]
			break
		}
	}

	if len(value) == 0 {
		return 0
	}

	prevNum := value[0]
	value = value[1:]
	var sum float64

	for _, num := range value {
		if math.IsNaN(num) || num <= prevNum {
			continue
		}
		sum += num - prevNum
		prevNum = num
	}

	return sum
}
