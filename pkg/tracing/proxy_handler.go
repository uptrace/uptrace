package tracing

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/koding/websocketproxy"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
)

type LokiProxyHandler struct {
	*bunapp.App
	proxy   *httputil.ReverseProxy
	wsProxy *websocketproxy.WebsocketProxy
}

func NewLokiProxyHandler(app *bunapp.App) *LokiProxyHandler {
	h := &LokiProxyHandler{
		App: app,
	}
	h.initProxy()
	h.initWSProxy()
	return h
}

func (h *LokiProxyHandler) initProxy() {
	lokiURL, _ := url.Parse(h.App.Config().Loki.Addr)
	lokiQuery := lokiURL.RawQuery

	errorLogger, _ := zap.NewStdLogAt(h.ZapLogger().Logger, zap.ErrorLevel)
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
			if project, _ := org.ProjectFromContext(req.Context()); project != nil {
				req.Header.Set("uptrace-project-id", strconv.Itoa(int(project.ID)))
			}
			req.Header.Del("Origin")
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

			if project, _ := org.ProjectFromContext(req.Context()); project != nil {
				query := u.Query()
				query.Set("uptrace-project-id", strconv.Itoa(int(project.ID)))
				u.RawQuery = query.Encode()
			}

			return u
		},
	}
}

func (h *LokiProxyHandler) Proxy(w http.ResponseWriter, req bunrouter.Request) error {
	h.proxy.ServeHTTP(w, req.Request)
	return nil
}

func (h *LokiProxyHandler) ProxyWS(w http.ResponseWriter, req bunrouter.Request) error {
	h.wsProxy.ServeHTTP(w, req.Request)
	return nil
}

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
