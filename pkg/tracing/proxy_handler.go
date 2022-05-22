package tracing

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
)

type LokiProxyHandler struct {
	*bunapp.App
	proxy *httputil.ReverseProxy
}

func NewLokiProxyHandler(app *bunapp.App) *LokiProxyHandler {
	lokiAddr, _ := url.Parse(app.Config().Loki.Addr)

	lokiQuery := lokiAddr.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = lokiAddr.Scheme
		req.URL.Host = lokiAddr.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(lokiAddr, req.URL)
		if lokiQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = lokiQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = lokiQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	errorLogger, _ := zap.NewStdLogAt(app.ZapLogger().Logger, zap.ErrorLevel)
	return &LokiProxyHandler{
		App: app,
		proxy: &httputil.ReverseProxy{
			Director: director,
			ErrorLog: errorLogger,
		},
	}
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

func (h *LokiProxyHandler) Proxy(w http.ResponseWriter, req bunrouter.Request) error {
	h.proxy.ServeHTTP(w, req.Request)
	return nil
}
