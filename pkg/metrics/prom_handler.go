package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	kitzap "github.com/go-kit/kit/log/zap"
	"github.com/grafana/regexp"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	promstorage "github.com/prometheus/prometheus/storage"
	promapi "github.com/prometheus/prometheus/web/api/v1"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/grafana"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap/zapcore"
)

type PromHandler struct {
	grafana.GrafanaBaseHandler

	promapi *promapi.API
	promeng *promql.Engine
}

func NewPromHandler(app *bunapp.App) *PromHandler {
	promeng := promql.NewEngine(promql.EngineOpts{
		Logger:                   kitzap.NewZapSugarLogger(app.ZapLogger().Logger, zapcore.InfoLevel),
		Reg:                      nil,
		MaxSamples:               100000,
		Timeout:                  time.Second * 30,
		ActiveQueryTracker:       nil,
		LookbackDelta:            0,
		NoStepSubqueryIntervalFn: nil,
		EnableAtModifier:         false,
		EnableNegativeOffset:     false,
	})
	return &PromHandler{
		GrafanaBaseHandler: grafana.GrafanaBaseHandler{
			App: app,
		},

		promapi: &promapi.API{
			Queryable:         nil,
			QueryEngine:       promeng,
			ExemplarQueryable: nil,
			CORSOrigin:        regexp.MustCompile("\\*"),
		},
		promeng: promeng,
	}
}

func (h *PromHandler) Metadata(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte(`{"status":"success","data":[]}`))
	return err
}

func (h *PromHandler) Labels(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodePromFilter(h.App, req)
	if err != nil {
		return err
	}

	var labels []string

	if err := h.CH().NewSelect().
		Distinct().
		ColumnExpr("arrayJoin(keys)").
		TableExpr("?", measureTableForWhere(h.App, &f.TimeFilter)).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		ScanColumns(ctx, &labels); err != nil {
		return err
	}

	labels = append(labels, "__name__")

	return httputil.JSON(w, map[string]any{
		"status": "success",
		"data":   labels,
	})
}

func (h *PromHandler) Series(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in struct {
		Match []string
		Start float64
		End   float64
	}

	if err := bunapp.UnmarshalValues(req, &in); err != nil {
		return err
	}

	queryable := newPromQueryable(ctx, h.App, project.ID)
	seriesSlice := make([][][]string, 0)

	for _, match := range in.Match {
		q, err := h.promapi.QueryEngine.NewRangeQuery(
			queryable,
			match,
			time.Unix(int64(in.Start), 0),
			time.Unix(int64(in.End), 0),
			time.Second,
		)
		if err != nil {
			return err
		}

		stmt, ok := q.Statement().(*parser.EvalStmt)
		if !ok {
			return fmt.Errorf("metrics: unexpected statement type: %T", q.Statement())
		}

		sel, ok := stmt.Expr.(*parser.VectorSelector)
		if !ok {
			return fmt.Errorf("metrics: unexpected expr type: %T", stmt.Expr)
		}

		series, err := queryable.querier().Series(&promstorage.SelectHints{
			Start: int64(in.Start * 1000),
			End:   int64(in.End * 1000),
		}, sel.LabelMatchers...)
		if err != nil {
			return err
		}

		seriesSlice = append(seriesSlice, series...)
	}

	data := make([]map[string]string, len(seriesSlice))

	for i, series := range seriesSlice {
		labels := make(map[string]string, len(series)*15/10)
		data[i] = labels

		for _, kv := range series {
			labels[kv[0]] = kv[1]
		}
	}

	return httputil.JSON(w, map[string]any{
		"status": "success",
		"data":   data,
	})
}

func (h *PromHandler) LabelValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodePromFilter(h.App, req)
	if err != nil {
		return err
	}

	if f.Label == "" {
		return errors.New("'label' query param is required")
	}

	q := h.CH().NewSelect().
		Distinct().
		ColumnExpr("?", chColumn(f.Label)).
		TableExpr("?", measureTableForWhere(h.App, &f.TimeFilter)).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT)

	if f.Label != "__name__" {
		q = q.Where("has(keys, ?)", f.Label)
	}

	var values []string

	if err := q.ScanColumns(ctx, &values); err != nil {
		return err
	}

	return httputil.JSON(w, map[string]any{
		"status": "success",
		"data":   values,
	})
}

