package tracing

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	ua "github.com/mileusna/useragent"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/logparser"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/sqlparser"
	"github.com/uptrace/uptrace/pkg/tracing/anyconv"
	"github.com/uptrace/uptrace/pkg/utf8util"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func (p *spanProcessorThread) initSpanOrEvent(ctx context.Context, span *Span) {
	project, ok := p.project(ctx, span.ProjectID)
	if !ok {
		return
	}

	p.processAttrs(span)

	if span.EventName != "" {
		p.assignSpanSystemAndGroupID(project, span)
		span.EventName = utf8util.TruncLC(span.EventName)
	} else {
		p.assignSpanSystemAndGroupID(project, span)
		span.Name = utf8util.TruncLC(span.Name)
	}

	if name, _ := span.Attrs[attrkey.DisplayName].(string); name != "" {
		span.DisplayName = name
		delete(span.Attrs, attrkey.DisplayName)
	} else if span.DisplayName == "" || p.forceSpanName(ctx, span) {
		span.DisplayName = span.EventOrSpanName()
	}

	span.System = utf8util.TruncSmall(span.System)
}

func (p *spanProcessorThread) processAttrs(span *Span) {
	normalizeAttrs(span.Attrs)

	if s, _ := span.Attrs[attrkey.UserAgentOriginal].(string); s != "" {
		initHTTPUserAgent(span.Attrs, s)
	}

	if span.TraceID.IsZero() {
		span.TraceID = idgen.RandTraceID()
		span.ID = 0
		span.ParentID = 0
		span.Standalone = true
	}
	if !span.Standalone {
		if span.ID == 0 {
			span.ID = idgen.RandSpanID()
		}
	}
	if span.Time.IsZero() {
		span.Time = time.Now()
	}
}

func populateSpanFromParams(span *Span, params AttrMap) {
	attrs := span.Attrs

	if span.TraceID.IsZero() {
		for _, key := range []string{"trace_id", "traceid"} {
			value, _ := params[key].(string)
			if value == "" {
				continue
			}

			traceID, err := idgen.ParseTraceID(value)
			if err != nil {
				continue
			}

			if !traceID.IsZero() {
				span.TraceID = traceID
				delete(params, key)
				break
			}
		}
	}

	if span.ParentID == 0 {
		for _, key := range []string{"span_id", "spanid"} {
			value, _ := params[key].(string)
			if value == "" {
				continue
			}

			spanID, err := idgen.ParseSpanID(value)
			if err != nil {
				continue
			}

			if spanID != 0 {
				span.ParentID = spanID
				delete(params, key)
				break
			}
		}
	}

	if span.Time.IsZero() {
		for _, key := range []string{"timestamp", "datetime", "time"} {
			value, _ := params[key].(string)
			if value == "" {
				continue
			}

			tm := anyconv.Time(value)
			if tm.IsZero() {
				continue
			}

			span.Time = tm
			delete(params, key)
			break
		}
	}

	if span.Duration == 0 {
		if val, ok := params.Get(attrkey.SpanDuration); ok {
			span.Duration = time.Duration(anyconv.Int64(val))
			params.Delete(attrkey.SpanDuration)
			span.EventName = ""
		}
	}

	for key, value := range params {
		attrs.SetClashingKeys(key, value)
	}
}

func initHTTPUserAgent(attrs AttrMap, str string) {
	agent := ua.Parse(str)

	if agent.Name != "" {
		attrs[attrkey.UserAgentName] = agent.Name
	}
	if agent.Version != "" {
		attrs[attrkey.UserAgentVersion] = agent.Version
	}

	if agent.OS != "" {
		attrs[attrkey.UserAgentOSName] = agent.OS
	}
	if agent.OSVersion != "" {
		attrs[attrkey.UserAgentOSVersion] = agent.OSVersion
	}

	if agent.Device != "" {
		attrs[attrkey.UserAgentDevice] = agent.Device
	}

	if agent.Bot {
		attrs[attrkey.UserAgentIsBot] = 1
	}
}

func newSpanLink(link *tracepb.Span_Link) *SpanLink {
	return &SpanLink{
		TraceID: idgen.TraceIDFromBytes(link.TraceId),
		SpanID:  idgen.SpanIDFromBytes(link.SpanId),
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

func (p *spanProcessorThread) assignSpanSystemAndGroupID(project *org.Project, span *Span) {
	if s := span.Attrs.Text(attrkey.RPCSystem); s != "" {
		span.Type = TypeSpanRPC
		span.System = TypeSpanRPC + ":" + span.Attrs.ServiceNameOrUnknown()
		span.GroupID = p.spanHash(func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.RPCSystem,
				attrkey.RPCService,
				attrkey.RPCMethod,
			)
		})
		return
	}

	if s := span.Attrs.Text(attrkey.MessagingSystem); s != "" {
		span.Type = TypeSpanMessaging
		span.System = TypeSpanMessaging + ":" + s
		span.GroupID = p.spanHash(func(digest *xxhash.Digest) {
			hashSpan(project, digest, span,
				attrkey.MessagingSystem,
				attrkey.MessagingOperation,
				attrkey.MessagingDestinationName,
				attrkey.MessagingDestinationKind,
			)
		})
		return
	}

	if dbSystem := span.Attrs.Text(attrkey.DBSystem); dbSystem != "" ||
		span.Attrs.Exists(attrkey.DBStatement) {
		if dbSystem == "" {
			dbSystem = "unknown_db"
		}

		span.Type = TypeSpanDB
		span.System = TypeSpanDB + ":" + dbSystem
		stmt, _ := span.Attrs[attrkey.DBStatement].(string)

		span.GroupID = p.spanHash(func(digest *xxhash.Digest) {
			hashSpan(project, digest, span, attrkey.DBName, attrkey.DBOperation, attrkey.DBSqlTable)
			if stmt != "" {
				hashDBStmt(digest, stmt)
			}
		})
		if stmt != "" {
			span.DisplayName = stmt
		}
		return
	}

	if span.Attrs.Exists(attrkey.HTTPRoute) ||
		span.Attrs.Exists(attrkey.HTTPRequestMethod) ||
		span.Attrs.Exists(attrkey.HTTPResponseStatusCode) {
		if span.Kind == SpanKindClient {
			span.Type = TypeSpanHTTPClient
		} else {
			span.Type = TypeSpanHTTPServer
		}
		span.System = span.Type + ":" + span.Attrs.ServiceNameOrUnknown()
		span.GroupID = p.spanHash(func(digest *xxhash.Digest) {
			hashSpan(project, digest, span, attrkey.HTTPRequestMethod, attrkey.HTTPRoute)
		})
		span.DisplayName = httpDisplayName(span)
		return
	}

	span.Type = TypeSpanFuncs
	if project.GroupFuncsByService {
		service := span.Attrs.ServiceNameOrUnknown()
		span.System = TypeSpanFuncs + ":" + service
	} else {
		span.System = TypeSpanFuncs
	}
	span.GroupID = p.spanHash(func(digest *xxhash.Digest) {
		hashSpan(project, digest, span)
	})
}

