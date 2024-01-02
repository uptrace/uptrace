package grafana

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	promparser "github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage"
	promstorage "github.com/prometheus/prometheus/storage"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type PromHandler struct {
	BaseGrafanaHandler

	promqlEngine *promql.Engine
}

func NewPromHandler(app *bunapp.App) *PromHandler {
	return &PromHandler{
		BaseGrafanaHandler: BaseGrafanaHandler{
			App: app,
		},

		promqlEngine: promql.NewEngine(promql.EngineOpts{
			MaxSamples:    1_000_000,
			Timeout:       30 * time.Second,
			LookbackDelta: 5 * time.Minute,
			// EnableAtModifier and EnableNegativeOffset have to be
			// always on for regular PromQL as of Prometheus v2.33.
			EnableAtModifier:     true,
			EnableNegativeOffset: true,
		}),
	}
}

func (h *PromHandler) Metadata(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte(`{"status":"success","data":[]}`))
	return err
}

func (h *PromHandler) LabelNames(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodePromFilter(h.App, req)
	if err != nil {
		return err
	}

	var labels []string

	tableName := metrics.DatapointTableForWhere(&f.TimeFilter)
	if err := h.CH.NewSelect().
		Distinct().
		ColumnExpr("arrayJoin(d.string_keys)").
		TableExpr("? AS d", ch.Name(tableName)).
		Where("d.project_id = ?", f.ProjectID).
		Where("d.time >= ?", f.TimeGTE).
		Where("d.time < ?", f.TimeLT).
		ScanColumns(ctx, &labels); err != nil {
		return err
	}

	labels = append(labels, promlabels.MetricName)

	return httputil.JSON(w, map[string]any{
		"status": "success",
		"data":   labels,
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

	tableName := metrics.DatapointTableForWhere(&f.TimeFilter)
	q := h.CH.NewSelect().
		Distinct().
		ColumnExpr("?", chExpr(f.Label)).
		TableExpr("? AS d", ch.Name(tableName)).
		Where("d.project_id = ?", f.ProjectID).
		Where("d.time >= ?", f.TimeGTE).
		Where("d.time < ?", f.TimeLT)

	if f.Label != promlabels.MetricName {
		q = q.Where("has(d.string_keys, ?)", f.Label)
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
	promStorage := NewPromStorage(h.App, project.ID)

	querier, err := promStorage.Querier(startMillis, endMillis)
	if err != nil {
		return err
	}
	defer querier.Close()

	hints := &promstorage.SelectHints{
		Start: startMillis,
		End:   endMillis,
		// There is no series function, this token is used for lookups that don't need samples.
		Func: "series",
	}

	var set storage.SeriesSet

	if len(matcherSets) > 1 {
		var sets []storage.SeriesSet
		for _, mset := range matcherSets {
			s := querier.Select(ctx, true, hints, mset...)
			sets = append(sets, s)
		}
		set = promstorage.NewMergeSeriesSet(sets, storage.ChainedSeriesMerge)
	} else {
		set = querier.Select(ctx, false, hints, matcherSets[0]...)
	}

	var metrics []promlabels.Labels
	for set.Next() {
		metrics = append(metrics, set.At().Labels())
	}

	warnings := set.Warnings()
	if err := set.Err(); err != nil {
		return err
	}

	return httputil.JSON(w, map[string]any{
		"status":   "success",
		"data":     metrics,
		"warnings": warnings,
	})
}

func parseMatchersParam(matchers []string) ([][]*promlabels.Matcher, error) {
	matcherSets := make([][]*promlabels.Matcher, 0, len(matchers))
	for _, s := range matchers {
		matchers, err := promparser.ParseMetricSelector(s)
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

	queryable := NewPromStorage(h.App, f.ProjectID)

	queryOpts, err := promQueryOpts(req.Request)
	if err != nil {
		return err
	}

	rangeQuery, err := h.promqlEngine.NewRangeQuery(
		ctx,
		queryable,
		queryOpts,
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

	queryable := NewPromStorage(h.App, project.ID)

	queryOpts, err := promQueryOpts(req.Request)
	if err != nil {
		return err
	}

	promQuery, err := h.promqlEngine.NewInstantQuery(
		ctx,
		queryable,
		queryOpts,
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

func (h *PromHandler) EnablePromCompat(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		project := org.ProjectFromContext(ctx)

		if !project.PromCompat {
			if _, err := h.PG.NewUpdate().
				Model(project).
				Set("prom_compat = TRUE").
				Where("id = ?", project.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return next(w, req)
	}
}

func promQueryOpts(r *http.Request) (promql.QueryOpts, error) {
	var duration time.Duration

	if strDuration := r.FormValue("lookback_delta"); strDuration != "" {
		parsedDuration, err := urlstruct.ParseDuration(strDuration)
		if err != nil {
			return nil, fmt.Errorf("error parsing lookback delta duration: %w", err)
		}
		duration = parsedDuration
	}

	return promql.NewPrometheusQueryOpts(r.FormValue("stats") == "all", duration), nil
}

type promQueryData struct {
	ResultType parser.ValueType `json:"resultType"`
	Result     parser.Value     `json:"result"`
}

func writePromqlResult(w http.ResponseWriter, res *promql.Result) error {
	if res.Err != nil {
		return res.Err
	}

	return httputil.JSON(w, bunrouter.H{
		"status": "success",
		"data": promQueryData{
			ResultType: res.Value.Type(),
			Result:     res.Value,
		},
	})
}

func isEmptyMatcherSet(matchers []*promlabels.Matcher) bool {
	for _, lm := range matchers {
		if lm != nil && !lm.Matches("") {
			return false
		}
	}
	return true
}
