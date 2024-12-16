package bunapp

import (
	"net/http"
	"net/http/pprof"

	"go.opentelemetry.io/otel/trace"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/bunrouterotel"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
)

func initRouter(conf *bunconf.Config) RouterResults {
	router := newRouter(conf)

	var routerGroup *bunrouter.Group
	if conf.Site.URL.Path != "/" {
		routerGroup = router.NewGroup(conf.Site.URL.Path)
	} else {
		routerGroup = router.NewGroup("")
	}

	if conf.Debug {
		adapter := bunrouter.HTTPHandlerFunc

		routerGroup.GET("/debug/pprof/", adapter(pprof.Index))
		routerGroup.GET("/debug/pprof/cmdline", adapter(pprof.Cmdline))
		routerGroup.GET("/debug/pprof/profile", adapter(pprof.Profile))
		routerGroup.GET("/debug/pprof/symbol", adapter(pprof.Symbol))
		routerGroup.GET(
			"/debug/pprof/:name", func(w http.ResponseWriter, req bunrouter.Request) error {
				h := pprof.Handler(req.Param("name"))
				h.ServeHTTP(w, req.Request)
				return nil
			})
	}

	return RouterResults{
		RouterBundle: RouterBundle{
			Router:           router,
			RouterGroup:      routerGroup,
			RouterInternalV1: routerGroup.NewGroup("/internal/v1"),
			RouterPublicV1:   routerGroup.NewGroup("/api/v1"),
		},
	}
}

func newRouter(conf *bunconf.Config, opts ...bunrouter.Option) *bunrouter.Router {
	opts = append(opts,
		bunrouter.WithMiddleware(reqlog.NewMiddleware(
			reqlog.WithVerbose(conf.Debug),
			reqlog.FromEnv("HTTPDEBUG", "DEBUG"),
		)),
	)

	opts = append(opts,
		bunrouter.WithMiddleware(httpErrorHandler),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware(
			bunrouterotel.WithClientIP(),
		)),
	)

	return bunrouter.New(opts...)
}

func httpErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}

		ctx := req.Context()
		httpErr := httperror.From(err)
		statusCode := httpErr.HTTPStatusCode()

		data := map[string]any{
			"statusCode": statusCode,
			"error":      httpErr,
		}

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			if statusCode >= 400 {
				trace.SpanFromContext(ctx).RecordError(err)
			}

			traceID := span.SpanContext().TraceID()
			data["traceId"] = traceID
		}

		w.WriteHeader(statusCode)
		_ = bunrouter.JSON(w, data)

		return err
	}
}
