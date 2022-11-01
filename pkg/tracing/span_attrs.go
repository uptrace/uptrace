package tracing

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	ua "github.com/mileusna/useragent"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/logparser"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/sqlparser"
	"github.com/uptrace/uptrace/pkg/tracing/anyconv"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"github.com/uptrace/uptrace/pkg/uuid"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
)

type spanContext struct {
	context.Context
	*bunapp.App

	projects map[uint32]*bunconf.Project
	digest   *xxhash.Digest
}

func newSpanContext(ctx context.Context, app *bunapp.App) *spanContext {
	return &spanContext{
		Context: ctx,
		App:     app,

		projects: make(map[uint32]*bunconf.Project),
		digest:   xxhash.New(),
	}
}

func (c *spanContext) Project(projectID uint32) (*bunconf.Project, bool) {
	if p, ok := c.projects[projectID]; ok {
		return p, true
	}

	project, err := org.SelectProject(c.Context, c.App, projectID)
	if err != nil {
		c.Zap(c.Context).Error("SelectProjectCached failed", zap.Error(err))
		return nil, false
	}

	c.projects[projectID] = project
	return project, true
}

// initSpan initializes spans.
func initSpanOrEvent(ctx *spanContext, span *Span) {
	project, ok := ctx.Project(span.ProjectID)
	if !ok {
		return
	}

	initSpanAttrs(ctx, span)
	if span.EventName != "" {
		assignEventSystemAndGroupID(ctx, project, span)
		span.EventName = utf8util.TruncMedium(span.EventName)
	} else {
		assignSpanSystemAndGroupID(ctx, project, span)
		span.Name = utf8util.TruncMedium(span.Name)
		if span.Name == "" {
			span.Name = "<empty>"
		}
	}
	span.System = utf8util.TruncSmall(span.System)
}

func initSpanAttrs(ctx *spanContext, span *Span) {
	if service := serviceNameAndVersion(span.Attrs); service != "" {
		span.Attrs[attrkey.Service] = service
	}
	if s, _ := span.Attrs[attrkey.HTTPUserAgent].(string); s != "" {
		initHTTPUserAgent(span.Attrs, s)
	}
	if msg, ok := span.Attrs[attrkey.LogMessage].(string); ok {
		initLogMessage(ctx, span, msg)
	}
	initLogSeverity(span.Attrs)
}

func initHTTPUserAgent(attrs AttrMap, str string) {
	agent := ua.Parse(str)

	if agent.Name != "" {
		attrs[attrkey.HTTPUserAgentName] = agent.Name
	}
	if agent.Version != "" {
		attrs[attrkey.HTTPUserAgentVersion] = agent.Version
	}

	if agent.OS != "" {
		attrs[attrkey.HTTPUserAgentOS] = agent.OS
	}
	if agent.OSVersion != "" {
		attrs[attrkey.HTTPUserAgentOSVersion] = agent.OSVersion
	}

	if agent.Device != "" {
		attrs[attrkey.HTTPUserAgentDevice] = agent.Device
	}

	if agent.Bot {
		attrs[attrkey.HTTPUserAgentBot] = 1
	}
}

func serviceNameAndVersion(attrs AttrMap) string {
	name, _ := attrs[attrkey.ServiceName].(string)
	if name == "" {
		return ""
	}
	if version := attrs.Text(attrkey.ServiceVersion); version != "" {
		return name + "@" + version
	}
	return name
}

//------------------------------------------------------------------------------

func initLogMessage(ctx *spanContext, span *Span, msg string) {
	if msg == "" {
		delete(span.Attrs, attrkey.LogMessage)
		return
	}

	if m, ok := logparser.IsJSON(msg); ok {
		// Delete the attribute so we can override the message.
		delete(span.Attrs, attrkey.LogMessage)

		spanFromJSONLog(span, m)

		if s, ok := span.Attrs[attrkey.LogMessage].(string); ok {
			msg = s
		} else {
			// Restore the attribute.
			span.Attrs[attrkey.LogMessage] = msg
		}
	}

	hash, params := messageHashAndParams(ctx, msg)
	span.logMessageHash = hash

	for k, v := range params {
		span.Attrs.SetDefault(k, v)
	}

	promoteLogParamsToSpan(span, params)
}

