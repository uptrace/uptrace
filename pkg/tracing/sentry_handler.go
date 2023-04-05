package tracing

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/uuid"
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

	span := new(Span)
	span.ProjectID = project.ID
	span.EventName = otelEventLog

	if err := h.spanFromEvent(span, event); err != nil {
		return err
	}

	traceID, err := uuid.Parse(event.EventID)
	if err != nil {
		return err
	}
	span.TraceID = traceID

	if event.Level != "" {
		span.Attrs[attrkey.LogSeverity] = event.Level
	}
	if event.Message != "" {
		span.Attrs[attrkey.LogMessage] = event.Message
	}

	h.sp.AddSpan(ctx, span)

	return nil
}

func (h *SentryHandler) spanFromEvent(span *Span, event *SentryEvent) error {
	span.Time = event.Timestamp
	span.Attrs = make(AttrMap, len(event.Tags)+1)

	for k, v := range event.Contexts {
		span.Attrs["contexts."+k] = v
	}
	for k, v := range event.Tags {
		span.Attrs["tags."+k] = v
	}
	for k, v := range event.Extra {
		span.Attrs["extra."+k] = v
	}

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
		span.Attrs["enduser.data."+k] = v
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
			span.Attrs["http.headers."+k] = v
		}
		for k, v := range req.Env {
			span.Attrs["http.env."+k] = v
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

	return nil
}

func (h *SentryHandler) Envelope(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	project, err := h.projectFromRequest(req)
	if err != nil {
		return err
	}

	rd := bufio.NewReader(req.Body)

	header := new(SentryEnvelopeHeader)
	if err := decodeNextLine(rd, &header); err != nil {
		return err
	}

	for {
		header := new(SentryItemHeader)
		if err := decodeNextLine(rd, header); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		b, err := rd.ReadBytes('\n')
		if err != nil {
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
		default:
			// ignore
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
		dest.Time = src.StartTime
		dest.Duration = src.EndTime.Sub(src.StartTime)

		for k, v := range src.Tags {
			dest.Attrs[k] = v
		}

		h.sp.AddSpan(ctx, dest)
	}

	span.ID = spanID
	span.TraceID = traceID
	span.Name = getString(trace, "op")
	span.Time = event.StartTime
	span.Duration = event.Timestamp.Sub(event.StartTime)

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

	auth := req.Header.Get("X-Sentry-Auth")
	if auth == "" {
		return nil, errors.New("sentry: X-Sentry-Auth header can't be empty")
	}

	var sentryKey string

	for _, kv := range strings.Split(auth, ", ") {
		const prefix = "sentry_key="
		if strings.HasPrefix(kv, prefix) {
			sentryKey = strings.TrimPrefix(kv, prefix)
			break
		}
	}

	if sentryKey == "" {
		return nil, fmt.Errorf("sentry: can't find sentry_key in %q", auth)
	}

	return org.SelectProjectByToken(ctx, h.App, sentryKey)
}

func decodeNextLine(rd *bufio.Reader, dest any) error {
	b, err := rd.ReadBytes('\n')
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func getString(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

type SentryEnvelopeHeader struct {
	DSN   string `json:"dsn"`
	Trace struct {
		PublicKey   string `json:"public_key"`
		Release     string `json:"release"`
		SampleRate  string `json:"sample_rate"`
		TraceID     string `json:"trace_id"`
		Transaction string `json:"transaction"`
	} `json:"trace"`
}

type SentryItemHeader struct {
	Type   string `json:"type"`
	Length int64  `json:"length"`
}

type SentryEvent struct {
	Breadcrumbs []*SentryBreadcrumb       `json:"breadcrumbs"`
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
	Timestamp   time.Time         `json:"timestamp"`
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
	Exception []SentryException `json:"exception"`
	// DebugMeta *DebugMeta        `json:"debug_meta"`

	// The fields below are only relevant for transactions.

	Type            string       `json:"type"`
	StartTime       time.Time    `json:"start_timestamp"`
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
	Timestamp time.Time      `json:"timestamp"`
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
	StartTime    time.Time         `json:"start_timestamp"`
	EndTime      time.Time         `json:"timestamp"`
	Data         map[string]any    `json:"data"`
}
