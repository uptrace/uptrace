package tracing

import (
	"context"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/sqlparser"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TraceServiceServer struct {
	collectortrace.UnimplementedTraceServiceServer

	*bunapp.App

	batchSize int
	ch        chan otlpSpan
	gate      *syncutil.Gate
}

var _ collectortrace.TraceServiceServer = (*TraceServiceServer)(nil)

func NewTraceServiceServer(app *bunapp.App) *TraceServiceServer {
	batchSize := scaleWithCPU(2000, 32000)
	s := &TraceServiceServer{
		App: app,

		batchSize: batchSize,
		ch:        make(chan otlpSpan, batchSize),
		gate:      syncutil.NewGate(runtime.GOMAXPROCS(0)),
	}

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		s.processLoop(app.Context())
	}()

	return s
}

func (s *TraceServiceServer) Export(
	ctx context.Context, req *collectortrace.ExportTraceServiceRequest,
) (*collectortrace.ExportTraceServiceResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
	}
	s.process(req.ResourceSpans)
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func (s *TraceServiceServer) process(resourceSpans []*tracepb.ResourceSpans) {
	for _, rss := range resourceSpans {
		resource := otlpAttrs(rss.Resource.Attributes)

		for _, ils := range rss.InstrumentationLibrarySpans {
			lib := ils.InstrumentationLibrary
			resource[xattr.OtelLibraryName] = lib.Name
			if lib.Version != "" {
				resource[xattr.OtelLibraryVersion] = lib.Version
			}

			for _, span := range ils.Spans {
				s.ch <- otlpSpan{
					Span:     span,
					resource: resource,
				}
			}
		}
	}
}

func (s *TraceServiceServer) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]otlpSpan, 0, s.batchSize)

loop:
	for {
		select {
		case span := <-s.ch:
			spans = append(spans, span)
		case <-timer.C:
			if len(spans) > 0 {
				s.flushSpans(ctx, spans)
				spans = make([]otlpSpan, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if len(spans) == s.batchSize {
			s.flushSpans(ctx, spans)
			spans = make([]otlpSpan, 0, len(spans))
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans)
	}
}

func (s *TraceServiceServer) flushSpans(ctx context.Context, spans []otlpSpan) {
	ctx, span := bunapp.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		index := make([]*SpanIndex, len(spans))
		data := make([]*SpanData, len(spans))

		for i, span := range spans {
			spanIndex := s.newSpanIndex(span)
			index[i] = spanIndex
			data[i] = s.newSpanData(span, spanIndex)
		}

		if _, err := s.CH().NewInsert().Model(&data).Exec(ctx); err != nil {
			s.Zap(ctx).Error("ch.Insert failed",
				zap.Error(err), zap.String("table", "spans_data"))
		}
		if _, err := s.CH().NewInsert().Model(&index).Exec(ctx); err != nil {
			s.Zap(ctx).Error("ch.Insert failed",
				zap.Error(err), zap.String("table", "spans_index"))
		}
	}()
}

func (s *TraceServiceServer) newSpanIndex(span otlpSpan) *SpanIndex {
	index := new(SpanIndex)

	index.ID = otlpSpanID(span.SpanId)
	index.ParentID = otlpSpanID(span.ParentSpanId)
	index.TraceID = otlpTraceID(span.TraceId)
	index.Name = span.Name
	index.Kind = otlpSpanKind(span.Kind)

	index.Time = time.Unix(0, int64(span.StartTimeUnixNano))
	index.Duration = time.Duration(span.EndTimeUnixNano - span.StartTimeUnixNano)

	if span.Status != nil {
		index.StatusCode = otlpStatusCode(span.Status.Code)
		index.StatusMessage = span.Status.Message
	}

	index.Attrs = make(AttrMap, len(span.resource)+len(span.Attributes))
	for k, v := range span.resource {
		index.Attrs[k] = v
	}
	otlpSetAttrs(index.Attrs, span.Attributes)

	index.AttrKeys, index.AttrValues = attrKeysAndValues(index.Attrs)
	index.ServiceName = index.Attrs.Text(xattr.ServiceName)
	index.HostName = index.Attrs.Text(xattr.HostName)

	digest := xxhash.New()
	digest.WriteString(index.Kind)
	digest.WriteString(index.Name)
	assignSystemAndGroupID(index, digest)
	index.GroupID = digest.Sum64()

	return index
}

func attrKeysAndValues(m AttrMap) ([]string, []string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, truncate(asString(v), 200))
	}
	return keys, values
}

func (s *TraceServiceServer) newSpanData(span otlpSpan, index *SpanIndex) *SpanData {
	attrs := index.Attrs.Clone()

	attrs[xattr.SpanSystem] = index.System
	attrs[xattr.SpanGroupID] = index.GroupID

	attrs[xattr.SpanName] = index.Name
	attrs[xattr.SpanKind] = index.Kind
	attrs[xattr.SpanTime] = index.Time
	attrs[xattr.SpanDuration] = index.Duration

	attrs[xattr.SpanStatusCode] = index.StatusCode
	if index.StatusMessage != "" {
		attrs[xattr.SpanStatusMessage] = index.StatusMessage
	}

	data := new(SpanData)
	data.TraceID = index.TraceID
	data.ID = index.ID
	data.ParentID = index.ParentID
	data.Time = index.Time
	data.Attrs = attrs

	data.Events = make([]*SpanEvent, len(span.Events))
	for i, event := range span.Events {
		data.Events[i] = s.newSpanEvent(index, event)
	}

	data.Links = make([]*SpanLink, len(span.Links))
	for i, link := range span.Links {
		data.Links[i] = s.newSpanLink(link)
	}

	return data
}

