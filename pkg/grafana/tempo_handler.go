package grafana

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/grafana/tempo/pkg/traceql"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

const tempoDefaultPeriod = time.Hour

var jsonMarshaler = &jsonpb.Marshaler{}

type TempoHandler struct {
	BaseGrafanaHandler
}

func NewTempoHandler(p BaseGrafanaHandlerParams) *TempoHandler {
	return &TempoHandler{
		BaseGrafanaHandler: BaseGrafanaHandler{&p},
	}
}

func registerTempoHandler(h *TempoHandler, p bunapp.RouterParams, m *org.Middleware) {
	// https://grafana.com/docs/tempo/latest/api_docs/
	p.Router.WithGroup("/api/tempo/:project_id", func(g *bunrouter.Group) {
		g = g.Use(m.UserAndProject)

		g.GET("/ready", h.Ready)
		g.GET("/api/echo", h.Echo)
		g.GET("/api/status/buildinfo", h.BuildInfo)

		g.GET("/api/traces/:trace_id", h.QueryTrace)
		g.GET("/api/traces/:trace_id/json", h.QueryTraceJSON)

		g.GET("/api/search", h.Search)

		g.GET("/api/v2/search/tags", h.Tags)
		g.GET("/api/v2/search/tag/:tag/values", h.TagValues)
	})
}

