package org

import (
	"context"
	"net/http"

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
	api := app.APIGroup()

	api.WithGroup("/users", func(g *bunrouter.Group) {
		userHandler := NewUserHandler(app)

		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)

		g = g.Use(middleware.User)

		g.GET("/current", userHandler.Current)
	})

	g.WithGroup("/sso", func(g *bunrouter.Group) {
		ssoHandler := NewSSOHandler(app, g)

		g.GET("/methods", ssoHandler.ListMethods)
	})

	g.GET("/projects/:project_id", func(w http.ResponseWriter, req bunrouter.Request) error {
		projectID, err := req.Params().Uint32("project_id")
		if err != nil {
			return err
		}

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
		WithGroup("/pinned-facets", func(g *bunrouter.Group) {
			handler := NewPinnedFacetHandler(app)

			g.GET("", handler.List)
			g.POST("", handler.Add)
			g.DELETE("", handler.Remove)
		})
}
