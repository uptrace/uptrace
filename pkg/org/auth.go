package org

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"

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

var GuestUser = &bunconf.User{
	ID:       1,
	Username: "guest",
}

type (
	userCtxKey    struct{}
	projectCtxKey struct{}
)

func UserFromContext(ctx context.Context) (*bunconf.User, error) {
	user, ok := ctx.Value(userCtxKey{}).(*bunconf.User)
	if !ok {
		return nil, ErrUnauthorized
	}
	return user, nil
}

func ContextWithUser(ctx context.Context, user *bunconf.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func ProjectFromContext(ctx context.Context) *bunconf.Project {
	project, _ := ctx.Value(projectCtxKey{}).(*bunconf.Project)
	return project
}

func ContextWithProject(ctx context.Context, project *bunconf.Project) context.Context {
	return context.WithValue(ctx, projectCtxKey{}, project)
}

type Middleware struct {
	app *bunapp.App
}

func NewMiddleware(app *bunapp.App) *Middleware {
	return &Middleware{
		app: app,
	}
}

func (m *Middleware) UserAndProject(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		user := UserFromRequest(m.app, req)
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

func UserFromRequest(app *bunapp.App, req bunrouter.Request) *bunconf.User {
	if !app.Config().AuthRequired() {
		return GuestUser
	}

	if user := tryCookieFromRequest(app, req); user != nil {
		return user
	}

	if user := tryCloudflareFromRequest(app, req); user != nil {
		return user
	}

	return nil
}

func tryCookieFromRequest(app *bunapp.App, req bunrouter.Request) *bunconf.User {
	if len(app.Config().Users) == 0 {
		return nil
	}

	ctx := req.Context()

	cookie, err := req.Cookie(tokenCookieName)

	if err != nil || cookie.Value == "" {
		return nil
	}

	userID, err := decodeUserToken(app, cookie.Value)
	if err != nil {
		app.Zap(ctx).Error("decodeUserToken failed", zap.Error(err))
		return nil
	}

	users := app.Config().Users
	for i := range users {
		user := &users[i]
		if user.ID == userID {
			return user
		}
	}

	return nil
}

func tryCloudflareFromRequest(app *bunapp.App, req bunrouter.Request) *bunconf.User {
	if !app.Config().CloudflareAuthEnabled() {
		return nil
	}

	// Adapted from https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/

	headers := req.Header
	accessJWT := headers.Get("Cf-Access-Jwt-Assertion")
	if accessJWT == "" {
		return nil
	}

	// TODO(aramperes): Initialize this on startup instead
	var (
		ctx        = req.Context()
		teamDomain = app.Config().UserProviders.Cloudflare.TeamURL
		certsURL   = fmt.Sprintf("%s/cdn-cgi/access/certs", teamDomain)
		policyAUD  = app.Config().UserProviders.Cloudflare.Audience

		config = &oidc.Config{
			ClientID: policyAUD,
		}
		keySet   = oidc.NewRemoteKeySet(ctx, certsURL)
		verifier = oidc.NewVerifier(teamDomain, keySet, config)
	)

	token, err := verifier.Verify(ctx, accessJWT)
	if err != nil {
		app.Zap(ctx).Error("verifyCloudflareToken failed", zap.Error(err))
		return nil
	}

	var claims struct {
		Email string `json:"email"`
	}

	if err := token.Claims(&claims); err != nil {
		app.Zap(ctx).Error("parseCloudflareToken failed", zap.Error(err))
		return nil
	}

	return &bunconf.User{
		// Note, all cloudflare users will share this ID for now.
		ID:       app.Config().UserProviders.Cloudflare.ID,
		Username: claims.Email,
	}
}

func ProjectFromRequest(app *bunapp.App, req bunrouter.Request) (*bunconf.Project, error) {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return nil, err
	}

	project, err := SelectProjectByID(ctx, app, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	return project, nil
}

//------------------------------------------------------------------------------

var jwtSigningMethod = jwt.SigningMethodHS256

func encodeUserToken(app *bunapp.App, userID uint64, ttl time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(userID, 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)

	key := []byte(app.Config().SecretKey)
	return token.SignedString(key)
}

func decodeUserToken(app *bunapp.App, jwtToken string) (uint64, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(app.Config().SecretKey), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid JWT token")
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

//------------------------------------------------------------------------------

func findUserByPassword(app *bunapp.App, username string, password string) *bunconf.User {
	users := app.Config().Users
	for i := range users {
		user := &users[i]
		if subtle.ConstantTimeCompare([]byte(user.Username), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) == 1 {
			return user
		}
	}
	return nil
}