func (h *TempoHandler) BuildInfo(w http.ResponseWriter, req bunrouter.Request) error {
	return httputil.JSON(w, bunrouter.H{})
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

	traceID, err := idgen.ParseTraceID(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans, _, err := tracing.SelectTraceSpans(ctx, h.CH, traceID)
	if err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	resp := newTempopbTrace(h.Conf, traceID, spans)

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

type Scope struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (h *TempoHandler) Tags(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	keys := make([]string, 0)

	if err := tracing.NewSpanIndexQuery(h.CH).
		Distinct().
		ColumnExpr("arrayJoin(all_keys) AS key").
		Where("project_id = ?", project.ID).
		Where("time >= ?", time.Now().Add(-tempoDefaultPeriod)).
		OrderExpr("key ASC").
		ScanColumns(ctx, &keys); err != nil {
		return err
	}

	scopes := []Scope{
		{
			Name: "span",
			Tags: keys,
		},
		{
			Name: "resource",
			Tags: []string{},
		},
		{
			Name: "intrinsic",
			Tags: []string{},
		},
	}

	return httputil.JSON(w, bunrouter.H{
		"scopes": scopes,
	})
}

type TagValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (h *TempoHandler) TagValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	tag := tempoAttrKey(req.Param("tag"))
	tagCHExpr := tempoCHExpr(tag)

	q := tracing.NewSpanIndexQuery(h.CH).
		Distinct().
		ColumnExpr("toString(?) AS value", tagCHExpr).
		Where("project_id = ?", project.ID).
		Where("time >= ?", time.Now().Add(-tempoDefaultPeriod)).
		OrderExpr("value ASC").
		Limit(1000)

	if tracing.IsIndexedAttr(tag) {
		q = q.Where("?0 != defaultValueOfArgumentType(?0)", tagCHExpr)
	} else {
		q = q.Where("has(all_keys, ?)", tag)
	}

	values := make([]string, 0)

	if err := q.ScanColumns(ctx, &values); err != nil {
		return err
	}

	tagValues := make([]TagValue, len(values))
	for i, value := range values {
		tagValues[i] = TagValue{
			Type:  "string",
			Value: value,
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"tagValues": tagValues,
	})
}

// https://grafana.com/docs/tempo/latest/api_docs/#search
type TempoSearchParams struct {
	Start time.Time
	End   time.Time

	MinDuration time.Duration `urlstruct:"minDuration"`
	MaxDuration time.Duration `urlstruct:"maxDuration"`

	Q     string
	Limit int
}

type TempoSearchTrace struct {
	TraceID           idgen.TraceID        `json:"traceID"`
	RootServiceName   string               `json:"rootServiceName" ch:",lc"`
	RootTraceName     string               `json:"rootTraceName" ch:",lc"`
	StartTimeUnixNano int64                `json:"startTimeUnixNano,string"`
	DurationMs        float64              `json:"durationMs"`
	SpanSets          []TempoSearchSpanSet `json:"spanSets"`
}

type TempoSearchSpanSet struct {
	Matched int               `json:"matched"`
	Spans   []TempoSearchSpan `json:"spans"`
}

type TempoSearchSpan struct {
	SpanID            idgen.SpanID      `json:"spanID"`
	StartTimeUnixNano int64             `json:"startTimeUnixNano,string"`
	DurationNanos     int64             `json:"durationNanos"`
	Attributes        []TempoSearchAttr `json:"attributes"`
}

type TempoSearchAttr struct {
	Key   string               `json:"key"`
	Value TempoSearchAttrValue `json:"value"`
}

type TempoSearchAttrValue struct {
	StringValue string `json:"stringValue"`
}

type TraceSpanID struct {
	TraceID idgen.TraceID
	ID      idgen.SpanID
}

func (h *TempoHandler) Search(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

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

	q := tracing.NewSpanIndexQuery(h.CH).
		DistinctOn("trace_id").
		ColumnExpr("trace_id").
		ColumnExpr("id").
		Where("project_id = ?", project.ID).
		Where("time >= ?", f.Start).
		Limit(f.Limit)

	if !f.End.IsZero() {
		q = q.Where("time <= ?", f.End)
	}
	if f.MinDuration != 0 {
		q = q.Where("duration >= ?", int64(f.MinDuration))
	}
	if f.MaxDuration != 0 {
		q = q.Where("duration <= ?", int64(f.MaxDuration))
	}

	_, fetchReq, err := traceql.NewEngine().Compile(f.Q)
	if err != nil {
		return err
	}

	if err := applyTraceql(q, fetchReq); err != nil {
		return err
	}

	found := make([]*TraceSpanID, 0)
	if err := q.Scan(ctx, &found); err != nil {
		return err
	}

	traces := make([]*TempoSearchTrace, 0, len(found))

	for _, item := range found {
		var data []*tracing.SpanData

		if err := h.CH.NewSelect().
			DistinctOn("id").
			ColumnExpr("trace_id, id, parent_id, time, data").
			Model(&data).
			Column("data").
			Where("trace_id = ?", item.TraceID).
			Where("id = ? OR parent_id = 0", item.ID).
			Limit(2).
			Scan(ctx); err != nil {
			return err
		}

		var root, span *tracing.Span
		for _, data := range data {
			dest := new(tracing.Span)
			if err := data.Decode(dest); err != nil {
				return err
			}

			if dest.ParentID == 0 {
				root = dest
			}
			if dest.ID == item.ID {
				span = dest
			}
		}

		if root == nil || span == nil {
			continue
		}

		attrs := make([]TempoSearchAttr, 0, len(span.Attrs))
		for key, val := range span.Attrs {
			attrs = append(attrs, TempoSearchAttr{
				Key:   key,
				Value: TempoSearchAttrValue{StringValue: fmt.Sprint(val)},
			})
		}

		traces = append(traces, &TempoSearchTrace{
			TraceID:           root.TraceID,
			RootServiceName:   root.Attrs.Text(attrkey.ServiceName),
			RootTraceName:     root.DisplayName,
			StartTimeUnixNano: root.Time.UnixNano(),
			DurationMs:        float64(root.Duration) / float64(time.Millisecond),

			SpanSets: []TempoSearchSpanSet{{
				Matched: 1,
				Spans: []TempoSearchSpan{{
					SpanID:            span.ID,
					StartTimeUnixNano: span.Time.UnixNano(),
					DurationNanos:     int64(span.Duration),
					Attributes:        attrs,
				}},
			}},
		})
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

func tempoCHExpr(attrKey string) ch.Safe {
	return ch.Safe(tracing.AppendCHAttr(nil, tql.Attr{Name: attrKey}))
}

func tempoAttrKey(attrKey string) string {
	switch attrKey {
	case "name":
		return attrkey.SpanName
	case "status":
		return attrkey.SpanStatusCode
	case "statusMessage":
		return attrkey.SpanStatusMessage
	case "kind":
		return attrkey.SpanKind
	}

	switch {
	case strings.HasPrefix(attrKey, "span."):
		attrKey = strings.TrimPrefix(attrKey, "span.")
	case strings.HasPrefix(attrKey, "resource."):
		attrKey = strings.TrimPrefix(attrKey, "resource.")
	}

	return attrkey.Clean(attrKey)
}

//------------------------------------------------------------------------------

func applyTraceql(q *ch.SelectQuery, fetchReq *traceql.FetchSpansRequest) error {
	if fetchReq.StartTimeUnixNanos != 0 {
		q = q.Where("time >= ?", fetchReq.StartTimeUnixNanos)
	}
	if fetchReq.StartTimeUnixNanos != 0 {
		q = q.Where("time <= ?", fetchReq.EndTimeUnixNanos)
	}
	for i := range fetchReq.Conditions {
		cond := &fetchReq.Conditions[i]

		filter, err := convTraceqlCondition(q, cond)
		if err != nil {
			return err
		}

		if filter == nil {
			continue
		}

		b, err := tracing.AppendFilter(*filter, 0)
		if err != nil {
			return err
		}
		q = q.Where(string(b))
	}

	return nil
}

func convTraceqlCondition(q *ch.SelectQuery, cond *traceql.Condition) (*tql.Filter, error) {
	expr := tempoFilterExpr(cond)
	if expr == nil {
		return nil, nil
	}

	value := tempoFilterValue(cond)
	if value == nil {
		return nil, nil
	}

	filterOp := tempoFilterOp(cond)
	if filterOp == "" {
		return nil, nil
	}

	return &tql.Filter{
		LHS: expr,
		Op:  filterOp,
		RHS: value,
	}, nil
}

func tempoFilterExpr(cond *traceql.Condition) tql.Expr {
	attrKey := tempoAttrName(&cond.Attribute)
	return tql.Attr{Name: attrKey}
}

func tempoAttrName(attr *traceql.Attribute) string {
	switch attr.Intrinsic {
	case traceql.IntrinsicDuration:
		return attrkey.SpanDuration
	case traceql.IntrinsicName:
		return attrkey.SpanName
	case traceql.IntrinsicStatus:
		return attrkey.SpanStatusCode
	case traceql.IntrinsicStatusMessage:
		return attrkey.SpanStatusMessage
	case traceql.IntrinsicKind:
		return attrkey.SpanKind
	}
	return attrkey.Clean(attr.Name)
}

func tempoFilterValue(cond *traceql.Condition) tql.Value {
	if len(cond.Operands) == 0 {
		return nil
	}

	if len(cond.Operands) == 1 {
		return _tempoFilterValue(&cond.Operands[0])
	}

	value := tql.StringValues{}
	for i := range cond.Operands {
		tmp := _tempoFilterValue(&cond.Operands[i])
		value.Strings = append(value.Strings, tmp.String())
	}
	return value
}

func _tempoFilterValue(static *traceql.Static) tql.Value {
	switch static.Type {
	case traceql.TypeInt:
		return tql.NumberValue{
			Kind: tql.NumberUnitless,
			Text: strconv.Itoa(static.N),
		}
	case traceql.TypeFloat:
		return tql.NumberValue{
			Kind: tql.NumberUnitless,
			Text: strconv.FormatFloat(static.F, 'f', -1, 64),
		}
	case traceql.TypeString:
		return tql.StringValue{Text: static.S}
	case traceql.TypeBoolean:
		return tql.NumberValue{
			Kind: tql.NumberUnitless,
			Text: strconv.FormatBool(static.B),
		}
	case traceql.TypeDuration:
		return tql.NumberValue{
			Kind: tql.NumberDuration,
			Text: static.D.String(),
		}
	case traceql.TypeStatus:
		return tql.StringValue{Text: static.Status.String()}
	case traceql.TypeKind:
		return tql.StringValue{Text: static.Kind.String()}
	default:
		return nil
	}
}

func tempoFilterOp(cond *traceql.Condition) tql.FilterOp {
	switch cond.Op {
	case traceql.OpEqual:
		if len(cond.Operands) > 1 {
			return tql.FilterIn
		}
		return tql.FilterEqual
	case traceql.OpNotEqual:
		if len(cond.Operands) > 1 {
			return tql.FilterNotIn
		}
		return tql.FilterNotEqual
	case traceql.OpRegex:
		return tql.FilterRegexp
	case traceql.OpNotRegex:
		return tql.FilterNotRegexp
	case traceql.OpGreater:
		return ">"
	case traceql.OpGreaterEqual:
		return ">="
	case traceql.OpLess:
		return "<"
	case traceql.OpLessEqual:
		return "<="
	default:
		return ""
	}
}
