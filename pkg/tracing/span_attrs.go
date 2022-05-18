package tracing

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/cespare/xxhash/v2"
	ua "github.com/mileusna/useragent"
	"github.com/uptrace/uptrace/pkg/logparser"
	"github.com/uptrace/uptrace/pkg/sqlparser"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

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
	errorEventType     = "error"
	messageEventType   = "message"
	eventType          = "event"
)

type spanContext struct {
	context.Context

	digest *xxhash.Digest
}

func newSpanContext(ctx context.Context) *spanContext {
	return &spanContext{
		Context: ctx,

		digest: xxhash.New(),
	}
}

func initSpan(ctx *spanContext, span *Span) {
	if span.Attrs.ServiceName() == "" {
		span.Attrs[xattr.ServiceName] = "unknown_service"
	}
	if span.Attrs.HostName() == "" {
		span.Attrs[xattr.HostName] = "unknown_host"
	}
	if s, _ := span.Attrs[xattr.HTTPUserAgent].(string); s != "" {
		initHTTPUserAgent(span.Attrs, s)
	}

	assignSpanSystemAndGroupID(ctx, span)
	if span.Name == "" {
		span.Name = "<empty>"
	}
}

func initHTTPUserAgent(attrs AttrMap, str string) {
	agent := ua.Parse(str)

	if agent.Name != "" {
		attrs[xattr.HTTPUserAgentName] = agent.Name
	}
	if agent.Version != "" {
		attrs[xattr.HTTPUserAgentVersion] = agent.Version
	}

	if agent.OS != "" {
		attrs[xattr.HTTPUserAgentOS] = agent.OS
	}
	if agent.OSVersion != "" {
		attrs[xattr.HTTPUserAgentOSVersion] = agent.OSVersion
	}

	if agent.Device != "" {
		attrs[xattr.HTTPUserAgentDevice] = agent.Device
	}

	if agent.Bot {
		attrs[xattr.HTTPUserAgentBot] = 1
	}
}

func newSpanLink(link *tracepb.Span_Link) *SpanLink {
	return &SpanLink{
		TraceID: otlpTraceID(link.TraceId),
		SpanID:  otlpSpanID(link.SpanId),
		Attrs:   otlpAttrs(link.Attributes),
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

func assignSpanSystemAndGroupID(ctx *spanContext, span *Span) {
	if s := span.Attrs.Text(xattr.RPCSystem); s != "" {
		span.System = rpcSpanType + ":" + span.Attrs.ServiceName()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span,
				xattr.RPCSystem,
				xattr.RPCService,
				xattr.RPCMethod,
			)
		})
		return
	}

	if s := span.Attrs.Text(xattr.MessagingSystem); s != "" {
		span.System = messagingSpanType + ":" + s
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span,
				xattr.MessagingSystem,
				xattr.MessagingOperation,
				xattr.MessagingDestination,
				xattr.MessagingDestinationKind,
			)
		})
		return
	}

	if s := span.Attrs.Text(xattr.DBSystem); s != "" {
		span.System = dbSpanType + ":" + s
		stmt, _ := span.Attrs[xattr.DBStatement].(string)

		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span, xattr.DBOperation, xattr.DBSqlTable)
			if stmt != "" {
				hashDBStmt(digest, stmt)
			}
		})
		if stmt != "" {
			span.Name = stmt
		}
		return
	}

	if span.Attrs.Has(xattr.HTTPRoute) || span.Attrs.Has(xattr.HTTPTarget) {
		span.System = httpSpanType + ":" + span.Attrs.ServiceName()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span, xattr.HTTPMethod, xattr.HTTPRoute)
		})
		return
	}

	if span.ParentID == 0 || span.Kind != internalSpanKind {
		span.System = serviceSpanType + ":" + span.Attrs.ServiceName()
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span)
		})
		return
	}

	span.System = internalSpanType
	span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
		hashSpan(digest, span)
	})
}

