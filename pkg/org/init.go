package org

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/vmihailenco/taskq/v4"
	"go.uber.org/fx"
)

var (
	ErrUnauthorized    = httperror.Unauthorized("please log in")
	ErrAccessDenied    = httperror.Forbidden("access denied")
	ErrProjectNotFound = httperror.NotFound("project not found")
)

var CreateErrorAlertTask = taskq.NewTask("create-error-alert")

type TrackableModel string

const (
	ModelUser      TrackableModel = "User"
	ModelProject   TrackableModel = "Project"
	ModelSpan      TrackableModel = "Span"
	ModelSpanGroup TrackableModel = "SpanGroup"
)

type OrgParams struct {
	fx.In

	Conf   *bunconf.Config
	Logger *otelzap.Logger
	PG     *bun.DB
	CH     *ch.DB
}

type OrgRunParams struct {
	fx.In

	Org    *Org
	Router *bunapp.Router
}

type Org struct {
	conf   *bunconf.Config
	logger *otelzap.Logger
	PG     *bun.DB
	CH     *ch.DB
}

func NewOrg(p OrgParams) *Org {
	return &Org{
		conf:   p.Conf,
		logger: p.Logger,
		PG:     p.PG,
		CH:     p.CH,
	}
}

func Init(params OrgRunParams) {
	registerRoutes(params)
}

func registerRoutes(p OrgRunParams) {
	fakeApp := &bunapp.App{
		Conf:   p.Org.conf,
		PG:     p.Org.PG,
		CH:     p.Org.CH,
		Logger: p.Org.logger,
	}
	middleware := NewMiddleware(fakeApp)
	api := p.Router.InternalV1

	api.WithGroup("/users", func(g *bunrouter.Group) {
		userHandler := NewUserHandler(p.Org)

		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)

		g = g.Use(middleware.User)

		g.GET("/current", userHandler.Current)
	})

	api.WithGroup("/sso", func(g *bunrouter.Group) {
		ssoHandler := NewSSOHandler(p.Org, g)

		g.GET("/methods", ssoHandler.ListMethods)
	})

	api.
		Use(middleware.User).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			projectHandler := NewProjectHandler(p.Org)

			g.GET("", projectHandler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			handler := NewUsageHandler(p.Org)

			g.GET("/data-usage", handler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("/pinned-facets", func(g *bunrouter.Group) {
			handler := NewPinnedFacetHandler(p.Org)

			g.GET("", handler.List)
			g.POST("", handler.Add)
			g.DELETE("", handler.Remove)
		})

	api.Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			handler := NewAchievementHandler(p.Org)

			g.GET("/achievements", handler.List)
		})

	{
		handler := NewAnnotationHandler(p.Org)
		api.POST("/annotations", handler.CreatePublic)

		api.Use(middleware.UserAndProject).
			WithGroup("/projects/:project_id/annotations", func(g *bunrouter.Group) {
				g.GET("", handler.List)
				g.POST("", handler.Create)

				g = g.Use(handler.AnnotationMiddleware)
				g.GET("/:annotation_id", handler.Show)
				g.PUT("/:annotation_id", handler.Update)
				g.DELETE("/:annotation_id", handler.Delete)
			})
	}
}
