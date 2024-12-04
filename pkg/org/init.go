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

type ModuleParams struct {
	fx.In
	bunapp.RouterParams

	Conf   *bunconf.Config
	Logger *otelzap.Logger
	PG     *bun.DB
	CH     *ch.DB
}

func Init(params ModuleParams) {
	registerRoutes(params)
}

func registerRoutes(p ModuleParams) {
	fakeApp := &bunapp.App{
		Conf:   p.Conf,
		Logger: p.Logger,
		PG:     p.PG,
		CH:     p.CH,
	}
	middleware := NewMiddleware(fakeApp)
	api := p.RouterParams.RouterInternalV1

	api.WithGroup("/users", func(g *bunrouter.Group) {
		userHandler := NewUserHandler(p.Conf, p.Logger, p.PG)

		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)

		g = g.Use(middleware.User)

		g.GET("/current", userHandler.Current)
	})

	api.WithGroup("/sso", func(g *bunrouter.Group) {
		ssoHandler := NewSSOHandler(p.Conf, p.Logger, p.PG, g)

		g.GET("/methods", ssoHandler.ListMethods)
	})

	api.
		Use(middleware.User).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			projectHandler := NewProjectHandler(p.Conf, p.Logger, p.PG)

			g.GET("", projectHandler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			handler := NewUsageHandler(p.Conf, p.Logger, p.CH)

			g.GET("/data-usage", handler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("/pinned-facets", func(g *bunrouter.Group) {
			handler := NewPinnedFacetHandler(p.Conf, p.Logger, p.PG)

			g.GET("", handler.List)
			g.POST("", handler.Add)
			g.DELETE("", handler.Remove)
		})

	api.Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			handler := NewAchievementHandler(p.PG)

			g.GET("/achievements", handler.List)
		})

	{
		handler := NewAnnotationHandler(p.PG)
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
