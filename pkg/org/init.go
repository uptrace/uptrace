package org

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/vmihailenco/taskq/v4"
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

func Init(ctx context.Context, app *bunapp.App) {
	registerRoutes(ctx, app)
}

func registerRoutes(ctx context.Context, app *bunapp.App) {
	middleware := NewMiddleware(app)
	api := app.InternalAPIV1()

	api.WithGroup("/users", func(g *bunrouter.Group) {
		userHandler := NewUserHandler(app)

		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)

		g = g.Use(middleware.User)

		g.GET("/current", userHandler.Current)
	})

	api.WithGroup("/sso", func(g *bunrouter.Group) {
		ssoHandler := NewSSOHandler(app, g)

		g.GET("/methods", ssoHandler.ListMethods)
	})

	api.
		Use(middleware.User).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			projectHandler := NewProjectHandler(app)

			g.GET("", projectHandler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			handler := NewUsageHandler(app)

			g.GET("/data-usage", handler.Show)
		})

	api.
		Use(middleware.User).
		WithGroup("/pinned-facets", func(g *bunrouter.Group) {
			handler := NewPinnedFacetHandler(app)

			g.GET("", handler.List)
			g.POST("", handler.Add)
			g.DELETE("", handler.Remove)
		})

	api.Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			handler := NewAchievementHandler(app)

			g.GET("/achievements", handler.List)
		})

	{
		handler := NewAnnotationHandler(app)
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
