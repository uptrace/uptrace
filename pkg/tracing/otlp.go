package tracing

import (
	"time"

	"github.com/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func initSpanFromOTLP(dest *Span, resource AttrMap, src *tracepb.Span) {
	dest.ID = idgen.SpanIDFromBytes(src.SpanId)
	dest.ParentID = idgen.SpanIDFromBytes(src.ParentSpanId)
	dest.TraceID = idgen.TraceIDFromBytes(src.TraceId)
	dest.Name = src.Name
	dest.Kind = otlpSpanKind(src.Kind)

	dest.Time = time.Unix(0, int64(src.StartTimeUnixNano))
	dest.Duration = time.Duration(src.EndTimeUnixNano - src.StartTimeUnixNano)

	if src.Status != nil {
		dest.StatusCode = otlpStatusCode(src.Status.Code)
		dest.StatusMessage = src.Status.Message
	}

	dest.Attrs = make(AttrMap, len(resource)+len(src.Attributes))
	for k, v := range resource {
		dest.Attrs[k] = v
	}
	otlpconv.ForEachKeyValue(src.Attributes, func(key string, value any) {
		dest.Attrs[key] = value
	})

	dest.Events = make([]*SpanEvent, len(src.Events))
	for i, event := range src.Events {
		dest.Events[i] = newSpanFromOTLPEvent(event)
	}

	dest.Links = make([]*SpanLink, len(src.Links))
	for i, link := range src.Links {
		dest.Links[i] = newSpanLink(link)
	}
}

func newSpanFromOTLPEvent(src *tracepb.Span_Event) *SpanEvent {
	dest := new(SpanEvent)
	dest.Name = src.Name
	dest.Time = time.Unix(0, int64(src.TimeUnixNano))

	dest.Attrs = make(AttrMap, len(src.Attributes))
	otlpconv.ForEachKeyValue(src.Attributes, func(key string, value any) {
		dest.Attrs[key] = value
	})

	return dest
}

func otlpSpanKind(kind tracepb.Span_SpanKind) string {
	switch kind {
	case tracepb.Span_SPAN_KIND_SERVER:
		return ServerSpanKind
	case tracepb.Span_SPAN_KIND_CLIENT:
		return ClientSpanKind
	case tracepb.Span_SPAN_KIND_PRODUCER:
		return ProducerSpanKind
	case tracepb.Span_SPAN_KIND_CONSUMER:
		return ConsumerSpanKind
	}
	return InternalSpanKind
}

func otlpStatusCode(code tracepb.Status_StatusCode) string {
	switch code {
	case tracepb.Status_STATUS_CODE_ERROR:
		return ErrorStatusCode
	default:
		return OKStatusCode
	}
}
