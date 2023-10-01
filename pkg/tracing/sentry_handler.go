package tracing

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"github.com/uptrace/uptrace/pkg/uuid"
	"go.uber.org/zap"
)

type SentryHandler struct {
	*bunapp.App

	sp *SpanProcessor
}

func NewSentryHandler(app *bunapp.App, sp *SpanProcessor) *SentryHandler {
	return &SentryHandler{
		App: app,
		sp:  sp,
	}
}

func (h *SentryHandler) Store(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	project, err := h.projectFromRequest(req)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	event := new(SentryEvent)
	if err := json.Unmarshal(b, &event); err != nil {
		return err
	}

	if err := h.processEvent(ctx, project, event); err != nil {
		return err
	}

	return nil
}

func (h *SentryHandler) processEvent(
	ctx context.Context, project *org.Project, event *SentryEvent,
) error {
	traceID, err := uuid.Parse(event.EventID)
	if err != nil {
		return err
	}

	span := new(Span)

	if err := h.spanFromEvent(span, event); err != nil {
		return err
	}

	span.ProjectID = project.ID
	span.TraceID = traceID
	span.Standalone = true

	if event.Message != "" {
		span.EventName = otelEventLog
		if event.Level != "" {
			span.Attrs[attrkey.LogSeverity] = event.Level
		}
		span.Attrs[attrkey.LogMessage] = event.Message
	}

	h.sp.AddSpan(ctx, span)

	return nil
}

func (h *SentryHandler) spanFromEvent(span *Span, event *SentryEvent) error {
	span.Attrs = make(AttrMap)
	span.Time = event.Timestamp.Time

	for k, m := range event.Contexts {
		forEachKV(m, k+".", func(k string, v any) {
			span.Attrs[k] = v
		})
	}
	for k, v := range event.Tags {
		k = attrkey.Clean(k)
		if k != "" {
			span.Attrs["tags."+k] = v
		}
	}
	forEachKV(event.Extra, "extra.", func(k string, v any) {
		span.Attrs[k] = v
	})

	if event.ServerName != "" {
		span.Attrs[attrkey.HostName] = event.ServerName
	}
	if event.Environment != "" {
		span.Attrs[attrkey.DeploymentEnvironment] = event.Environment
	}
	if event.Release != "" {
		span.Attrs["deployment.name"] = event.Release
	}

	if event.User.ID != "" {
		span.Attrs[attrkey.EnduserID] = event.User.ID
	}
	if event.User.Email != "" {
		span.Attrs["enduser.email"] = event.User.Email
	}
	if event.User.IPAddress != "" {
		span.Attrs["enduser.ip"] = event.User.IPAddress
	}
	if event.User.Username != "" {
		span.Attrs["enduser.username"] = event.User.Username
	}
	if event.User.Name != "" {
		span.Attrs["enduser.name"] = event.User.Name
	}
	if event.User.Segment != "" {
		span.Attrs["enduser.segment"] = event.User.Segment
	}
	for k, v := range event.User.Data {
		k = attrkey.Clean(k)
		if k != "" {
			span.Attrs["enduser.data."+k] = v
		}
	}

	if req := event.Request; req != nil {
		if req.URL != "" {
			span.Attrs[attrkey.HTTPUrl] = req.URL
		}
		if req.Method != "" {
			span.Attrs[attrkey.HTTPMethod] = req.Method
		}
		if req.Data != "" {
			span.Attrs["http.data"] = req.Data
		}
		if req.QueryString != "" {
			span.Attrs["http.query_string"] = req.QueryString
		}
		if req.Cookies != "" {
			span.Attrs["http.cookies"] = req.Cookies
		}
		for k, v := range req.Headers {
			k = attrkey.Clean(k)
			if k != "" {
				span.Attrs["http.headers."+k] = v
			}
		}
		for k, v := range req.Env {
			k = attrkey.Clean(k)
			if k != "" {
				span.Attrs["http.env."+k] = v
			}
		}
	}

	if event.SDK.Name != "" {
		span.Attrs[attrkey.TelemetrySDKName] = event.SDK.Name
	}
	if event.SDK.Version != "" {
		span.Attrs[attrkey.TelemetrySDKVersion] = event.SDK.Version
	}
	if event.Platform != "" {
		span.Attrs[attrkey.TelemetrySDKLanguage] = event.Platform
	}
	if len(event.SDK.Integrations) > 0 {
		// ignore
	}
	if len(event.SDK.Packages) > 0 {
		// ignore
	}

	breadcrumbs, err := h.decodeBreadcrumbs(event.Breadcrumbs)
	if err != nil {
		return err
	}

	for i := range breadcrumbs {
		bc := &breadcrumbs[i]
		span.Events = append(span.Events, h.newSpanFromBreadcrumb(bc))
	}

	exceptions, err := h.decodeExceptions(event.Exception)
	if err != nil {
		return err
	}

	if len(exceptions) > 0 {
		if span.EventName == "" {
			span.EventName = otelEventException
		}
		exc := &exceptions[0]
		if exc.Type != "" {
			span.Attrs[attrkey.ExceptionType] = exc.Type
		}
		if exc.Value != "" {
			span.Attrs[attrkey.ExceptionMessage] = exc.Value
		}
		if exc.Stacktrace != nil {
			if stacktrace := exc.Stacktrace.String(); stacktrace != "" {
				span.Attrs[attrkey.ExceptionStacktrace] = stacktrace
			}
		}
	}

	return nil
}

