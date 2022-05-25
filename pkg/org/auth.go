package org

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"

	"go.uber.org/zap"
)

const (
	tokenCookieName = "token"
	tokenTTL        = 7 * 24 * time.Hour
)

var (
	ErrUnauthorized   = httperror.Unauthorized("please log in")
	ErrAccessedDenied = httperror.Forbidden("access denied")
)

var GuestUser = &bunapp.User{
	ID:       1,
	Username: "guest",
}

type (
	userCtxKey    struct{}
	projectCtxKey struct{}
)

func UserFromContext(ctx context.Context) (*bunapp.User, error) {
	user, ok := ctx.Value(userCtxKey{}).(*bunapp.User)
	if !ok {
		return nil, ErrUnauthorized
	}
	return user, nil
}

func ContextWithUser(ctx context.Context, user *bunapp.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func ProjectFromContext(ctx context.Context) (*bunapp.Project, error) {
	project, ok := ctx.Value(projectCtxKey{}).(*bunapp.Project)
	if !ok {
		return nil, ErrUnauthorized
	}
	return project, nil
}

func ContextWithProject(ctx context.Context, project *bunapp.Project) context.Context {
	return context.WithValue(ctx, projectCtxKey{}, project)
}

func NewAuthMiddleware(app *bunapp.App) bunrouter.MiddlewareFunc {
	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return func(w http.ResponseWriter, req bunrouter.Request) error {
			ctx := req.Context()

			user := UserFromRequest(app, req)
			if user == nil {
				return ErrUnauthorized
			}
			ctx = ContextWithUser(ctx, user)

			project, err := ProjectFromRequest(app, req)
			if err != nil {
				return err
			}
			ctx = ContextWithProject(ctx, project)

			return next(w, req.WithContext(ctx))
		}
	}
}

func UserFromRequest(app *bunapp.App, req bunrouter.Request) *bunapp.User {
	ctx := req.Context()

	users := app.Config().Users
	if len(users) == 0 {
		return GuestUser
	}

	cookie, err := req.Cookie(tokenCookieName)
	if err != nil {
		return nil
	}
	if cookie.Value == "" {
		return nil
	}

	userID, err := decodeUserToken(app, cookie.Value)
	if err != nil {
		app.Zap(ctx).Error("decodeUserToken failed", zap.Error(err))
		return nil
	}

	for i := range users {
		user := &users[i]
		if user.ID == userID {
			return user
		}
	}

	return nil
}

func ProjectFromRequest(app *bunapp.App, req bunrouter.Request) (*bunapp.Project, error) {
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
	claims := &jwt.StandardClaims{
		Subject:   strconv.FormatUint(userID, 10),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)

	key := []byte(app.Config().SecretKey)
	return token.SignedString(key)
}

func decodeUserToken(app *bunapp.App, jwtToken string) (uint64, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.StandardClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(app.Config().SecretKey), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid JWT token")
	}

	claims := token.Claims.(*jwt.StandardClaims)

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

//------------------------------------------------------------------------------

func findUserByPassword(app *bunapp.App, username string, password string) *bunapp.User {
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
