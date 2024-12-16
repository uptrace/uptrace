package org

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	"go.uber.org/fx"
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

type MiddlewareParams struct {
	fx.In

	Logger   *otelzap.Logger
	Conf     *bunconf.Config
	PG       *bun.DB
	Projects *ProjectGateway
}

type Middleware struct {
	*MiddlewareParams

	userProviders []UserProvider
}

func NewMiddleware(p MiddlewareParams) *Middleware {
	var userProviders []UserProvider

	if len(p.Conf.Auth.Users) > 0 || len(p.Conf.Auth.OIDC) > 0 {
		userProviders = append(userProviders, NewJWTProvider(p.Conf.SecretKey))
	}
	for _, cloudflare := range p.Conf.Auth.Cloudflare {
		userProviders = append(userProviders, NewCloudflareProvider(cloudflare))
	}

	return &Middleware{
		MiddlewareParams: &p,
		userProviders:    userProviders,
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

		project, err := m.projectFromRequest(req)
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
				m.Logger.Error("provider.Auth failed", zap.Error(err))
			}
			continue
		}

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetAttributes(
				semconv.EnduserIDKey.String(user.Email),
			)
		}

		if err := GetOrCreateUser(ctx, m.PG, user); err != nil {
			m.Logger.Error("GetOrCreateUser failed", zap.Error(err))
			continue
		}

		return user, nil
	}

	token, err := tokenFromRequest(req)
	if err != nil {
		return nil, httperror.Unauthorized(err.Error())
	}

	user, err := SelectUserByToken(ctx, m.PG, token)
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

func (m *Middleware) projectFromRequest(req bunrouter.Request) (*Project, error) {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return nil, err
	}

	project, err := m.Projects.SelectByID(ctx, projectID)
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