func (h *SentryHandler) decodeBreadcrumbs(b []byte) ([]SentryBreadcrumb, error) {
	if len(b) <= 2 {
		return nil, nil
	}

	if b[0] == '{' && b[len(b)-1] == '}' {
		var in struct {
			Values []SentryBreadcrumb `json:"values"`
		}
		if err := json.Unmarshal(b, &in); err != nil {
			return nil, err
		}
		return in.Values, nil
	}

	var exceptions []SentryBreadcrumb
	if err := json.Unmarshal(b, &exceptions); err != nil {
		return nil, err
	}
	return exceptions, nil
}

func (h *SentryHandler) newSpanFromBreadcrumb(bc *SentryBreadcrumb) *SpanEvent {
	event := new(SpanEvent)
	event.Time = bc.Timestamp.Time
	event.Attrs = bc.Data

	if bc.Type != "" {
		if event.Attrs == nil {
			event.Attrs = make(AttrMap)
		}
		event.Attrs["breadcrumb.type"] = bc.Type
	}

	if bc.Level != "" && bc.Message != "" {
		event.Name = bc.Level + " " + bc.Message
		return event
	}

	if bc.Category != "" {
		event.Name = bc.Category
	} else {
		event.Name = bc.Message
	}
	return event
}

func (h *SentryHandler) decodeExceptions(b []byte) ([]SentryException, error) {
	if len(b) <= 2 {
		return nil, nil
	}

	if b[0] == '{' && b[len(b)-1] == '}' {
		var in struct {
			Values []SentryException `json:"values"`
		}
		if err := json.Unmarshal(b, &in); err != nil {
			return nil, err
		}
		return in.Values, nil
	}

	var exceptions []SentryException
	if err := json.Unmarshal(b, &exceptions); err != nil {
		return nil, err
	}
	return exceptions, nil
}

func (h *SentryHandler) Envelope(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	project, err := h.projectFromRequest(req)
	if err != nil {
		return err
	}

	rd := bufio.NewReader(req.Body)

	b, err := rd.ReadBytes('\n')
	if err != nil {
		return err
	}

	header := new(SentryEnvelopeHeader)
	if err := json.Unmarshal(b, header); err != nil {
		return err
	}

	for {
		b, err := rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		header := new(SentryItemHeader)
		if err := json.Unmarshal(b, header); err != nil {
			return err
		}

		b, err = rd.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		switch header.Type {
		case "transaction":
			event := new(SentryEvent)
			if err := json.Unmarshal(b, &event); err != nil {
				return err
			}

			if err := h.processTransaction(ctx, project, event); err != nil {
				return err
			}
		case "event":
			event := new(SentryEvent)
			if err := json.Unmarshal(b, &event); err != nil {
				return err
			}

			if err := h.processEvent(ctx, project, event); err != nil {
				return err
			}
		case "client_report":
			// ignore
		default:
			h.Zap(ctx).Error("sentry: unsupported item type", zap.String("type", header.Type))
		}
	}

	return nil
}

