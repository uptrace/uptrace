package org

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
)

const (
	tokenCookieName = "token"
	tokenTTL        = 7 * 24 * time.Hour
)

type (
	userCtxKey    struct{}
	projectCtxKey struct{}
)

func UserFromContext(ctx context.Context) *User {
	user := ctx.Value(userCtxKey{}).(*User)
	return user
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func ProjectFromContext(ctx context.Context) *Project {
	project := ctx.Value(projectCtxKey{}).(*Project)
	return project
}

func ContextWithProject(ctx context.Context, project *Project) context.Context {
	return context.WithValue(ctx, projectCtxKey{}, project)
}

type Middleware struct {
	app *bunapp.App

	userProviders []UserProvider
}

func NewMiddleware(app *bunapp.App) *Middleware {
	var userProviders []UserProvider

	conf := app.Config()

	if len(conf.Auth.Users) > 0 || len(conf.Auth.OIDC) > 0 {
		userProviders = append(userProviders, NewJWTProvider(app, conf.SecretKey))
	}
	for _, cloudflare := range conf.Auth.Cloudflare {
		userProviders = append(userProviders, NewCloudflareProvider(cloudflare))
	}

	return &Middleware{
		app:           app,
		userProviders: userProviders,
	}
}

func (m *Middleware) User(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		user := m.userFromRequest(req)
		if user == nil {
			return ErrUnauthorized
		}
		ctx = ContextWithUser(ctx, user)

		return next(w, req.WithContext(ctx))
	}
}

func (m *Middleware) UserAndProject(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		user := m.userFromRequest(req)
		if user == nil {
			return ErrUnauthorized
		}
		ctx = ContextWithUser(ctx, user)

		project, err := ProjectFromRequest(m.app, req)
		if err != nil {
			return err
		}
		ctx = ContextWithProject(ctx, project)

		return next(w, req.WithContext(ctx))
	}
}

func (m *Middleware) userFromRequest(req bunrouter.Request) *User {
	ctx := req.Context()

	if len(m.userProviders) == 0 {
		return nil
	}

	for _, provider := range m.userProviders {
		user, err := provider.Auth(req)
		if err != nil {
			if err != errNoUser {
				m.app.Zap(ctx).Error("provider.Auth failed", zap.Error(err))
			}
			continue
		}

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetAttributes(
				semconv.EnduserIDKey.String(user.Email),
			)
		}

		if err := GetOrCreateUser(ctx, m.app, user); err != nil {
			m.app.Zap(ctx).Error("GetOrCreateUser failed", zap.Error(err))
			continue
		}

		return user
	}

	return nil
}

func ProjectFromRequest(app *bunapp.App, req bunrouter.Request) (*Project, error) {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return nil, err
	}

	project, err := SelectProject(ctx, app, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return project, nil
}
