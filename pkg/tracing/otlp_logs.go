package tracing

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type LogsServiceServer struct {
	collectorlogspb.UnimplementedLogsServiceServer

	*bunapp.App

	sp *SpanProcessor
}

var _ collectorlogspb.LogsServiceServer = (*LogsServiceServer)(nil)

func NewLogsServiceServer(app *bunapp.App, sp *SpanProcessor) *LogsServiceServer {
	return &LogsServiceServer{
		App: app,
		sp:  sp,
	}
}

func (s *LogsServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		logsReq := new(collectorlogspb.ExportLogsServiceRequest)
		if err := protojson.Unmarshal(body, logsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, logsReq.ResourceLogs, project)
		if err != nil {
			return err
		}

		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	case xprotobufContentType, protobufContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		logsReq := new(collectorlogspb.ExportLogsServiceRequest)
		if err := proto.Unmarshal(body, logsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, logsReq.ResourceLogs, project)
		if err != nil {
			return err
		}

		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported content type: %q", contentType)
	}
}

func (s *LogsServiceServer) Export(
	ctx context.Context, req *collectorlogspb.ExportLogsServiceRequest,
) (*collectorlogspb.ExportLogsServiceResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
	}

	dsn, err := org.DSNFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn)
	if err != nil {
		return nil, err
	}

	return s.export(ctx, req.ResourceLogs, project)
}

func (s *LogsServiceServer) export(
	ctx context.Context, resourceLogs []*logspb.ResourceLogs, project *org.Project,
) (*collectorlogspb.ExportLogsServiceResponse, error) {
	p := new(otlpLogProcessor)
	for _, rl := range resourceLogs {
		var resource AttrMap
		if rl.Resource != nil {
			resource = AttrMap(otlpconv.Map(rl.Resource.Attributes))
		}

		for _, sl := range rl.ScopeLogs {
			var scope AttrMap
			if sl.Scope != nil {
				scope = maps.Clone(resource)
				if sl.Scope.Name != "" {
					scope[attrkey.OtelLibraryName] = sl.Scope.Name
				}
				if sl.Scope.Version != "" {
					scope[attrkey.OtelLibraryVersion] = sl.Scope.Version
				}
				otlpconv.ForEachKeyValue(sl.Scope.Attributes, func(key string, value any) {
					scope[key] = value
				})
			} else {
				scope = resource
			}

			for _, lr := range sl.LogRecords {
				span := p.processLogRecord(scope, lr)
				span.ProjectID = project.ID
				s.sp.AddSpan(ctx, span)
			}
		}
	}
	return &collectorlogspb.ExportLogsServiceResponse{}, nil
}

//-----------------------------------------------------------------------------------------

type otlpLogProcessor struct {
	baseLogProcessor
}

func (p *otlpLogProcessor) processLogRecord(resource AttrMap, lr *logspb.LogRecord) *Span {
	span := new(Span)

	span.ID = rand.Uint64()
	span.ParentID = otlpSpanID(lr.SpanId)
	if lr.TraceId != nil {
		span.TraceID = otlpTraceID(lr.TraceId)
	}

	span.EventName = otelEventLog
	span.Kind = InternalSpanKind
	span.Time = time.Unix(0, int64(minNonZero(lr.TimeUnixNano, lr.ObservedTimeUnixNano)))

	span.Attrs = make(AttrMap, len(resource)+len(lr.Attributes)+1)
	span.Attrs.Merge(resource)
	otlpconv.ForEachKeyValue(lr.Attributes, func(key string, value any) {
		span.Attrs[key] = value
	})

	if lr.SeverityText != "" {
		span.Attrs[attrkey.LogSeverity] = lr.SeverityText
	}
	if lr.Body.Value != nil {
		p.processLogRecordBody(span, lr.Body.Value)
	}

	if !span.Attrs.Has(attrkey.LogMessage) {
		if msg := popLogMessageParam(span.Attrs); msg != "" {
			span.Attrs[attrkey.LogMessage] = msg
		}
	}

	return span
}

func minNonZero(u1, u2 uint64) uint64 {
	if u1 != 0 && u2 != 0 {
		return min(u1, u2)
	}
	if u1 != 0 {
		return u1
	}
	return u2
}

func (p *otlpLogProcessor) processLogRecordBody(span *Span, bodyValue any) {
	switch v := bodyValue.(type) {
	case *commonpb.AnyValue_StringValue:
		msg := v.StringValue
		if params, ok := bunutil.IsJSON(msg); ok {
			p.parseJSONLogMessage(span, params)
		} else {
			span.Attrs[attrkey.LogMessage] = msg
		}
	case *commonpb.AnyValue_KvlistValue:
		params := otlpconv.Map(v.KvlistValue.Values)
		populateSpanFromParams(span, params)
	}
}

//------------------------------------------------------------------------------

type baseLogProcessor struct {
	buf []byte
}

func (p *baseLogProcessor) parseJSONLogMessage(span *Span, params AttrMap) {
	type kv struct {
		key string
		val any
	}

	keys := make([]string, 0, len(params))
	var renamed []kv

	for oldKey, val := range params {
		newKey := attrkey.Clean(oldKey)
		if newKey == "" {
			delete(params, oldKey)
			continue
		}

		if newKey != oldKey {
			delete(params, oldKey)
			renamed = append(renamed, kv{key: newKey, val: val})
		}
		keys = append(keys, newKey)
	}

	for i := range renamed {
		kv := &renamed[i]
		params[kv.key] = kv.val
	}

	msg := popLogMessageParam(params)

	if msg != "" {
		span.Attrs[attrkey.LogMessage] = msg
		populateSpanFromParams(span, params)
		return
	}

	if span.EventName != otelEventLog {
		return
	}

	slices.Sort(keys)
	buf := p.buf[:0]

	for i, key := range keys {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, key...)
		buf = append(buf, '=')
		buf = appendParamValue(buf, params[key])
	}

	span.Attrs[attrkey.LogMessage] = string(buf)
	p.buf = buf
}

func popLogMessageParam(params AttrMap) string {
	for _, key := range []string{"log", "message", "msg"} {
		if value, _ := params[key].(string); value != "" {
			delete(params, key)
			return value
		}
	}
	return ""
}

func appendParamValue(b []byte, val any) []byte {
	switch val := val.(type) {
	case string:
		if strings.IndexByte(val, ' ') == -1 && strings.IndexByte(val, '"') == -1 {
			return append(b, val...)
		}
		return strconv.AppendQuote(b, fmt.Sprint(val))
	case json.Number:
		return append(b, val...)
	case float64:
		return strconv.AppendFloat(b, val, 'f', -1, 64)
	case int64:
		return strconv.AppendInt(b, val, 10)
	case uint64:
		return strconv.AppendUint(b, val, 10)
	case bool:
		return strconv.AppendBool(b, val)
	case nil:
		return append(b, "<nil>"...)
	default:
		return strconv.AppendQuote(b, fmt.Sprint(val))
	}
}
