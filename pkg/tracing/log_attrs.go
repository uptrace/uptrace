package tracing

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/logparser"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/anyconv"
	"github.com/uptrace/uptrace/pkg/tracing/norm"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"go.uber.org/zap"
)

func (p *logProcessorThread) initLogOrEvent(ctx context.Context, log *Span) {
	p.logger.Debug("Initializing log", zap.String("logID", log.ID.String()), zap.Any("attributes", log.Attrs))
	project, ok := p.project(ctx, log.ProjectID)
	if !ok {
		return
	}

	p.processLogAttrs(log)

	if log.EventName != "" {
		p.assignEventSystemAndGroupIDForLog(project, log)
		log.EventName = utf8util.TruncLC(log.EventName)
	} else {
		p.assignLogSystemAndGroupID(project, log)
		log.Name = utf8util.TruncLC(log.Name)
	}

	if name, _ := log.Attrs[attrkey.DisplayName].(string); name != "" {
		log.DisplayName = name
		delete(log.Attrs, attrkey.DisplayName)
	} else if log.DisplayName == "" || p.forceLogName(ctx, log) {
		log.DisplayName = log.EventOrSpanName()
	}

	log.System = utf8util.TruncSmall(log.System)
}

func (p *logProcessorThread) processLogAttrs(log *Span) {
	normalizeAttrs(log.Attrs)

	if msg, _ := log.Attrs[attrkey.LogMessage].(string); msg != "" {
		p.parseLogMessage(log, msg)
	}
	if s, _ := log.Attrs[attrkey.UserAgentOriginal].(string); s != "" {
		initHTTPUserAgent(log.Attrs, s)
	}

	switch log.EventName {
	case otelEventLog:
		if _, ok := log.Attrs[attrkey.LogSeverity]; !ok {
			log.Attrs[attrkey.LogSeverity] = norm.SeverityInfo
		}
	case otelEventException:
		if _, ok := log.Attrs[attrkey.LogSeverity]; !ok {
			log.Attrs[attrkey.LogSeverity] = norm.SeverityError
		}
	}
}

func (p *logProcessorThread) parseLogMessage(log *Span, msg string) {
	hash, params := p.messageHashAndParams(msg)
	if log.EventName == otelEventLog {
		log.logMessageHash = hash
	}
	populateLogFromParams(log, params)
}

func (p *logProcessorThread) messageHashAndParams(msg string) (uint64, map[string]any) {
	digest := p.digest
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

func populateLogFromParams(log *Span, params AttrMap) {
	attrs := log.Attrs

	if log.TraceID.IsZero() {
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
				log.TraceID = traceID
				delete(params, key)
				break
			}
		}
	}

	if log.ParentID == 0 {
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
				log.ParentID = spanID
				delete(params, key)
				break
			}
		}
	}

	if log.Time.IsZero() {
		for _, key := range []string{"timestamp", "datetime", "time"} {
			value, _ := params[key].(string)
			if value == "" {
				continue
			}

			tm := anyconv.Time(value)
			if tm.IsZero() {
				continue
			}

			log.Time = tm
			delete(params, key)
			break
		}
	}

	for key, value := range params {
		attrs.SetClashingKeys(key, value)
	}
}

