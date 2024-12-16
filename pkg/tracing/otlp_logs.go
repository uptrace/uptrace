package tracing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/norm"
)

type LogsServiceServer struct {
	collectorlogspb.UnimplementedLogsServiceServer

	PG       *bun.DB
	Projects *org.ProjectGateway
	sp       *SpanConsumer
}

var _ collectorlogspb.LogsServiceServer = (*LogsServiceServer)(nil)

func NewLogsServiceServer(pg *bun.DB, sp *SpanConsumer) *LogsServiceServer {
	return &LogsServiceServer{PG: pg, sp: sp}
}

func (s *LogsServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := s.Projects.SelectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := io.ReadAll(req.Body)
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
		body, err := io.ReadAll(req.Body)
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

	project, err := s.Projects.SelectByDSN(ctx, dsn)
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

	span.ID = idgen.RandSpanID()
	span.ParentID = idgen.SpanIDFromBytes(lr.SpanId)
	if lr.TraceId != nil {
		span.TraceID = idgen.TraceIDFromBytes(lr.TraceId)
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
	} else if lr.SeverityNumber > 0 {
		sev := p.severityFromNumber(int32(lr.SeverityNumber))
		span.Attrs[attrkey.LogSeverity] = sev
	}
	if lr.Body.Value != nil {
		p.processLogRecordBody(span, lr.Body.Value)
	}

	if !span.Attrs.Exists(attrkey.LogMessage) {
		if msg := popLogMessageParam(span.Attrs); msg != "" {
			span.Attrs[attrkey.LogMessage] = msg
		}
	}

	return span
}

func (p *otlpLogProcessor) severityFromNumber(num int32) string {
	switch num {
	case 1:
		return norm.SeverityTrace
	case 2:
		return norm.SeverityTrace2
	case 3:
		return norm.SeverityTrace3
	case 4:
		return norm.SeverityTrace4
	case 5:
		return norm.SeverityDebug
	case 6:
		return norm.SeverityDebug2
	case 7:
		return norm.SeverityDebug3
	case 8:
		return norm.SeverityDebug4
	case 9:
		return norm.SeverityInfo
	case 10:
		return norm.SeverityInfo2
	case 11:
		return norm.SeverityInfo3
	case 12:
		return norm.SeverityInfo4
	case 13:
		return norm.SeverityWarn
	case 14:
		return norm.SeverityWarn2
	case 15:
		return norm.SeverityWarn3
	case 16:
		return norm.SeverityWarn4
	case 17:
		return norm.SeverityError
	case 18:
		return norm.SeverityError2
	case 19:
		return norm.SeverityError3
	case 20:
		return norm.SeverityError4
	case 21:
		return norm.SeverityFatal
	case 22:
		return norm.SeverityFatal2
	case 23:
		return norm.SeverityFatal3
	case 24:
		return norm.SeverityFatal4
	default:
		return norm.SeverityError
	}
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