func (h *SentryHandler) processTransaction(
	ctx context.Context, project *org.Project, event *SentryEvent,
) error {
	span := new(Span)
	span.ProjectID = project.ID

	trace, ok := event.Contexts["trace"]
	if !ok {
		h.Zap(ctx).Error("sentry: trace context is missing (transaction dropped)")
		return nil
	}
	delete(event.Contexts, "trace")

	traceID, err := uuid.Parse(getString(trace, "trace_id"))
	if err != nil {
		return err
	}

	spanID, err := parseSpanID(getString(trace, "span_id"))
	if err != nil {
		return err
	}

	if err := h.spanFromEvent(span, event); err != nil {
		return err
	}

	for i := range event.Spans {
		src := event.Spans[i]

		dest := new(Span)
		dest.ProjectID = span.ProjectID
		dest.Attrs = span.Attrs.Clone()

		traceID, err := uuid.Parse(src.TraceID)
		if err != nil {
			return err
		}
		dest.TraceID = traceID

		spanID, err := parseSpanID(src.SpanID)
		if err != nil {
			return err
		}
		dest.ID = spanID

		parentSpanID, err := parseSpanID(src.ParentSpanID)
		if err != nil {
			return err
		}
		dest.ParentID = parentSpanID

		dest.Name = src.Op
		dest.Time = src.StartTime.Time
		dest.Duration = src.EndTime.Sub(src.StartTime.Time)

		forEachKV(src.Data, "", func(k string, v any) {
			dest.Attrs[k] = v
		})
		for k, v := range src.Tags {
			k = attrkey.Clean(k)
			if k != "" {
				dest.Attrs["tags."+k] = v
			}
		}

		h.sp.AddSpan(ctx, dest)
	}

	span.ID = spanID
	span.TraceID = traceID
	span.Name = getString(trace, "op")
	span.Time = event.StartTime.Time
	span.Duration = event.Timestamp.Sub(event.StartTime.Time)

	if event.Transaction != "" {
		if span.Name == "" {
			span.Name = event.Transaction
		} else {
			span.Attrs["display.name"] = event.Transaction
		}
	}

	h.sp.AddSpan(ctx, span)

	return nil
}

func (h *SentryHandler) projectFromRequest(req bunrouter.Request) (*org.Project, error) {
	ctx := req.Context()

	sentryKey, err := h.sentryKey(req)
	if err != nil {
		return nil, err
	}

	return org.SelectProjectByToken(ctx, h.App, sentryKey)
}

func (h *SentryHandler) sentryKey(req bunrouter.Request) (string, error) {
	if s := req.URL.Query().Get("sentry_key"); s != "" {
		return s, nil
	}

	auth := req.Header.Get("X-Sentry-Auth")
	if auth == "" {
		return "", errors.New("sentry: X-Sentry-Auth header can't be empty")
	}

	var sentryKey string

	auth = strings.TrimPrefix(auth, "Sentry ")
	for _, kv := range strings.Split(auth, ",") {
		kv = strings.Trim(kv, " ")
		const prefix = "sentry_key="
		if strings.HasPrefix(kv, prefix) {
			sentryKey = strings.TrimPrefix(kv, prefix)
			break
		}
	}

	if sentryKey == "" {
		return "", fmt.Errorf("sentry: can't find sentry_key in %q", auth)
	}

	return sentryKey, nil
}

func getString(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

type SentryEnvelopeHeader struct {
	DSN string `json:"dsn"`
}

type SentryItemHeader struct {
	Type   string `json:"type"`
	Length int64  `json:"length"`
}

type SentryEvent struct {
	Breadcrumbs json.RawMessage           `json:"breadcrumbs"`
	Contexts    map[string]map[string]any `json:"contexts"`
	Dist        string                    `json:"dist"`
	Environment string                    `json:"environment"`
	EventID     string                    `json:"event_id"`
	Extra       map[string]any            `json:"extra"`
	Fingerprint []string                  `json:"fingerprint"`
	Level       string                    `json:"level"`
	Message     string                    `json:"message"`
	Platform    string                    `json:"platform"`
	Release     string                    `json:"release"`
	SDK         struct {
		Name         string          `json:"name"`
		Version      string          `json:"version"`
		Integrations []string        `json:"integrations"`
		Packages     []SentryPackage `json:"packages"`
	} `json:"sdk"`
	ServerName  string            `json:"server_name"`
	Threads     []SentryThread    `json:"threads"`
	Tags        map[string]string `json:"tags"`
	Timestamp   SentryTime        `json:"timestamp"`
	Transaction string            `json:"transaction"`
	User        struct {
		ID        string            `json:"id"`
		Email     string            `json:"email"`
		IPAddress string            `json:"ip_address"`
		Username  string            `json:"username"`
		Name      string            `json:"name"`
		Segment   string            `json:"segment"`
		Data      map[string]string `json:"data"`
	} `json:"user"`
	Logger    string            `json:"logger"`
	Modules   map[string]string `json:"modules"`
	Request   *SentryRequest    `json:"request"`
	Exception json.RawMessage   `json:"exception"`
	// DebugMeta *DebugMeta        `json:"debug_meta"`

	// The fields below are only relevant for transactions.

	Type            string       `json:"type"`
	StartTime       SentryTime   `json:"start_timestamp"`
	Spans           []SentrySpan `json:"spans"`
	TransactionInfo struct {
		Source string `json:"source"`
	} `json:"transaction_info"`
}

type SentryBreadcrumb struct {
	Type      string         `json:"type"`
	Category  string         `json:"category"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data"`
	Level     string         `json:"level"`
	Timestamp SentryTime     `json:"timestamp"`
}

type SentryPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SentryThread struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Stacktrace *SentryStacktrace `json:"stacktrace"`
	Crashed    bool              `json:"crashed"`
	Current    bool              `json:"current"`
}

