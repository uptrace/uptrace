package tracing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/koding/websocketproxy"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
	"go.uber.org/zap"
)

type LokiProxyHandler struct {
	GrafanaBaseHandler

	sp *SpanProcessor

	proxy   *httputil.ReverseProxy
	wsProxy *websocketproxy.WebsocketProxy
}

func NewLokiProxyHandler(app *bunapp.App, sp *SpanProcessor) *LokiProxyHandler {
	h := &LokiProxyHandler{
		GrafanaBaseHandler: GrafanaBaseHandler{
			App: app,
		},
		sp: sp,
	}
	h.initProxy()
	h.initWSProxy()
	return h
}

func (h *LokiProxyHandler) initProxy() {
	lokiURL, _ := url.Parse(h.App.Config().Loki.Addr)
	lokiQuery := lokiURL.RawQuery

	errorLogger, _ := zap.NewStdLogAt(h.ZapLogger().Logger, zap.WarnLevel)
	h.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = lokiURL.Scheme
			req.URL.Host = lokiURL.Host
			req.URL.Path, req.URL.RawPath = joinURLPath(lokiURL, req.URL)
			if lokiQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = lokiQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = lokiQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}

			project := org.ProjectFromContext(req.Context())
			req.Header.Set("uptrace-project-id", strconv.Itoa(int(project.ID)))

			req.Header.Del("Origin")
		},
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Del("access-control-allow-origin")
			return nil
		},
		ErrorLog: errorLogger,
	}
}

func (h *LokiProxyHandler) initWSProxy() {
	u, _ := url.Parse(h.Config().Loki.Addr)

	h.wsProxy = &websocketproxy.WebsocketProxy{
		// Director: func(req *http.Request, out http.Header) {
		// 	if project, _ := org.ProjectFromContext(req.Context()); project != nil {
		// 		out.Set("uptrace-project-id", strconv.Itoa(int(project.ID)))
		// 	}
		// },
		Backend: func(req *http.Request) *url.URL {
			clone := *u
			u := &clone

			project := org.ProjectFromContext(req.Context())

			query := u.Query()
			query.Set("uptrace-project-id", strconv.Itoa(int(project.ID)))
			u.RawQuery = query.Encode()

			return u
		},
	}
}

type LokiStream struct {
	Stream xotel.AttrMap     `json:"stream"`
	Values []LokiStreamValue `json:"values"`
}

type LokiStreamValue []string

func (h *LokiProxyHandler) Push(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(req.Context())

	switch req.Header.Get("content-type") {
	case "application/json", "":
		// continue
	default:
		h.proxy.ServeHTTP(w, req.Request)
		return nil
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(b))

	if err := h.processStreams(ctx, project, b); err != nil {
		h.Zap(ctx).Error("processStreams failed", zap.Error(err))
	}

	h.proxy.ServeHTTP(w, req.Request)
	return nil
}

func (h *LokiProxyHandler) processStreams(
	ctx context.Context, project *bunapp.Project, data []byte,
) error {
	var in struct {
		Streams []LokiStream `json:"streams"`
	}

	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	p := &lokiLogProcessor{
		ctx:     ctx,
		app:     h.App,
		sp:      h.sp,
		project: project,
	}
	defer p.close()

	for i := range in.Streams {
		stream := &in.Streams[i]
		spans := make([]Span, len(stream.Values))
		for i, value := range stream.Values {
			if err := p.processLogValue(&spans[i], stream.Stream, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *LokiProxyHandler) Proxy(w http.ResponseWriter, req bunrouter.Request) error {
	h.proxy.ServeHTTP(w, req.Request)
	return nil
}

func (h *LokiProxyHandler) ProxyWS(w http.ResponseWriter, req bunrouter.Request) error {
	h.wsProxy.ServeHTTP(w, req.Request)
	return nil
}

//------------------------------------------------------------------------------

type lokiLogProcessor struct {
	ctx context.Context
	app *bunapp.App

	sp      *SpanProcessor
	project *bunapp.Project

	logger *otelzap.Logger
}

func (p *lokiLogProcessor) close() {}

func (p *lokiLogProcessor) processLogValue(
	span *Span, resource xotel.AttrMap, value LokiStreamValue,
) error {
	if len(value) != 2 {
		return fmt.Errorf("got %d values, expected 2", len(value))
	}

	span.ID = rand.Uint64()
	span.ProjectID = p.project.ID
	span.Kind = internalSpanKind
	span.EventName = logEventType
	span.StatusCode = okStatusCode
	span.Attrs = resource.Clone()

	ts, err := strconv.ParseInt(value[0], 10, 64)
	if err != nil {
		return err
	}

	span.Time = time.Unix(0, ts)
	span.Attrs[xattr.LogMessage] = value[1]

	p.sp.AddSpan(span)

	return nil
}

//------------------------------------------------------------------------------

// joinURLPath from golang.org/src/net/http/httputil/reverseproxy.go
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

// singleJoiningSlash from golang.org/src/net/http/httputil/reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