func spanFromJSONLog(span *Span, src AttrMap) {
	attrs := span.Attrs
	for key, value := range src {
		switch key {
		case "level", "severity":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.LogSeverity, s)
			}
		case "message", "msg":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.LogMessage, s)
			}
		case "time":
			if tm := anyconv.Time(value); !tm.IsZero() {
				span.Time = tm
			}
		case "trace_id", "traceid":
			span.TraceID = anyconv.UUID(value)
		case "span_id", "spanid":
			span.ParentID = anyconv.Uint64(value)
		case "service":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.ServiceName, s)
			}
		case "host", "hostname":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.HostName, s)
			}
		default:
			attrs.SetDefault(key, value)
		}
	}
}

func promoteLogParamsToSpan(span *Span, params map[string]any) {
	if span.TraceID.IsZero() {
		traceID := anyconv.UUID(params["trace_id"])
		if traceID.IsZero() {
			// Standalone span.
			span.TraceID = uuid.Rand()
			span.ParentID = 0
			span.Standalone = true
			return
		}
		span.TraceID = traceID
	}

	if span.ParentID == 0 {
		if id := span.Attrs.Uint64("span_id"); id != 0 {
			span.ParentID = id
		} else {
			// Assign random ID and handle it when assembling the trace tree.
			// Otherwise, the span will be treated as a root.
			span.ParentID = rand.Uint64()
		}
	}
}

func initLogSeverity(attrs AttrMap) {
	if found, ok := attrs[attrkey.LogSeverity].(string); ok {
		if normalized := normalizeLogSeverity(found); normalized != "" {
			if normalized != found {
				attrs[attrkey.LogSeverity] = normalized
			}
			return
		}
		// We can't normalize the severity. Set the default.
		attrs[attrkey.LogSeverity] = bunotel.InfoSeverity
	}

	if attrs.Has(attrkey.LogSeverityNumber) {
		num := attrs.Uint64(attrkey.LogSeverityNumber)
		if sev := logSeverityFromNumber(num); sev != "" {
			attrs[attrkey.LogSeverity] = sev
			return
		}
	}
}

func normalizeLogSeverity(s string) string {
	if s := _normalizeLogSeverity(s); s != "" {
		return s
	}
	return _normalizeLogSeverity(strings.ToLower(s))
}

func _normalizeLogSeverity(s string) string {
	switch s {
	case "trace":
		return bunotel.TraceSeverity
	case "debug":
		return bunotel.DebugSeverity
	case "info", "information":
		return bunotel.InfoSeverity
	case "warn", "warning":
		return bunotel.WarnSeverity
	case "error", "err", "alert":
		return bunotel.ErrorSeverity
	case "fatal", "crit", "critical", "emerg", "emergency", "panic":
		return bunotel.FatalSeverity
	default:
		return ""
	}
}

func logSeverityFromNumber(n uint64) string {
	switch {
	case n == 0:
		return ""
	case n <= 4:
		return bunotel.TraceSeverity
	case n <= 8:
		return bunotel.DebugSeverity
	case n <= 12:
		return bunotel.InfoSeverity
	case n <= 16:
		return bunotel.WarnSeverity
	case n <= 20:
		return bunotel.ErrorSeverity
	case n <= 24:
		return bunotel.FatalSeverity
	}
	return ""
}

