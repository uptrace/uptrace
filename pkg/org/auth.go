package org

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

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
	ctx := req.Context()

	users := app.Config().Users
	if len(users) == 0 && len(app.Config().UserTemplates) == 0 {
		return GuestUser
	}

	cookie, err := req.Cookie(tokenCookieName)
	if err != nil {
		return nil
	}
	if cookie.Value == "" {
		return nil
	}

	userID, username, err := decodeUserToken(app, cookie.Value)
	if err != nil {
		app.Zap(ctx).Error("decodeUserToken failed", zap.Error(err))
		return nil
	}

	if username != "" {
		return &bunconf.User{
			ID:       userID,
			Username: username,
		}
	}

	for i := range users {
		user := &users[i]
		if user.ID == userID {
			return user
		}
	}

	return nil
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

func decodeUserToken(app *bunapp.App, jwtToken string) (uint64, string, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(app.Config().SecretKey), nil
		})
	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", errors.New("invalid JWT token")
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	for _, template := range app.Config().UserTemplates {
		if claims.VerifyAudience(template.Audience, true) && claims.Subject != "" {
			return template.ID, claims.Subject, nil
		}
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, "", err
	}

	return id, "", nil
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
