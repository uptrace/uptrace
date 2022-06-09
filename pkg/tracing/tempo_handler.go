package tracing

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/uuid"
	"go.uber.org/zap"
)

const projectIDHeaderKey = "uptrace-project-id"

var jsonMarshaler = &jsonpb.Marshaler{}

type TempoHandler struct {
	GrafanaBaseHandler

	*bunapp.App
}

func NewTempoHandler(app *bunapp.App) *TempoHandler {
	return &TempoHandler{
		GrafanaBaseHandler: GrafanaBaseHandler{
			App: app,
		},
		App: app,
	}
}

func (h *TempoHandler) QueryTrace(w http.ResponseWriter, req bunrouter.Request) error {
	contentType := req.Header.Get("Accept")
	if contentType == "" {
		contentType = protobufContentType
	}
	return h.queryTrace(w, req, contentType)
}

func (h *TempoHandler) QueryTraceJSON(w http.ResponseWriter, req bunrouter.Request) error {
	return h.queryTrace(w, req, jsonContentType)
}

func (h *TempoHandler) queryTrace(
	w http.ResponseWriter, req bunrouter.Request, contentType string,
) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans, err := SelectTraceSpans(ctx, h.App, traceID)
	if err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	resp := newTempopbTrace(h.App, traceID, spans)

	switch contentType {
	case "*/*", jsonContentType:
		w.Header().Set("Content-Type", jsonContentType)

		return jsonMarshaler.Marshal(w, resp)
	case protobufContentType:
		w.Header().Set("Content-Type", protobufContentType)

		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	default:
		return fmt.Errorf("unknown content type: %q", contentType)
	}
}

func (h *TempoHandler) Tags(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	q := h.CH().NewSelect().
		Model((*SpanIndex)(nil)).
		Distinct().
		ColumnExpr("arrayJoin(all_keys) AS key").
		Where("`span.time` >= ?", time.Now().Add(-24*time.Hour)).
		OrderExpr("key ASC")

	if projectID := h.tempoProjectID(req); projectID != 0 {
		q = q.Where("project_id = ?", projectID)
	}

	keys := make([]string, 0)

	if err := q.ScanColumns(ctx, &keys); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"tagNames": keys,
	})
}

func (h *TempoHandler) TagValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	tag := tempoAttrKey(req.Param("tag"))

	q := h.CH().NewSelect().
		Distinct().
		ColumnExpr("toString(?) AS value", chColumn(tag)).
		Model((*SpanIndex)(nil)).
		Where("`span.time` >= ?", time.Now().Add(-24*time.Hour)).
		OrderExpr("value ASC").
		Limit(1000)

	if projectID := h.tempoProjectID(req); projectID != 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if !strings.HasPrefix(tag, "span.") {
		q = q.Where("has(all_keys, ?)", tag)
	}

	values := make([]string, 0)

	if err := q.ScanColumns(ctx, &values); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"tagValues": values,
	})
}

// https://grafana.com/docs/tempo/latest/api_docs/#search
type TempoSearchParams struct {
	Start time.Time
	End   time.Time

	MinDuration time.Duration `urlstruct:"minDuration"`
	MaxDuration time.Duration `urlstruct:"maxDuration"`

	Tags  string // logfmt
	Limit int
}

type TempoSearchItem struct {
	TraceID           uuid.UUID `json:"traceID"`
	RootServiceName   string    `json:"rootServiceName" ch:",lc"`
	RootTraceName     string    `json:"rootTraceName" ch:",lc"`
	StartTimeUnixNano int64     `json:"startTimeUnixNano,string"`
	DurationMs        float64   `json:"durationMs"`
}

func (h *TempoHandler) Search(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(TempoSearchParams)
	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	if f.Start.IsZero() {
		f.Start = time.Now().Add(-time.Hour)
	}
	if f.Limit == 0 {
		f.Limit = 20
	}

	q := h.CH().NewSelect().
		ColumnExpr("`span.trace_id` AS trace_id").
		ColumnExpr("`service.name` AS root_service_name").
		ColumnExpr("`span.name` AS root_trace_name").
		ColumnExpr("toInt64(toUnixTimestamp(`span.time`) * 1e9) AS start_time_unix_nano").
		ColumnExpr("`span.duration` / 1e6 AS duration_ms").
		Model((*SpanIndex)(nil)).
		Where("`span.time` >= ?", f.Start).
		Where("`span.parent_id` = 0").
		Limit(f.Limit)

	if projectID := h.tempoProjectID(req); projectID != 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if !f.End.IsZero() {
		q = q.Where("`span.time` <= ?", f.End)
	}
	if f.MinDuration != 0 {
		q = q.Where("`span.duration` >= ?", int64(f.MinDuration))
	}
	if f.MaxDuration != 0 {
		q = q.Where("`span.duration` <= ?", int64(f.MaxDuration))
	}

	if f.Tags != "" {
		d := logfmt.NewDecoder(strings.NewReader(f.Tags))
		for d.ScanRecord() {
			for d.ScanKeyval() {
				key := tempoAttrKey(string(d.Key()))
				value := string(d.Value())

				var b []byte
				b = appendCHColumn(b, key)
				b = append(b, " = "...)
				b = chschema.AppendString(b, value)
				q = q.Where(string(b))
			}
		}
		if err := d.Err(); err != nil {
			return err
		}
	}

	traces := make([]TempoSearchItem, 0)

	if err := q.Scan(ctx, &traces); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"traces": traces,
		"metrics": bunrouter.H{
			"inspectedTraces": 0,
			"inspectedBytes":  0,
			"inspectedBlocks": 0,
		},
	})
}

func (h *TempoHandler) tempoProjectID(req bunrouter.Request) uint32 {
	s := req.Header.Get(projectIDHeaderKey)
	if s == "" {
		return 0
	}
	projectID, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		h.Zap(req.Context()).Error("can't parse project id", zap.Error(err))
		return 0
	}
	return uint32(projectID)
}

func tempoAttrKey(key string) string {
	switch key {
	case "name":
		return xattr.SpanName
	default:
		return key
	}
}
