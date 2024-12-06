package grafana

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/fx"
)

const (
	protobufContentType = "application/protobuf"
	jsonContentType     = "application/json"
)

type ModuleParams struct {
	fx.In
	bunapp.RouterParams

	Conf   *bunconf.Config
	Logger *otelzap.Logger
	PG     *bun.DB
	CH     *ch.DB
}

func Init(p ModuleParams) {
	initRoutes(&p)
}

func initRoutes(p *ModuleParams) {
	fakeApp := &bunapp.App{
		Conf:   p.Conf,
		Logger: p.Logger,
		PG:     p.PG,
		CH:     p.CH,
	}
	middleware := org.NewMiddleware(fakeApp)

	// https://grafana.com/docs/tempo/latest/api_docs/
	p.Router.WithGroup("/api/tempo/:project_id", func(g *bunrouter.Group) {
		tempoHandler := NewTempoHandler(p.Logger, p.Conf, p.PG, p.CH)

		g = g.Use(middleware.UserAndProject)

		g.GET("/ready", tempoHandler.Ready)
		g.GET("/api/echo", tempoHandler.Echo)
		g.GET("/api/status/buildinfo", tempoHandler.BuildInfo)

		g.GET("/api/traces/:trace_id", tempoHandler.QueryTrace)
		g.GET("/api/traces/:trace_id/json", tempoHandler.QueryTraceJSON)

		g.GET("/api/search", tempoHandler.Search)

		g.GET("/api/v2/search/tags", tempoHandler.Tags)
		g.GET("/api/v2/search/tag/:tag/values", tempoHandler.TagValues)
	})

	p.Router.WithGroup("/api/prometheus/:project_id", func(g *bunrouter.Group) {
		promHandler := NewPromHandler(p.Logger, p.Conf, p.PG, p.CH)

		g = g.Use(
			middleware.UserAndProject,
			promHandler.EnablePromCompat,
			promErrorHandler,
		)

		g.GET("/api/v1/metadata", promHandler.Metadata)
		g.GET("/api/v1/labels", promHandler.LabelNames)
		g.POST("/api/v1/labels", promHandler.LabelNames)
		g.GET("/api/v1/label/:label/values", promHandler.LabelValues)
		g.POST("/api/v1/query_range", promHandler.QueryRange)
		g.GET("/api/v1/query_range", promHandler.QueryRange)
		g.POST("/api/v1/query", promHandler.QueryInstant)
		g.GET("/api/v1/query", promHandler.QueryInstant)
		g.GET("/api/v1/series", promHandler.Series)
		g.POST("/api/v1/series", promHandler.Series)
	})
}

func promErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}
		switch err := err.(type) {
		case *promError:
			return err
		default:
			return newPromError(err)
		}
	}
}

//------------------------------------------------------------------------------

type promError struct {
	Wrapped error `json:"error"`
}

var _ httperror.Error = (*promError)(nil)

func newPromError(err error) *promError {
	return &promError{
		Wrapped: err,
	}
}

func (e *promError) Error() string {
	if e.Wrapped == nil {
		return ""
	}
	return e.Wrapped.Error()
}

func (e *promError) Unwrap() error {
	return e.Wrapped
}

func (e *promError) HTTPStatusCode() int {
	return http.StatusBadRequest
}

func (e *promError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"status":    "error",
		"errorType": "error",
		"error":     e.Error(),
	})
}