func spanHash(digest *xxhash.Digest, fn func(digest *xxhash.Digest)) uint64 {
	digest.Reset()
	fn(digest)
	return digest.Sum64()
}

func hashSpan(digest *xxhash.Digest, span *Span, keys ...string) {
	digest.WriteString(span.System)
	digest.WriteString(span.Kind)
	if span.EventName != "" {
		digest.WriteString(span.EventName)
	} else {
		digest.WriteString(span.Name)
	}

	if env, _ := span.Attrs[xattr.DeploymentEnvironment].(string); env != "" {
		digest.WriteString(env)
	}

	for _, key := range keys {
		if value, ok := span.Attrs[key]; ok {
			digest.WriteString(key)
			digest.WriteString(fmt.Sprint(value))
		}
	}
}

//------------------------------------------------------------------------------

func initSpanEvent(ctx *spanContext, dest *Span, hostSpan *Span) {
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

	assignEventSystemAndGroupID(ctx, dest)
}

func assignEventSystemAndGroupID(ctx *spanContext, span *Span) {
	switch span.EventName {
	case logEventType:
		sev, _ := span.Attrs[xattr.LogSeverity].(string)
		if sev == "" {
			sev = "INFO"
		}

		span.System = logEventType + ":" + strings.ToLower(sev)
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span,
				xattr.LogSeverity,
				xattr.LogSource,
				xattr.LogFilepath,
			)
			if s, _ := span.Attrs[xattr.LogMessage].(string); s != "" {
				hashMessage(ctx.digest, s)
			}
		})
		span.EventName = spanLogEventName(span)
		return
	case exceptionEventType, errorEventType:
		span.System = exceptionEventType
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span, xattr.ExceptionType)
			if s, _ := span.Attrs[xattr.ExceptionMessage].(string); s != "" {
				hashMessage(ctx.digest, s)
			}
		})
		span.EventName = joinTypeMessage(
			span.Attrs.Text(xattr.ExceptionType),
			span.Attrs.Text(xattr.ExceptionMessage),
		)
		return
	case messageEventType:
		span.System = spanMessageEventType(span)
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span,
				xattr.RPCSystem,
				xattr.RPCService,
				xattr.RPCMethod,
				"message.type",
			)
		})
		span.EventName = spanMessageEventName(span)
		return
	default:
		span.System = eventType
		span.GroupID = spanHash(ctx.digest, func(digest *xxhash.Digest) {
			hashSpan(digest, span)
		})
		if span.EventName == "" {
			span.EventName = "<empty>"
		}
		return
	}
}

func spanLogEventName(span *Span) string {
	if msg, _ := span.Attrs[xattr.LogMessage].(string); msg != "" {
		if sev, _ := span.Attrs[xattr.LogSeverity].(string); sev != "" {
			msg = sev + " " + msg
		}
		return msg
	}

	typ, _ := span.Attrs[xattr.ExceptionType].(string)
	msg, _ := span.Attrs[xattr.ExceptionMessage].(string)
	if typ != "" || msg != "" {
		return joinTypeMessage(typ, msg)
	}

	return span.EventName
}

func spanMessageEventType(span *Span) string {
	if sys := span.Attrs.Text(xattr.RPCSystem); sys != "" {
		return messageEventType + ":" + sys
	}
	if sys := span.Attrs.Text(xattr.MessagingSystem); sys != "" {
		return messageEventType + ":" + sys
	}
	return messageEventType + ":unknown"
}

func spanMessageEventName(span *Span) string {
	if span.EventName != messageEventType {
		return span.EventName
	}
	if op := span.Attrs.Text(xattr.MessagingOperation); op != "" {
		return join(span.Name, op)
	}
	if typ := span.Attrs.Text("message.type"); typ != "" {
		return join(span.Name, typ)
	}
	if span.Kind != internalSpanKind {
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