func httpDisplayName(span *Span) string {
	httpMethod := span.Attrs.GetString(attrkey.HTTPRequestMethod)
	httpRoute := span.Attrs.GetString(attrkey.HTTPRoute)
	httpServerName := span.Attrs.GetString(attrkey.ServerAddress)
	httpTarget := span.Attrs.GetString(attrkey.URLPath)

	switch span.Name {
	case "", httpMethod, httpRoute, httpServerName, httpTarget, "HTTP " + httpMethod:
	default:
		return span.Name
	}

	if httpRoute != "" {
		if httpMethod != "" {
			return httpMethod + " " + httpRoute
		}
		return httpRoute
	}

	if httpTarget != "" {
		if idx := strings.IndexByte(httpTarget, '?'); idx >= 0 {
			httpTarget = httpTarget[:idx]
		}
		if httpMethod != "" {
			return httpMethod + " " + httpTarget
		}
		return httpTarget
	}

	if httpURL := span.Attrs.GetString(attrkey.URLFull); httpURL != "" {
		u, err := url.Parse(httpURL)
		if err == nil && u.Path != "" {
			if !span.Attrs.Exists(attrkey.URLScheme) {
				span.Attrs.PutString(attrkey.URLScheme, u.Scheme)
			}

			if httpMethod != "" {
				return httpMethod + " " + u.Path
			}
			return u.Path
		}
	}

	if httpMethod != "" {
		return httpMethod + " <unknown_route>"
	}

	return span.Name
}

func (p *spanProcessorThread) spanHash(fn func(digest *xxhash.Digest)) uint64 {
	p.digest.Reset()
	fn(p.digest)
	return p.digest.Sum64()
}

func hashSpan(project *org.Project, digest *xxhash.Digest, span *Span, keys ...string) {
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

type AttrName struct {
	Canonical string
	Alts      []string
}

var attrNames = []AttrName{
	{
		Canonical: attrkey.DeploymentEnvironment,
		Alts:      []string{"environment", "env"},
	},
	{Canonical: attrkey.ServiceName, Alts: []string{"service", "component"}},
	{Canonical: attrkey.URLScheme, Alts: []string{"http_scheme"}},
	{Canonical: attrkey.URLFull, Alts: []string{"http_url"}},
	{Canonical: attrkey.URLPath, Alts: []string{"http_target"}},
	{Canonical: attrkey.HTTPRequestMethod, Alts: []string{"http_method"}},
	{Canonical: attrkey.HTTPResponseStatusCode, Alts: []string{"http_status_code"}},
	{Canonical: attrkey.HTTPResponseStatusClass, Alts: []string{"http_status_class"}},
	{Canonical: attrkey.DBSystem, Alts: []string{"db_type"}},
}

func normalizeAttrs(attrs AttrMap) {
	for _, name := range attrNames {
		if _, ok := attrs[name.Canonical]; ok {
			continue
		}

		for _, key := range name.Alts {
			if val, ok := attrs[key]; ok {
				delete(attrs, key)
				attrs[name.Canonical] = val
				break
			}
		}
	}

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

func initEventFromHostSpan(dest *Span, event *SpanEvent, hostSpan *Span) {
	dest.EventName = event.Name
	dest.Time = event.Time
	dest.Attrs = event.Attrs

	dest.ProjectID = hostSpan.ProjectID
	dest.TraceID = hostSpan.TraceID
	dest.ID = idgen.RandSpanID()
	dest.ParentID = hostSpan.ID

	dest.Name = hostSpan.Name
	dest.Kind = hostSpan.Kind
	for k, v := range hostSpan.Attrs {
		if _, ok := dest.Attrs[k]; !ok {
			dest.Attrs[k] = v
		}
	}
	dest.StatusCode = hostSpan.StatusCode
}

func (p *spanProcessorThread) initEvent(ctx context.Context, span *Span) {
	project, ok := p.project(ctx, span.ProjectID)
	if !ok {
		return
	}

	p.processAttrs(span)
	p.assignSpanSystemAndGroupID(project, span)
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
