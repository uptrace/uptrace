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
	userHandler := NewUserHandler(app)

	g := app.APIGroup()

	g.WithGroup("/users", func(g *bunrouter.Group) {
		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)
		g.GET("/current", userHandler.Current)
	})

	g.GET("/projects/:project_id", func(w http.ResponseWriter, req bunrouter.Request) error {
		projectID, err := req.Params().Uint32("project_id")
		if err != nil {
			return err
		}

		project, err := SelectProjectByID(ctx, app, projectID)
		if err != nil {
			return err
		}

		return httputil.JSON(w, bunrouter.H{
			"grpc": bunrouter.H{
				"endpoint": app.Config().GRPCEndpoint(project),
				"dsn":      app.Config().GRPCDsn(project),
			},
			"http": bunrouter.H{
				"endpoint": app.Config().HTTPEndpoint(project),
				"dsn":      app.Config().HTTPDsn(project),
			},
		})
	})
}
