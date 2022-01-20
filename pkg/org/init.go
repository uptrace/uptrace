package org

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

func init() {
	bunapp.OnStart("tracing.registerRoutes", registerRoutes)
}

func registerRoutes(ctx context.Context, app *bunapp.App) error {
	userHandler := NewUserHandler(app)

	g := app.APIGroup()

	g.WithGroup("/users", func(g *bunrouter.Group) {
		g.POST("/login", userHandler.Login)
		g.POST("/logout", userHandler.Logout)
		g.GET("/current", userHandler.Current)
	})

	return nil
}