func (h *PromHandler) QueryRange(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodePromFilter(h.App, req)
	if err != nil {
		return err
	}

	if f.Step <= 0 {
		return errors.New("zero or negative query resolution step widths are not accepted. Try a positive integer")
	}
	// For safety, limit the number of returned points per timeseries.
	// This is sufficient for 60s resolution for a week or 1h resolution for a year.
	if f.End.Sub(f.Start)/f.Step > 11000 {
		return errors.New("exceeded maximum resolution of 11,000 points per timeseries. Try decreasing the query resolution (?step=XX)")
	}
	if f.Query == "" {
		return errors.New("'query' param is required")
	}

	queryable := newPromQueryable(ctx, h.App, f.ProjectID)

	rangeQuery, err := h.promapi.QueryEngine.NewRangeQuery(
		queryable,
		f.Query,
		f.TimeGTE,
		f.TimeLT,
		f.Step,
	)
	if err != nil {
		return err
	}

	res := rangeQuery.Exec(req.Context())
	if res.Err != nil {
		return res.Err
	}
	if err := writePromqlResult(w, res); err != nil {
		return err
	}

	return nil
}

func (h *PromHandler) QueryInstant(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in struct {
		Time  float64
		Query string
	}

	if err := bunapp.UnmarshalValues(req, &in); err != nil {
		return err
	}

	if in.Query == "" {
		return errors.New("'query' param is required")
	}
	if int64(in.Time) == 0 {
		in.Time = float64(time.Now().Unix())
	}

	queryable := newPromQueryable(ctx, h.App, project.ID)

	promQuery, err := h.promapi.QueryEngine.NewInstantQuery(
		queryable,
		in.Query,
		time.Unix(int64(in.Time), 0),
	)
	if err != nil {
		return err
	}

	res := promQuery.Exec(ctx)
	if res.Err != nil {
		return res.Err
	}
	if err := writePromqlResult(w, res); err != nil {
		return err
	}

	return nil
}

func writePromqlResult(w http.ResponseWriter, res *promql.Result) error {
	var result any

	switch value := res.Value.(type) {
	case promql.Matrix:
		result = promqlMatrixValue(value)
	case promql.Vector:
		result = promqlVectorValue(value)
	case promql.Scalar:
		result = []any{value.T, fmt.Sprint(value.V)}
	default:
		return fmt.Errorf("unsupported promql value: %T", res.Value)
	}

	return httputil.JSON(w, bunrouter.H{
		"status": "success",
		"data": bunrouter.H{
			"resultType": res.Value.Type(),
			"result":     result,
		},
	})
}

type matrixItem struct {
	Metric map[string]string `json:"metric"`
	Values [][]any           `json:"values"`
}

func promqlMatrixValue(value promql.Matrix) []matrixItem {
	matrix := make([]matrixItem, value.Len())
	for i, sample := range value {
		item := matrixItem{
			Metric: make(map[string]string, sample.Metric.Len()*15/10),
			Values: make([][]any, len(sample.Points)),
		}
		matrix[i] = item

		for _, v := range sample.Metric {
			item.Metric[v.Name] = v.Value
		}

		for j, point := range sample.Points {
			item.Values[j] = []any{float64(point.T) / 1000, fmt.Sprintf("%f", point.V)}
		}

	}
	return matrix
}

type vectorItem struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

func promqlVectorValue(value promql.Vector) []vectorItem {
	vector := make([]vectorItem, len(value))
	for i, sample := range value {
		entry := vectorItem{
			Metric: make(map[string]string, sample.Metric.Len()*15/10),
			Value:  []any{float64(sample.T / 1000), fmt.Sprint(sample.V)},
		}
		vector[i] = entry

		for _, label := range sample.Metric {
			entry.Metric[label.Name] = label.Value
		}
	}
	return vector
}
