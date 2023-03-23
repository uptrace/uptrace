package org

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
)

const (
	tokenCookieName = "token"
	tokenTTL        = 7 * 24 * time.Hour
)

var (
	ErrUnauthorized = httperror.Unauthorized("please log in")
	ErrAccessDenied = httperror.Forbidden("access denied")
)

var AnonymousUser = &bunconf.User{
	Username: "anonymous",
}

type (
	userCtxKey    struct{}
	projectCtxKey struct{}
)

func init() {
	AnonymousUser.Init()
}

func UserFromContext(ctx context.Context) *bunconf.User {
	user := ctx.Value(userCtxKey{}).(*bunconf.User)
	return user
}

func ContextWithUser(ctx context.Context, user *bunconf.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func ProjectFromContext(ctx context.Context) *bunconf.Project {
	project := ctx.Value(projectCtxKey{}).(*bunconf.Project)
	return project
}

func ContextWithProject(ctx context.Context, project *bunconf.Project) context.Context {
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
		userProviders = append(userProviders, NewJWTProvider(conf.SecretKey))
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

func (m *Middleware) userFromRequest(req bunrouter.Request) *bunconf.User {
	ctx := req.Context()

	if len(m.userProviders) == 0 {
		return AnonymousUser
	}

	for _, provider := range m.userProviders {
		user, err := provider.Auth(req)
		if err != nil {
			if err != errNoUser {
				m.app.Zap(ctx).Error("Auth failed", zap.Error(err))
			}
			continue
		}

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetAttributes(
				semconv.EnduserIDKey.String(user.Username),
			)
		}

		return user
	}

	return nil
}

func ProjectFromRequest(app *bunapp.App, req bunrouter.Request) (*bunconf.Project, error) {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return nil, err
	}

	project, err := SelectProject(ctx, app, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	return project, nil
}

//------------------------------------------------------------------------------

func findUserByPassword(app *bunapp.App, username string, password string) *bunconf.User {
	users := app.Config().Auth.Users
	for i := range users {
		user := &users[i]
		if subtle.ConstantTimeCompare([]byte(user.Username), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) == 1 {
			return user
		}
	}
	return nil
}