func messageHashAndParams(
	ctx *spanContext, msg string,
) (uint64, map[string]any) {
	digest := ctx.digest
	digest.Reset()

	var params map[string]any

	tok := logparser.NewTokenizer(msg)
loop:
	for {
		tok := tok.NextToken()
		switch tok.Type {
		case logparser.InvalidToken:
			break loop
		case logparser.WordToken:
			digest.WriteString(tok.Text)
		case logparser.ParamToken:
			if k, v, ok := logparser.IsLogfmt(tok.Text); ok {
				if params == nil {
					params = make(map[string]any)
				}
				params[k] = v
			}
		}
	}

	return digest.Sum64(), params
}

//------------------------------------------------------------------------------

func newSpanLink(link *tracepb.Span_Link) *SpanLink {
	return &SpanLink{
		TraceID: otlpTraceID(link.TraceId),
		SpanID:  otlpSpanID(link.SpanId),
		Attrs:   otlpconv.Map(link.Attributes),
	}
}

func hashDBStmt(digest *xxhash.Digest, s string) uint64 {
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
	return digest.Sum64()
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

func assignSpanSystemAndGroupID(ctx *spanContext, project *bunconf.Project, span *Span) {
	if s := span.Attrs.Text(attrkey.RPCSystem); s != "" {
		span.System = SystemSpanRPC + ":" + span.Attrs.ServiceNameOrUnknown()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.RPCSystem,
				attrkey.RPCService,
				attrkey.RPCMethod,
			)
		})
		return
	}

	if s := span.Attrs.Text(attrkey.MessagingSystem); s != "" {
		span.System = SystemSpanMessaging + ":" + s
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.MessagingSystem,
				attrkey.MessagingOperation,
				attrkey.MessagingDestination,
				attrkey.MessagingDestinationKind,
			)
		})
		return
	}

	if s := span.Attrs.Text(attrkey.DBSystem); s != "" {
		span.System = SystemSpanDB + ":" + s
		stmt, _ := span.Attrs[attrkey.DBStatement].(string)

		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span, attrkey.DBOperation, attrkey.DBSqlTable)
			if stmt != "" {
				hashDBStmt(digest, stmt)
			}
		})
		if stmt != "" {
			span.Name = stmt
		}
		return
	}

	if span.Attrs.Has(attrkey.HTTPRoute) || span.Attrs.Has(attrkey.HTTPTarget) {
		span.System = SystemSpanHTTP + ":" + span.Attrs.ServiceNameOrUnknown()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span, attrkey.HTTPMethod, attrkey.HTTPRoute)
		})
		return
	}

	if project.GroupFuncsByService &&
		(span.ParentID == 0 ||
			span.Kind != InternalSpanKind ||
			span.Attrs.Has(attrkey.CodeFunction)) {
		span.System = SystemSpanFuncs + ":" + span.Attrs.ServiceNameOrUnknown()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span)
		})
		return
	}

	span.System = SystemSpanFuncs
	span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
		hashSpan(project, digest, span)
	})
}

func spanHash(digest *xxhash.Digest, fn func(digest *xxhash.Digest)) uint64 {
	digest.Reset()
	fn(digest)
	return digest.Sum64()
}

func hashSpan(project *bunconf.Project, digest *xxhash.Digest, span *Span, keys ...string) {
	if project.GroupByEnv {
		if env := span.Attrs.Text(attrkey.DeploymentEnvironment); env != "" {
			digest.WriteString(env)
		}
	}
	digest.WriteString(span.System)
	digest.WriteString(span.Kind)
	if span.EventName != "" {
		digest.WriteString(span.EventName)
	} else {
		digest.WriteString(span.Name)
	}

	for _, key := range keys {
		if value, ok := span.Attrs[key]; ok {
			digest.WriteString(key)
			digest.WriteString(fmt.Sprint(value))
		}
	}
}

//------------------------------------------------------------------------------

