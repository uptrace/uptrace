package org

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type SSOHandler struct {
	app  *bunapp.App
	oidc *OIDCProvider
}

type Methods struct {
	OIDC *OIDCMethod `json:"oidc"`
}

func NewSSOHandler(app *bunapp.App) *SSOHandler {
	conf := app.Config()
	ctx := app.Context()

	var oidc *OIDCProvider
	if conf.UserProviders.OIDC != nil {
		if provider, err := NewOIDCProvider(ctx, conf.UserProviders.OIDC); provider != nil {
			oidc = provider
		} else if err != nil {
			app.Zap(ctx).Error("Failed to initialize OIDC user provider", zap.Error(err))
		}
	}

	return &SSOHandler{
		app:  app,
		oidc: oidc,
	}
}

func (h *SSOHandler) ListMethods(w http.ResponseWriter, req bunrouter.Request) error {
	app := h.app
	ctx := req.Context()

	methods := &Methods{}

	if h.oidc != nil {
		oidc, err := h.oidc.Start(w, req)
		if err != nil {
			app.Zap(ctx).Error("Failed to start OIDC flow", zap.Error(err))
		} else if oidc != nil {
			methods.OIDC = oidc
		}
	}

	return httputil.JSON(w, methods)
}

//------------------------------------------------------------------------------

type OIDCMethod struct {
	DisplayName string `json:"name"`
	RedirectURL string `json:"url"`
}

type OIDCProvider struct {
	conf *bunconf.OIDCProvider

	provider *oidc.Provider
	oauth    *oauth2.Config
}

const stateCookieName = "oidc-state"

func NewOIDCProvider(ctx context.Context, conf *bunconf.OIDCProvider) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, conf.IssuerURL)

	if err != nil {
		return nil, err
	}

	scopes := []string{oidc.ScopeOpenID}

	if len(conf.Scopes) > 0 {
		scopes = append(scopes, conf.Scopes...)
	} else {
		scopes = append(scopes, "profile")
	}

	oauth := &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &OIDCProvider{
		conf:     conf,
		provider: provider,
		oauth:    oauth,
	}, nil
}

func (p *OIDCProvider) Start(w http.ResponseWriter, req bunrouter.Request) (*OIDCMethod, error) {
	// Generates a 'state' token to prevent CSRF attacks (will be validated when redirected back to app)
	state, err := randState(32)

	if err != nil {
		return nil, err
	}

	stateCookie := &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
	}

	http.SetCookie(w, stateCookie)

	method := &OIDCMethod{
		RedirectURL: p.oauth.AuthCodeURL(state),
		DisplayName: p.conf.DisplayName,
	}

	return method, nil
}

func (p *OIDCProvider) Exchange(w http.ResponseWriter, req bunrouter.Request) (*bunconf.User, error) {
	ctx := req.Context()

	existingState, _ := req.Cookie(stateCookieName)
	if existingState == nil {
		return nil, errors.New("oidc: no state")
	}

	if req.URL.Query().Get("state") != existingState.Value {
		return nil, errors.New("oidc: state did not match")
	}

	// Unset state to prevent repeat calls
	emptyState := &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, emptyState)

	token, err := p.oauth.Exchange(ctx, req.URL.Query().Get("code"))
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to exchange code: %w", err)
	}

	userInfo, err := p.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to get user info: %w", err)
	}

	var claims *map[string]interface{}
	err = userInfo.Claims(&claims)

	if err != nil {
		return nil, fmt.Errorf("oidc: failed to read claims: %w", err)
	}

	claim := "preferred_username"
	if len(p.conf.Claim) > 0 {
		claim = p.conf.Claim
	}

	var username string
	usernameClaim := (*claims)[claim]

	switch usernameClaim := usernameClaim.(type) {
	case string:
		username = usernameClaim
	case nil:
		return nil, fmt.Errorf("oidc: claim is unset: %s", claim)
	default:
		return nil, fmt.Errorf("oidc: claim must be a string: %s", claim)
	}

	if len(username) == 0 {
		return nil, fmt.Errorf("oidc: claim is empty: %s", claim)
	}

	return &bunconf.User{
		Username: username,
	}, nil
}

func randState(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (h *SSOHandler) OIDCCallback(w http.ResponseWriter, req bunrouter.Request) error {
	user, err := h.oidc.Exchange(w, req)
	if err != nil {
		return err
	}

	token, err := encodeUserToken(h.app.Config().SecretKey, user.Username, tokenTTL)
	if err != nil {
		return err
	}

	cookie := bunapp.NewCookie(h.app, req)
	cookie.Name = tokenCookieName
	cookie.Value = token
	cookie.MaxAge = int(tokenTTL.Seconds())
	http.SetCookie(w, cookie)

	http.Redirect(w, req.Request, "/", http.StatusFound)
	return nil
}
