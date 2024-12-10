package tracing

import (
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type PublicSpan struct {
	ID         idgen.SpanID  `json:"id"`
	ParentID   idgen.SpanID  `json:"parentId"`
	TraceID    idgen.TraceID `json:"traceId"`
	Standalone bool          `json:"standalone,omitempty"`

	Type      string `json:"type"`
	System    string `json:"system"`
	Operation string `json:"-"`
	GroupID   uint64 `json:"groupId,string"`
	Kind      string `json:"kind"`

	Name        string `json:"name"`
	EventName   string `json:"eventName,omitempty"`
	DisplayName string `json:"displayName"`

	Time     time.Time `json:"time"`
	Duration int64     `json:"duration"`

	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage,omitempty"`

	Attrs  AttrMap      `json:"attrs"`
	Events []*SpanEvent `json:"events"`
	Links  []*SpanLink  `json:"links"`
}

func convPublicSpan(dest *PublicSpan, src *Span) {
	dest.ID = src.ID
	dest.ParentID = src.ParentID
	dest.TraceID = src.TraceID
	dest.Standalone = src.Standalone

	dest.Type = src.Type
	dest.System = src.System
	dest.GroupID = src.GroupID
	dest.Kind = src.Kind

	dest.Name = src.Name
	dest.EventName = src.EventName
	dest.DisplayName = src.DisplayName

	dest.Time = src.Time
	dest.Duration = src.Duration.Microseconds()

	dest.StatusCode = src.StatusCode
	dest.StatusMessage = src.StatusMessage

	dest.Attrs = src.Attrs
	dest.Events = src.Events
	dest.Links = src.Links
}

type PublicSpanFilter struct {
	urlstruct.Pager
	org.TimeFilter

	ProjectID uint32 `urlstruct:"-"`
	TraceID   idgen.TraceID
	ID        uint64
	ParentID  uint64
}

type PublicHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	CH     *ch.DB
}

type PublicHandler struct {
	*PublicHandlerParams
}

func NewPublicHandler(p PublicHandlerParams) *PublicHandler {
	return &PublicHandler{&p}
}

func registerPublicHandler(h *PublicHandler, p bunapp.RouterParams, m *org.Middleware) {
	p.RouterPublicV1.
		Use(m.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			g.GET("/spans", h.Spans)
			g.GET("/groups", h.Groups)
		})
}

func (h *PublicHandler) Spans(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	f := new(PublicSpanFilter)

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}
	f.ProjectID = project.ID

	f.Pager.DefaultLimit = 100
	f.Pager.MaxLimit = 100000
	limit := f.Pager.GetLimit()

	var spansData []SpanData
	q := h.CH.NewSelect().
		Model(&spansData).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		Limit(limit)

	if !f.TraceID.IsZero() {
		q = q.Where("trace_id = ?", f.TraceID)
	}
	if f.ID != 0 {
		q = q.Where("id = ?", f.ID)
	}
	if f.ParentID != 0 {
		q = q.Where("id = ?", f.ParentID)
	}

	if err := q.Scan(ctx, &spansData); err != nil {
		return err
	}

	spans := make([]PublicSpan, len(spansData))

	for i := range spansData {
		src, err := spansData[i].FilledSpan()
		if err != nil {
			return err
		}
		convPublicSpan(&spans[i], src)
	}

	return httputil.JSON(w, bunrouter.H{
		"spans":   spans,
		"hasMore": len(spans) == limit,
	})
}

func (h *PublicHandler) Groups(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(SpanFilter)
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	f.Pager.DefaultLimit = 10000
	f.Pager.MaxLimit = 100000
	limit := f.Pager.GetLimit()

	selq, _ := BuildSpanIndexQuery(h.CH, f, 0)

	items := make([]map[string]any, 0)
	if err := selq.
		Apply(f.CHOrder).
		Limit(limit).
		Scan(ctx, &items); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"groups":  items,
		"hasMore": len(items) == limit,
	})
}