func (s *TraceServiceServer) newSpanEvent(span *SpanIndex, in *tracepb.Span_Event) *SpanEvent {
	event := &SpanEvent{
		Name:  in.Name,
		Attrs: otlpAttrs(in.Attributes),
		Time:  time.Unix(0, int64(in.TimeUnixNano)),
	}
	if s := eventName(span, event); s != "" {
		event.Name = s
	}
	return event
}

func eventName(span *SpanIndex, event *SpanEvent) string {
	switch event.Name {
	case logEventType:
		if msg := event.Attrs.Text(xattr.LogMessage); msg != "" {
			if sev := event.Attrs.Text(xattr.LogSeverity); sev != "" {
				return sev + " " + msg
			}
			return msg
		}

		typ := event.Attrs.Text(xattr.ExceptionType)
		msg := event.Attrs.Text(xattr.ExceptionMessage)
		if typ != "" || msg != "" {
			return joinTypeMessage(typ, msg)
		}
	case exceptionEventType:
		return joinTypeMessage(
			event.Attrs.Text(xattr.ExceptionType),
			event.Attrs.Text(xattr.ExceptionMessage),
		)
	case messageEventType:
		if op := event.Attrs.Text(xattr.MessagingOperation); op != "" {
			return span.Name + " " + op
		}
		if typ := event.Attrs.Text("message.type"); typ != "" {
			return span.Name + " " + typ
		}
	}
	return ""
}

func (s *TraceServiceServer) newSpanLink(link *tracepb.Span_Link) *SpanLink {
	return &SpanLink{
		TraceID: otlpTraceID(link.TraceId),
		SpanID:  otlpSpanID(link.SpanId),
		Attrs:   otlpAttrs(link.Attributes),
	}
}

type otlpSpan struct {
	*tracepb.Span
	resource AttrMap
}

const (
	allSpanType      = "all"
	internalSpanType = "internal"

	httpSpanType      = "http"
	dbSpanType        = "db"
	rpcSpanType       = "rpc"
	messagingSpanType = "messaging"
	serviceSpanType   = "service"

	logEventType       = "log"
	exceptionEventType = "exception"
	messageEventType   = "message"
	eventType          = "event"
)

func assignSystemAndGroupID(span *SpanIndex, digest *xxhash.Digest) {
	if s := span.Attrs.Text(xattr.RPCSystem); s != "" {
		span.System = rpcSpanType + ":" + span.Attrs.ServiceName()
		digest.WriteString(span.System)
		return
	}

	if s := span.Attrs.Text(xattr.MessagingSystem); s != "" {
		span.System = messagingSpanType + ":" + s
		digest.WriteString(span.System)
		return
	}

	if s := span.Attrs.Text(xattr.DBSystem); s != "" {
		span.System = dbSpanType + ":" + s
		digest.WriteString(span.System)

		if s := span.Attrs.Text(xattr.DBSqlTable); s != "" {
			digest.WriteString(s)
		}
		if s := span.Attrs.Text(xattr.DBStatement); s != "" {
			span.Name = s
			hashDBStmt(digest, s)
		}

		return
	}

	if span.Attrs.Has(xattr.HTTPRoute) || span.Attrs.Has(xattr.HTTPTarget) {
		span.System = httpSpanType + ":" + span.Attrs.ServiceName()
		digest.WriteString(span.System)
		return
	}

	if span.ParentID == 0 || span.Kind != internalSpanKind {
		span.System = serviceSpanType + ":" + span.Attrs.ServiceName()
		digest.WriteString(span.System)
		return
	}

	span.System = internalSpanType
	digest.WriteString(span.System)
}

func hashDBStmt(digest *xxhash.Digest, s string) {
	tok := sqlparser.NewTokenizer(s)
	for {
		token, err := tok.NextToken()
		if err == io.EOF {
			break
		}
		if token.Type == sqlparser.IdentToken && isSQLKeyword(token.Text) {
			digest.WriteString(token.Text)
		}
	}
}

func isSQLKeyword(s string) bool {
	switch strings.ToUpper(s) {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "TRUNCATE",
		"WITH", "FROM", "TABLE", "JOIN", "UNION", "WHERE", "GROUP", "LIMIT", "ORDER", "HAVING":
		return true
	default:
		return false
	}
}

func joinTypeMessage(typ, msg string) string {
	if msg == "" {
		if typ == "" {
			return ""
		}
		return typ
	}
	if strings.HasPrefix(msg, typ) {
		return msg
	}
	return typ + ": " + msg
}
