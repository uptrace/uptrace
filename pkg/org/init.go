package org

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

func Init(ctx context.Context, app *bunapp.App) {
	registerRoutes(ctx, app)
}

func registerRoutes(ctx context.Context, app *bunapp.App) {
	middleware := NewMiddleware(app)
	userHandler := NewUserHandler(app)

	g := app.APIGroup()

	g.WithGroup("/users", func(g *bunrouter.Group) {
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

		project, err := SelectProject(ctx, app, projectID)
		if err != nil {
			return err
		}

		return httputil.JSON(w, bunrouter.H{
			"project": project,
			"grpc": bunrouter.H{
				"endpoint": app.Config().GRPCEndpoint(),
				"dsn":      app.Config().GRPCDsn(project),
			},
			"http": bunrouter.H{
				"endpoint": app.Config().HTTPEndpoint(),
				"dsn":      app.Config().HTTPDsn(project),
			},
		})
	})
}