type SentryRequest struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Data        string            `json:"data"`
	QueryString string            `json:"query_string"`
	Cookies     string            `json:"cookies"`
	Headers     map[string]string `json:"headers"`
	Env         map[string]string `json:"env"`
}

type SentryException struct {
	Type       string            `json:"type"`  // used as the main issue title
	Value      string            `json:"value"` // used as the main issue subtitle
	Module     string            `json:"module"`
	ThreadID   string            `json:"thread_id"`
	Stacktrace *SentryStacktrace `json:"stacktrace"`
	Mechanism  *SentryMechanism  `json:"mechanism"`
}

type SentryStacktrace struct {
	Frames        []SentryFrame `json:"frames"`
	FramesOmitted []uint        `json:"frames_omitted"`
}

func (s *SentryStacktrace) String() string {
	b := make([]byte, 0, 40*len(s.Frames))
	for i := range s.Frames {
		b = s.Frames[i].AppendString(b)
	}
	return unsafeconv.String(b)
}

type SentryFrame struct {
	Function string `json:"function"`
	Symbol   string `json:"symbol"`
	// Module is, despite the name, the Sentry protocol equivalent of a Go
	// package's import path.
	Module      string         `json:"module"`
	Filename    string         `json:"filename"`
	AbsPath     string         `json:"abs_path"`
	Lineno      int            `json:"lineno"`
	Colno       int            `json:"colno"`
	PreContext  []string       `json:"pre_context"`
	ContextLine string         `json:"context_line"`
	PostContext []string       `json:"post_context"`
	InApp       bool           `json:"in_app"`
	Vars        map[string]any `json:"vars"`
	// Package and the below are not used for Go stack trace frames.  In
	// other platforms it refers to a container where the Module can be
	// found.  For example, a Java JAR, a .NET Assembly, or a native
	// dynamic library.  They exists for completeness, allowing the
	// construction and reporting of custom event payloads.
	Package         string `json:"package"`
	InstructionAddr string `json:"instruction_addr"`
	AddrMode        string `json:"addr_mode"`
	SymbolAddr      string `json:"symbol_addr"`
	ImageAddr       string `json:"image_addr"`
	Platform        string `json:"platform"`
	StackStart      bool   `json:"stack_start"`
}

func (f *SentryFrame) AppendString(b []byte) []byte {
	if f.Module != "" {
		b = append(b, f.Module...)
		b = append(b, '.')
	}
	if f.Function != "" {
		b = append(b, f.Function...)
	}

	b = append(b, ' ')

	if f.AbsPath != "" {
		b = append(b, f.AbsPath...)
	} else if f.Filename != "" {
		b = append(b, f.Filename...)
	}
	if f.Lineno > 0 {
		b = append(b, ':')
		b = strconv.AppendInt(b, int64(f.Lineno), 10)
	}
	if f.Colno > 0 {
		b = append(b, ':')
		b = strconv.AppendInt(b, int64(f.Colno), 10)
	}

	b = append(b, '\n')

	return b
}

type SentryMechanism struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	HelpLink    string         `json:"help_link"`
	Handled     *bool          `json:"handled"`
	Data        map[string]any `json:"data"`
}

type SentrySpan struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id"`
	Name         string            `json:"name"`
	Op           string            `json:"op"`
	Description  string            `json:"description"`
	Status       uint8             `json:"status"`
	Tags         map[string]string `json:"tags"`
	StartTime    SentryTime        `json:"start_timestamp"`
	EndTime      SentryTime        `json:"timestamp"`
	Data         map[string]any    `json:"data"`
}

type SentryTime struct {
	time.Time
}

func (t *SentryTime) UnmarshalJSON(b []byte) error {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return t.Time.UnmarshalJSON(b)
	}

	var unix float64

	if err := json.Unmarshal(b, &unix); err != nil {
		return err
	}

	secs := int64(unix)
	nanosecs := (unix - float64(secs)) * float64(time.Second)
	t.Time = time.Unix(secs, int64(nanosecs))

	return nil
}

func forEachKV(m map[string]any, prefix string, fn func(string, any)) {
	for k, v := range m {
		k = attrkey.Clean(k)
		if k == "" {
			continue
		}

		switch v := v.(type) {
		case map[string]any:
			forEachKV(v, k+".", fn)
		default:
			fn(k, v)
		}
	}
}
