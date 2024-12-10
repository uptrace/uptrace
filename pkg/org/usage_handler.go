package org

import (
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/fx"
	"golang.org/x/exp/constraints"
)

type UsageHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	CH     *ch.DB
}

type UsageHandler struct {
	*UsageHandlerParams
}

func NewUsageHandler(p UsageHandlerParams) *UsageHandler {
	return &UsageHandler{&p}
}

func registerUsageHandler(h *UsageHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.User).
		WithGroup("", func(g *bunrouter.Group) {
			g.GET("/data-usage", h.Show)
		})
}

type Usage struct {
	Spans      []uint64    `json:"spans"`
	Bytes      []uint64    `json:"bytes"`
	Timeseries []uint64    `json:"timeseries"`
	Time       []time.Time `json:"time"`
}

type Usage2 struct {
	Datapoints []uint64
	Time       []time.Time
	Minutes    []uint64
}

func (h *UsageHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	interval := 24 * time.Hour
	timeLT := time.Now().Truncate(interval).Add(interval)
	timeGTE := timeLT.AddDate(0, -1, 0)

	usage := new(Usage)
	if err := h.CH.NewSelect().
		ColumnExpr("sum(rows) AS spans").
		ColumnExpr("sum(data_uncompressed_bytes) AS bytes").
		ColumnExpr("parseDateTime(partition, '%Y-%m-%d') AS time").
		TableExpr("system.parts").
		Where("database = ?", h.CH.Config().Database).
		Where("table = ?", "spans_data").
		Where("min_time >= ?", timeGTE).
		Where("active").
		GroupExpr("partition").
		OrderExpr("time ASC").
		ScanColumns(ctx, usage); err != nil {
		return err
	}

	usage.Spans = bunutil.Fill(usage.Spans, usage.Time, 0, timeGTE, timeLT, interval)
	usage.Bytes = bunutil.Fill(usage.Bytes, usage.Time, 0, timeGTE, timeLT, interval)
	usage.Time = bunutil.FillTime(usage.Time, timeGTE, timeLT, interval)

	subq := h.CH.NewSelect().
		ColumnExpr("60 * uniqCombined64(15)(d.attrs_hash) AS datapoints").
		ColumnExpr("d.time").
		ColumnExpr("60 AS minutes").
		TableExpr("datapoint_hours AS d").
		Where("d.time >= ?", timeGTE).
		Where("d.time < ?", timeLT).
		GroupExpr("d.project_id, d.time")

	usage2 := new(Usage2)
	if err := h.CH.NewSelect().
		ColumnExpr("sum(d.datapoints) AS datapoints").
		ColumnExpr("toStartOfDay(d.time) AS time").
		ColumnExpr("sum(d.minutes) AS minutes").
		TableExpr("(?) AS d", subq).
		GroupExpr("toStartOfDay(d.time)").
		OrderExpr("time ASC").
		ScanColumns(ctx, usage2); err != nil {
		return err
	}

	usage.Timeseries = make([]uint64, len(usage2.Datapoints))
	for i, datapoints := range usage2.Datapoints {
		usage.Timeseries[i] = datapoints / usage2.Minutes[i]
	}
	usage.Timeseries = bunutil.Fill(usage.Timeseries, usage2.Time, 0, timeGTE, timeLT, interval)

	spans := sum(usage.Spans)
	bytes := sum(usage.Bytes)
	datapoints := sum(usage2.Datapoints)
	minutes := sum(usage2.Minutes)

	var timeseries uint64
	if minutes > 0 {
		timeseries = datapoints / minutes
	}

	return httputil.JSON(w, bunrouter.H{
		"usage":      usage,
		"spans":      spans,
		"bytes":      bytes,
		"timeseries": timeseries,
		"startTime":  timeGTE,
		"endTime":    timeLT,
	})
}

func sum[T constraints.Integer](slice []T) T {
	var sum T
	for _, x := range slice {
		sum += x
	}
	return sum
}