func (p *logProcessorThread) assignLogSystemAndGroupID(project *org.Project, log *Span) {
	if s := log.Attrs.Text(attrkey.RPCSystem); s != "" {
		log.Type = TypeSpanRPC
		log.System = TypeSpanRPC + ":" + log.Attrs.ServiceNameOrUnknown()
		log.GroupID = p.logHash(func(digest *xxhash.Digest) {
			hashLog(project, digest, log,
				attrkey.RPCSystem,
				attrkey.RPCService,
				attrkey.RPCMethod,
			)
		})
		return
	}

	if s := log.Attrs.Text(attrkey.MessagingSystem); s != "" {
		log.Type = TypeSpanMessaging
		log.System = TypeSpanMessaging + ":" + s
		log.GroupID = p.logHash(func(digest *xxhash.Digest) {
			hashLog(project, digest, log,
				attrkey.MessagingSystem,
				attrkey.MessagingOperation,
				attrkey.MessagingDestinationName,
				attrkey.MessagingDestinationKind,
			)
		})
		return
	}

	if dbSystem := log.Attrs.Text(attrkey.DBSystem); dbSystem != "" ||
		log.Attrs.Exists(attrkey.DBStatement) {
		if dbSystem == "" {
			dbSystem = "unknown_db"
		}

		log.Type = TypeSpanDB
		log.System = TypeSpanDB + ":" + dbSystem
		stmt, _ := log.Attrs[attrkey.DBStatement].(string)

		log.GroupID = p.logHash(func(digest *xxhash.Digest) {
			hashLog(project, digest, log, attrkey.DBName, attrkey.DBOperation, attrkey.DBSqlTable)
			if stmt != "" {
				hashDBStmt(digest, stmt)
			}
		})
		if stmt != "" {
			log.DisplayName = stmt
		}
		return
	}

	if log.Attrs.Exists(attrkey.HTTPRoute) ||
		log.Attrs.Exists(attrkey.HTTPRequestMethod) ||
		log.Attrs.Exists(attrkey.HTTPResponseStatusCode) {
		if log.Kind == SpanKindClient {
			log.Type = TypeSpanHTTPClient
		} else {
			log.Type = TypeSpanHTTPServer
		}
		log.System = log.Type + ":" + log.Attrs.ServiceNameOrUnknown()
		log.GroupID = p.logHash(func(digest *xxhash.Digest) {
			hashLog(project, digest, log, attrkey.HTTPRequestMethod, attrkey.HTTPRoute)
		})
		log.DisplayName = httpDisplayNameForLog(log)
		return
	}

	log.Type = TypeSpanFuncs
	if project.GroupFuncsByService {
		service := log.Attrs.ServiceNameOrUnknown()
		log.System = TypeSpanFuncs + ":" + service
	} else {
		log.System = TypeSpanFuncs
	}
	log.GroupID = p.logHash(func(digest *xxhash.Digest) {
		hashLog(project, digest, log)
	})
}

