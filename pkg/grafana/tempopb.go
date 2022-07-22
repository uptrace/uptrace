package grafana

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/grafana/tempo/pkg/tempopb"
	commonpb "github.com/grafana/tempo/pkg/tempopb/common/v1"
	resourcepb "github.com/grafana/tempo/pkg/tempopb/resource/v1"
	tracepb "github.com/grafana/tempo/pkg/tempopb/trace/v1"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
	"github.com/uptrace/uptrace/pkg/uuid"
	"go.uber.org/zap"
)

func newTempopbTrace(app *bunapp.App, traceID uuid.UUID, spans []*tracing.Span) *tempopb.Trace {
	cfg := app.Config()
	backlink := &commonpb.KeyValue{
		Key: "uptrace.url",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: fmt.Sprintf(
					"%s://%s:%s/traces/%s",
					cfg.Site.Scheme, cfg.Site.Host, cfg.Listen.HTTPPort, traceID.String(),
				),
			},
		},
	}

	resourceSpans := make([]*tracepb.ResourceSpans, 0, len(spans))

	for _, span := range spans {
		tempoSpan := tempoResourceSpans(span, backlink)
		resourceSpans = append(resourceSpans, tempoSpan)
	}

	return &tempopb.Trace{
		Batches: resourceSpans,
	}
}

var tempoResourceKeys = []string{xattr.ServiceName, xattr.HostName}

func tempoResourceSpans(s *tracing.Span, backlink *commonpb.KeyValue) *tracepb.ResourceSpans {
	resource, attributes := tempoResourceAndAttributes(s.Attrs, tempoResourceKeys)
	attributes = append(attributes, backlink)

	tracepbSpan := newTracepbSpan(s)
	tracepbSpan.Attributes = attributes
	tracepbSpans := []*tracepb.Span{tracepbSpan}

	return &tracepb.ResourceSpans{
		// Grafana does not work without resource attributes.
		Resource: &resourcepb.Resource{
			Attributes: resource,
		},

		// ScopeSpans: []*tracepb.ScopeSpans{{
		// 	Scope: nil,
		// 	Spans: tracepbSpans,
		// }},

		// InstrumentationLibrarySpans field is deprecated in favor of ScopeSpans.
		// It will be removed eventually.
		InstrumentationLibrarySpans: []*tracepb.InstrumentationLibrarySpans{{
			InstrumentationLibrary: &commonpb.InstrumentationLibrary{},
			Spans:                  tracepbSpans,
		}},
	}
}

func newTracepbSpan(s *tracing.Span) *tracepb.Span {
	events := make([]*tracepb.Span_Event, len(s.Events))
	for i, event := range s.Events {
		events[i] = newTracepbSpanEvent(event)
	}

	links := make([]*tracepb.Span_Link, len(s.Links))
	for i, link := range s.Links {
		links[i] = newTracepbSpanLink(link)
	}

	out := &tracepb.Span{
		TraceId: s.TraceID[:],
		SpanId:  tempoSpanID(s.ID),

		Name:              s.Name,
		Kind:              tempoSpanKind(s.Kind),
		StartTimeUnixNano: uint64(s.Time.UnixNano()),
		EndTimeUnixNano:   uint64(s.Time.UnixNano()) + uint64(s.Duration),

		Events: events,
		Links:  links,
		Status: &tracepb.Status{
			Code:    tempoStatusCode(s.StatusCode),
			Message: s.StatusMessage,
		},
	}
	if s.ParentID != 0 {
		out.ParentSpanId = tempoSpanID(s.ParentID)
	}
	return out
}

func newTracepbSpanEvent(s *tracing.Span) *tracepb.Span_Event {
	return &tracepb.Span_Event{
		TimeUnixNano: uint64(s.Time.UnixNano()),
		Name:         s.Name,
		Attributes:   tempoAttributes(s.Attrs),
	}
}

func newTracepbSpanLink(l *tracing.SpanLink) *tracepb.Span_Link {
	return &tracepb.Span_Link{
		TraceId:    l.TraceID[:],
		SpanId:     tempoSpanID(l.SpanID),
		Attributes: tempoAttributes(l.Attrs),
	}
}

func tempoSpanID(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func tempoStatusCode(s string) tracepb.Status_StatusCode {
	switch s {
	case "ok":
		return tracepb.Status_STATUS_CODE_OK
	case "error":
		return tracepb.Status_STATUS_CODE_ERROR
	default:
		return tracepb.Status_STATUS_CODE_UNSET
	}
}

func tempoSpanKind(s string) tracepb.Span_SpanKind {
	switch s {
	case tracing.InternalSpanKind:
		return tracepb.Span_SPAN_KIND_INTERNAL
	case tracing.ServerSpanKind:
		return tracepb.Span_SPAN_KIND_SERVER
	case tracing.ClientSpanKind:
		return tracepb.Span_SPAN_KIND_CLIENT
	case tracing.ProducerSpanKind:
		return tracepb.Span_SPAN_KIND_PRODUCER
	case tracing.ConsumerSpanKind:
		return tracepb.Span_SPAN_KIND_CONSUMER
	default:
		return tracepb.Span_SPAN_KIND_UNSPECIFIED
	}
}

func tempoResourceAndAttributes(
	m xotel.AttrMap, resourceKeys []string,
) (resource, attrs []*commonpb.KeyValue) {
	isResource := make(map[string]bool, len(resourceKeys))
	for _, k := range resourceKeys {
		isResource[k] = true
	}

	resource = make([]*commonpb.KeyValue, 0, len(resourceKeys))
	attrs = make([]*commonpb.KeyValue, 0, len(m))

	for k, v := range m {
		av := tempoAnyValue(v)
		if av == nil {
			continue
		}

		if isResource[k] {
			resource = append(resource, &commonpb.KeyValue{
				Key:   k,
				Value: av,
			})
		} else {
			attrs = append(attrs, &commonpb.KeyValue{
				Key:   k,
				Value: av,
			})
		}
	}

	return resource, attrs
}

func tempoAttributes(m xotel.AttrMap) []*commonpb.KeyValue {
	kvs := make([]*commonpb.KeyValue, 0, len(m))
	for k, v := range m {
		if av := tempoAnyValue(v); av != nil {
			kvs = append(kvs, &commonpb.KeyValue{
				Key:   k,
				Value: av,
			})
		}
	}
	return kvs
}

func tempoAnyValue(v any) *commonpb.AnyValue {
	switch v := v.(type) {
	case string:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: v,
			},
		}
	case int64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: v,
			},
		}
	case uint64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: int64(v),
			},
		}
	case float64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_DoubleValue{
				DoubleValue: v,
			},
		}
	case bool:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BoolValue{
				BoolValue: v,
			},
		}
	// case []byte:
	// 	return &commonpb.AnyValue{
	// 		Value: &commonpb.AnyValue_BytesValue{
	// 			BytesValue: v,
	// 		},
	// 	}
	case []string:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = tempoAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	case []int64:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = tempoAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	case []float64:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = tempoAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	case []any:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = tempoAnyValue(fmt.Sprint(el))
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	default:
		otelzap.L().Error("unsupported attribute type",
			zap.String("type", reflect.TypeOf(v).String()))
		return nil
	}
}
