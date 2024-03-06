package org

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	"go.uber.org/zap"
)

var errTokenEmpty = httperror.Unauthorized("token is missing or empty")

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

		user, err := m.userFromRequest(req)
		if err != nil {
			return err
		}
		ctx = ContextWithUser(ctx, user)

		return next(w, req.WithContext(ctx))
	}
}

func (m *Middleware) UserAndProject(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		user, err := m.userFromRequest(req)
		if err != nil {
			return err
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

func (m *Middleware) userFromRequest(req bunrouter.Request) (*User, error) {
	ctx := req.Context()

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

		return user, nil
	}

	token, err := tokenFromRequest(req)
	if err != nil {
		return nil, httperror.Unauthorized(err.Error())
	}

	user, err := SelectUserByToken(ctx, m.app, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.Unauthorized(err.Error())
		}
		return nil, err
	}
	return user, nil
}

func tokenFromRequest(req bunrouter.Request) (string, error) {
	if auth := req.Header.Get("Authorization"); auth != "" {
		const bearerPrefix = "Bearer "
		auth = strings.TrimPrefix(auth, bearerPrefix)
		if auth != "" {
			return auth, nil
		}
	}
	return "", errTokenEmpty
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

func DSNFromRequest(req bunrouter.Request, extraHeaders ...string) (string, error) {
	if dsn := dsnFromRequest(req, extraHeaders); dsn != "" {
		return dsn, nil
	}
	return "", errors.New("uptrace-dsn header is empty or missing")
}

func dsnFromRequest(req bunrouter.Request, extraHeaders []string) string {
	if dsn := req.Header.Get("uptrace-dsn"); dsn != "" {
		return dsn
	}
	if dsn := req.Header.Get("uptrace-dns"); dsn != "" {
		return dsn
	}

	if auth := req.Header.Get("Authorization"); auth != "" {
		const bearer = "Bearer "
		return strings.TrimPrefix(auth, bearer)
	}

	for _, headerKey := range extraHeaders {
		if dsn := req.Header.Get(headerKey); dsn != "" {
			return dsn
		}
	}

	if dsn := req.URL.Query().Get("dsn"); dsn != "" {
		return dsn
	}

	return ""
}

func DSNFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is empty")
	}

	if dsn := dsnFromMetadata(md); dsn != "" {
		return dsn, nil
	}
	return "", errors.New("uptrace-dsn header is empty or missing")
}

func dsnFromMetadata(md metadata.MD) string {
	if dsn := md.Get("uptrace-dsn"); len(dsn) > 0 {
		return dsn[0]
	}
	if dsn := md.Get("uptrace-dns"); len(dsn) > 0 {
		return dsn[0]
	}
	return ""
}