func httpDisplayNameForLog(log *Span) string {
	httpMethod := log.Attrs.GetString(attrkey.HTTPRequestMethod)
	httpRoute := log.Attrs.GetString(attrkey.HTTPRoute)
	httpServerName := log.Attrs.GetString(attrkey.ServerAddress)
	httpTarget := log.Attrs.GetString(attrkey.URLPath)

	switch log.Name {
	case "", httpMethod, httpRoute, httpServerName, httpTarget, "HTTP " + httpMethod:
	default:
		return log.Name
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

	if httpURL := log.Attrs.GetString(attrkey.URLFull); httpURL != "" {
		u, err := url.Parse(httpURL)
		if err == nil && u.Path != "" {
			if !log.Attrs.Exists(attrkey.URLScheme) {
				log.Attrs.PutString(attrkey.URLScheme, u.Scheme)
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

	return log.Name
}

func (p *logProcessorThread) logHash(fn func(digest *xxhash.Digest)) uint64 {
	p.digest.Reset()
	fn(p.digest)
	return p.digest.Sum64()
}

func hashLog(project *org.Project, digest *xxhash.Digest, log *Span, keys ...string) {
	if project.GroupByEnv {
		if env := log.Attrs.Text(attrkey.DeploymentEnvironment); env != "" {
			digest.WriteString(env)
		}
	}
	digest.WriteString(log.System)
	digest.WriteString(log.Kind)
	if log.EventName != "" {
		digest.WriteString(log.EventName)
	} else {
		digest.WriteString(log.Name)
	}

	for _, key := range keys {
		if value, ok := log.Attrs[key]; ok {
			digest.WriteString(key)
			digest.WriteString(fmt.Sprint(value))
		}
	}
}

func (p *logProcessorThread) assignEventSystemAndGroupIDForLog(project *org.Project, log *Span) {
	if log.EventName == otelEventError {
		log.EventName = otelEventException
	}

	switch log.EventName {
	case otelEventLog:
		p.handleLogEventForLog(project, log)
		return
	case otelEventException:
		p.handleExceptionEventForLog(project, log)
		return
	case otelEventMessage:
		system := eventMessageSystemForLog(log)
		log.Type = TypeEventMessage
		log.System = TypeEventMessage + ":" + system
		log.GroupID = p.logHash(func(digest *xxhash.Digest) {
			hashLog(project, digest, log,
				attrkey.RPCSystem,
				attrkey.RPCService,
				attrkey.RPCMethod,
				attrkey.MessagingMessageType,
			)
		})
		log.DisplayName = eventMessageDisplay(log)
		return
	}

	if log.Attrs.Exists(attrkey.LogMessage) {
		p.handleLogEventForLog(project, log)
		return
	}

	if log.Attrs.Exists(attrkey.ExceptionMessage) {
		p.handleExceptionEventForLog(project, log)
		return
	}

	log.Type = TypeEventOther
	log.System = TypeEventOther
	log.GroupID = p.logHash(func(digest *xxhash.Digest) {
		hashLog(project, digest, log)
	})
	log.DisplayName = log.EventName
}

func (p *logProcessorThread) handleLogEventForLog(project *org.Project, log *Span) {
	sev, _ := log.Attrs[attrkey.LogSeverity].(string)
	log.Type = TypeLog
	log.System = TypeLog + ":" + lowerSeverity(sev)
	log.GroupID = p.logHash(func(digest *xxhash.Digest) {
		hashLog(project, digest, log, attrkey.LogSeverity)
		if log.logMessageHash != 0 {
			digest.WriteString(strconv.FormatUint(log.logMessageHash, 10))
		}
	})
	log.DisplayName = logDisplayNameForLog(log)
}

func (p *logProcessorThread) handleExceptionEventForLog(project *org.Project, log *Span) {
	log.Type = TypeLog
	log.System = SystemLogError
	log.GroupID = p.logHash(func(digest *xxhash.Digest) {
		hashLog(project, digest, log, attrkey.ExceptionType)
		if s, _ := log.Attrs[attrkey.ExceptionMessage].(string); s != "" {
			hashMessage(digest, s)
		}
	})
	log.DisplayName = exceptionDisplayNameForLog(log)
}

func logDisplayNameForLog(log *Span) string {
	if msg, _ := log.Attrs[attrkey.LogMessage].(string); msg != "" {
		log.Attrs.Delete(attrkey.LogMessage)
		sev, _ := log.Attrs[attrkey.LogSeverity].(string)
		if sev != "" && !strings.HasPrefix(msg, sev) {
			msg = sev + " " + msg
		}
		return msg
	}

	if name := exceptionDisplayNameForLog(log); name != log.EventName {
		return name
	}

	return log.EventName
}

func exceptionDisplayNameForLog(log *Span) string {
	msg, _ := log.Attrs[attrkey.ExceptionMessage].(string)
	if msg != "" {
		log.Attrs.Delete(attrkey.ExceptionMessage)
		typ, _ := log.Attrs[attrkey.ExceptionType].(string)
		return joinTypeMessage(typ, msg)
	}
	return log.EventName
}

func eventMessageSystemForLog(log *Span) string {
	if sys := log.Attrs.Text(attrkey.RPCSystem); sys != "" {
		return sys
	}
	if sys := log.Attrs.Text(attrkey.MessagingSystem); sys != "" {
		return sys
	}
	return SystemUnknown
}

func eventMessageDisplay(log *Span) string {
	if log.EventName != otelEventMessage {
		return log.EventName
	}
	if op := log.Attrs.Text(attrkey.MessagingOperation); op != "" {
		return join(log.Name, op)
	}
	if typ := log.Attrs.Text("message.type"); typ != "" {
		return join(log.Name, typ)
	}
	if log.Kind != InternalSpanKind {
		return join(log.Name, log.Kind)
	}
	return log.EventName
}

func lowerSeverity(sev string) string {
	switch sev {
	case norm.SeverityTrace, norm.SeverityTrace2, norm.SeverityTrace3, norm.SeverityTrace4:
		return "trace"
	case norm.SeverityDebug, norm.SeverityDebug2, norm.SeverityDebug3, norm.SeverityDebug4:
		return "debug"
	case norm.SeverityInfo, norm.SeverityInfo2, norm.SeverityInfo3, norm.SeverityInfo4:
		return "info"
	case norm.SeverityWarn, norm.SeverityWarn2, norm.SeverityWarn3, norm.SeverityWarn4:
		return "warn"
	case norm.SeverityError, norm.SeverityError2, norm.SeverityError3, norm.SeverityError4:
		return "error"
	case norm.SeverityFatal, norm.SeverityFatal2, norm.SeverityFatal3, norm.SeverityFatal4:
		return "fatal"
	default:
		return "error"
	}
}
