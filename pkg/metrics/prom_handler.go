package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage"
	promstorage "github.com/prometheus/prometheus/storage"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type PromHandler struct {
	org.BaseGrafanaHandler
}

func NewPromHandler(app *bunapp.App) *PromHandler {
	return &PromHandler{
		BaseGrafanaHandler: org.BaseGrafanaHandler{
			App: app,
		},
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

	if err := h.CH.NewSelect().
		Distinct().
		ColumnExpr("arrayJoin(keys)").
		TableExpr("?", measureTableForWhere(h.App, &f.TimeFilter)).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		ScanColumns(ctx, &labels); err != nil {
		return err
	}

	labels = append(labels, promlabels.MetricName)

	return httputil.JSON(w, map[string]any{
		"status": "success",
		"data":   labels,
	})
}

func (h *PromHandler) Series(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in struct {
		Match        []string
		StartSeconds float64 `urlstruct:"start"`
		EndSeconds   float64 `urlstruct:"end"`
	}

	if err := bunapp.UnmarshalValues(req, &in); err != nil {
		return err
	}

	if len(in.Match) == 0 {
		return fmt.Errorf("'match' query param is required")
	}
	matcherSets, err := parseMatchersParam(in.Match)
	if err != nil {
		return err
	}

	startMillis := int64(in.StartSeconds * 1000)
	endMillis := int64(in.EndSeconds * 1000)
	promStorage := NewPromStorage(ctx, h.App, project.ID)

	q, err := promStorage.Querier(ctx, startMillis, endMillis)
	if err != nil {
		return err
	}
	defer q.Close()

	var sets []promstorage.SeriesSet

	hints := &promstorage.SelectHints{
		Start: startMillis,
		End:   endMillis,
		// There is no series function, this token is used for lookups that don't need samples.
		Func: "series",
	}
	for _, mset := range matcherSets {
		// We need to sort this select results to merge (deduplicate) the series sets later.
		s := q.Select(true, hints, mset...)
		sets = append(sets, s)
	}

	var metrics []promlabels.Labels

	set := storage.NewMergeSeriesSet(sets, storage.ChainedSeriesMerge)
	for set.Next() {
		metrics = append(metrics, set.At().Labels())
	}
	if err := set.Err(); err != nil {
		return err
	}

	return httputil.JSON(w, map[string]any{
		"status":   "success",
		"data":     metrics,
		"warnings": set.Warnings(),
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

	q := h.CH.NewSelect().
		Distinct().
		ColumnExpr("?", chColumn(f.Label)).
		TableExpr("?", measureTableForWhere(h.App, &f.TimeFilter)).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT)

	if f.Label != promlabels.MetricName {
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
		return fmt.Errorf(`invalid "step" value: %q`, f.Step)
	}
	if f.End.Sub(f.Start)/f.Step > 10000 {
		return fmt.Errorf(`"step" is too small: %q`, f.Step)
	}
	if f.Query == "" {
		return errors.New(`"query" param is required`)
	}

	queryable := NewPromStorage(ctx, h.App, f.ProjectID)

	rangeQuery, err := h.QueryEngine.NewRangeQuery(
		queryable,
		&promql.QueryOpts{},
		f.Query,
		f.TimeGTE,
		f.TimeLT,
		f.Step,
	)
	if err != nil {
		return err
	}
	defer rangeQuery.Close()

	res := rangeQuery.Exec(ctx)
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
		return errors.New(`"query" param is required`)
	}
	if int64(in.Time) == 0 {
		in.Time = float64(time.Now().Unix())
	}

	queryable := NewPromStorage(ctx, h.App, project.ID)

	promQuery, err := h.QueryEngine.NewInstantQuery(
		queryable,
		&promql.QueryOpts{},
		in.Query,
		time.Unix(int64(in.Time), 0),
	)
	if err != nil {
		return err
	}

	res := promQuery.Exec(ctx)
	if err := writePromqlResult(w, res); err != nil {
		return err
	}

	return nil
}

func writePromqlResult(w http.ResponseWriter, res *promql.Result) error {
	if res.Err != nil {
		return res.Err
	}

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
			item.Values[j] = []any{
				float64(point.T) / 1000,
				strconv.FormatFloat(point.V, 'f', -1, 64),
			}
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

func parseMatchersParam(matchers []string) ([][]*promlabels.Matcher, error) {
	matcherSets := make([][]*promlabels.Matcher, 0, len(matchers))
	for _, s := range matchers {
		matchers, err := parser.ParseMetricSelector(s)
		if err != nil {
			return nil, err
		}
		if isEmptyMatcherSet(matchers) {
			return nil, errors.New("match[] must contain at least one non-empty matcher")
		}
		matcherSets = append(matcherSets, matchers)
	}
	return matcherSets, nil
}