func initEventFromHostSpan(dest *Span, hostSpan *Span) {
	dest.ProjectID = hostSpan.ProjectID
	dest.TraceID = hostSpan.TraceID
	dest.ID = rand.Uint64()
	dest.ParentID = hostSpan.ID

	dest.Name = hostSpan.Name
	dest.Kind = hostSpan.Kind
	for k, v := range hostSpan.Attrs {
		dest.Attrs.SetDefault(k, v)
	}
	dest.Duration = hostSpan.Duration
	dest.StatusCode = hostSpan.StatusCode
}

func initEvent(ctx *spanContext, span *Span) {
	project, ok := ctx.Project(span.ProjectID)
	if !ok {
		return
	}

	initSpanAttrs(ctx, span)
	assignEventSystemAndGroupID(ctx, project, span)
}

func assignEventSystemAndGroupID(ctx *spanContext, project *bunconf.Project, span *Span) {
	switch span.EventName {
	case SystemEventLog:
		sev, _ := span.Attrs[attrkey.LogSeverity].(string)
		if sev == "" {
			sev = bunotel.InfoSeverity
		}

		span.System = SystemEventLog + ":" + strings.ToLower(sev)
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.LogSeverity,
			)
			if span.logMessageHash != 0 {
				digest.WriteString(strconv.FormatUint(span.logMessageHash, 10))
			}
		})
		span.EventName = logEventName(span)
		return
	case "exception", "error":
		span.System = SystemEventExceptions
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span, attrkey.ExceptionType)
			if s, _ := span.Attrs[attrkey.ExceptionMessage].(string); s != "" {
				hashMessage(ctx.digest, s)
			}
		})
		span.EventName = joinTypeMessage(
			span.Attrs.Text(attrkey.ExceptionType),
			span.Attrs.Text(attrkey.ExceptionMessage),
		)
		return
	case SystemEventMessage:
		span.System = spanSystemEventMessage(span)
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.RPCSystem,
				attrkey.RPCService,
				attrkey.RPCMethod,
				"message.type",
			)
		})
		span.EventName = spanMessageEventName(span)
		return
	default:
		span.System = SystemEventOther
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(project, digest, span)
		})
		if span.EventName == "" {
			span.EventName = "<empty>"
		}
		return
	}
}

func logEventName(span *Span) string {
	if msg, _ := span.Attrs[attrkey.LogMessage].(string); msg != "" {
		sev, _ := span.Attrs[attrkey.LogSeverity].(string)
		if sev != "" && !strings.HasPrefix(msg, sev) {
			msg = sev + " " + msg
		}
		return msg
	}

	typ, _ := span.Attrs[attrkey.ExceptionType].(string)
	msg, _ := span.Attrs[attrkey.ExceptionMessage].(string)
	if typ != "" || msg != "" {
		return joinTypeMessage(typ, msg)
	}

	return span.EventName
}

func spanSystemEventMessage(span *Span) string {
	if sys := span.Attrs.Text(attrkey.RPCSystem); sys != "" {
		return SystemEventMessage + ":" + sys
	}
	if sys := span.Attrs.Text(attrkey.MessagingSystem); sys != "" {
		return SystemEventMessage + ":" + sys
	}
	return SystemEventMessage + ":unknown"
}

func spanMessageEventName(span *Span) string {
	if span.EventName != SystemEventMessage {
		return span.EventName
	}
	if op := span.Attrs.Text(attrkey.MessagingOperation); op != "" {
		return join(span.Name, op)
	}
	if typ := span.Attrs.Text("message.type"); typ != "" {
		return join(span.Name, typ)
	}
	if span.Kind != InternalSpanKind {
		return join(span.Name, span.Kind)
	}
	return span.EventName
}

func hashMessage(digest *xxhash.Digest, msg string) {
	tok := logparser.NewTokenizer(msg)
loop:
	for {
		tok := tok.NextToken()
		switch tok.Type {
		case logparser.InvalidToken:
			break loop
		case logparser.WordToken:
			digest.WriteString(tok.Text)
		}
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

func join(s1, s2 string) string {
	if s1 != "" {
		return s1 + " " + s2
	}
	return s2
}
