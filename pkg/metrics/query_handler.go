package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bunrouter"
	"golang.org/x/exp/slices"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
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

	f, err := decodeQueryFilter(h.App, req)
	if err != nil {
		return err
	}

	var grouping []string

	for _, part := range f.allParts {
		if part.Error.Wrapped != nil {
			continue
		}

		switch expr := part.AST.(type) {
		case *ast.Grouping:
			for _, name := range expr.Names {
				if strings.HasPrefix(name, "$") {
					part.Error.Wrapped = errors.New("individual grouping is forbidden")
					continue
				}
				grouping = append(grouping, name)
			}
		}
	}

	storage := NewCHStorage(ctx, h.App.CH, &CHStorageConfig{
		Projects:   []uint32{f.ProjectID},
		TimeFilter: f.TimeFilter,
		MetricMap:  f.metricMap,

		TableName:      measureTableForWhere(h.App, &f.TimeFilter),
		GroupingPeriod: f.TimeFilter.Duration(),
	})
	engine := upql.NewEngine(storage)
	result := engine.Run(f.allParts)

	columns, table := convertToTable(result.Timeseries, result.Columns)
	sortTable(columns, table, f)

	var hasMore bool
	if len(table) > 1000 {
		table = table[:1000]
		hasMore = true
	}

	return httputil.JSON(w, bunrouter.H{
		"baseQueryParts": f.baseQueryParts,
		"queryParts":     f.queryParts,
		"columns":        columns,
		"items":          table,
		"hasMore":        hasMore,
		"order":          f.OrderByMixin,
	})
}

type ColumnInfo struct {
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	IsGroup bool   `json:"isGroup"`
}

func convertToTable(
	timeseries []upql.Timeseries, columnNames []string,
) ([]*ColumnInfo, []map[string]any) {
	columnMap := make(map[string]*ColumnInfo)
	var columns []*ColumnInfo
	rowMap := make(map[uint64]map[string]any)

	for _, columnName := range columnNames {
		if strings.HasPrefix(columnName, "_") {
			continue
		}
		col := &ColumnInfo{
			Name: columnName,
		}
		columnMap[columnName] = col
		columns = append(columns, col)
	}

	var buf []byte
	for i := range timeseries {
		ts := &timeseries[i]
		metricName := ts.MetricName()

		if col, ok := columnMap[metricName]; ok {
			col.Unit = ts.Unit
		} else {
			col := &ColumnInfo{
				Name: metricName,
				Unit: ts.Unit,
			}
			columnMap[metricName] = col
			columns = append(columns, col)
		}

		buf = ts.Attrs.Bytes(buf[:0])
		hash := xxhash.Sum64(buf)

		row, ok := rowMap[hash]
		if !ok {
			row = make(map[string]any)
			rowMap[hash] = row

			row[attrkey.ItemQuery] = ts.WhereQuery()

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
		}

		row[metricName] = ts.Value[0]
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

func sortTable(columns []*ColumnInfo, table []map[string]any, f *QueryFilter) {
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
		slices.SortFunc(table, func(a, b map[string]any) bool {
			v1, _ := a[f.SortBy].(float64)
			v2, _ := b[f.SortBy].(float64)
			if f.SortDesc {
				return v1 > v2
			}
			return v1 < v2
		})
	case string:
		slices.SortFunc(table, func(a, b map[string]any) bool {
			v1, _ := a[f.SortBy].(string)
			v2, _ := b[f.SortBy].(string)
			if f.SortDesc {
				return strings.Compare(v1, v2) == 1
			}
			return strings.Compare(v1, v2) == -1
		})
	default:
		panic(fmt.Errorf("unsupported value type: %T", v))
	}
}

//------------------------------------------------------------------------------

func (h *QueryHandler) Gauge(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeQueryFilter(h.App, req)
	if err != nil {
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

	storage := NewCHStorage(ctx, h.App.CH, &CHStorageConfig{
		Projects:   []uint32{f.ProjectID},
		TimeFilter: f.TimeFilter,
		MetricMap:  f.metricMap,

		TableName:      measureTableForWhere(h.App, &f.TimeFilter),
		GroupingPeriod: f.TimeFilter.Duration(),
	})
	engine := upql.NewEngine(storage)
	result := engine.Run(f.allParts)

	columns, table := convertToTable(result.Timeseries, result.Columns)

	var values map[string]any
	if len(table) > 0 {
		values = table[0]
		delete(values, attrkey.ItemQuery)
	} else {
		values = make(map[string]any)
	}

	return httputil.JSON(w, bunrouter.H{
		"baseQueryParts": f.baseQueryParts,
		"queryParts":     f.queryParts,
		"columns":        columns,
		"values":         values,
	})
}

//------------------------------------------------------------------------------

type Timeseries struct {
	Name   string      `json:"name"`
	Metric string      `json:"metric"`
	Unit   string      `json:"unit"`
	Attrs  upql.Attrs  `json:"attrs"`
	Value  []float64   `json:"value"`
	Time   []time.Time `json:"time"`
}

func (h *QueryHandler) Timeseries(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeQueryFilter(h.App, req)
	if err != nil {
		return err
	}

	if len(f.queryParts) == 0 {
		return httputil.JSON(w, bunrouter.H{
			"baseQueryParts": f.baseQueryParts,
			"queryParts":     f.queryParts,
			"timeseries":     []struct{}{},
		})
	}

	tableName, groupingPeriod := measureTableForGroup(h.App, &f.TimeFilter, org.GroupPeriod)
	storage := NewCHStorage(ctx, h.CH, &CHStorageConfig{
		Projects:   []uint32{f.ProjectID},
		TimeFilter: f.TimeFilter,
		MetricMap:  f.metricMap,

		TableName:      tableName,
		GroupingPeriod: groupingPeriod,

		GroupByTime: true,
		FillHoles:   true,
	})
	engine := upql.NewEngine(storage)
	result := engine.Run(f.allParts)

	jsonTimeseries := make([]Timeseries, len(result.Timeseries))

	for i := range result.Timeseries {
		src := &result.Timeseries[i]
		dest := &jsonTimeseries[i]

		dest.Name = src.Name()
		dest.Metric = src.MetricName()
		dest.Unit = src.Unit
		dest.Attrs = src.Attrs
		dest.Value = src.Value
		dest.Time = src.Time
	}

	return httputil.JSON(w, bunrouter.H{
		"baseQueryParts": f.baseQueryParts,
		"queryParts":     f.queryParts,
		"timeseries":     jsonTimeseries,
	})
}
