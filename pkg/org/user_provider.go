package org

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunconf"
)

var errNoUser = errors.New("org: no user")

type UserProvider interface {
	Auth(req bunrouter.Request) (*bunconf.User, error)
}

type JWTProvider struct {
	secretKey string
}

func NewJWTProvider(secretKey string) *JWTProvider {
	return &JWTProvider{
		secretKey: secretKey,
	}
}

var _ UserProvider = (*JWTProvider)(nil)

func (p *JWTProvider) Auth(req bunrouter.Request) (*bunconf.User, error) {
	cookie, err := req.Cookie(tokenCookieName)
	if err != nil || cookie.Value == "" {
		return nil, errNoUser
	}

	username, err := decodeUserToken(p.secretKey, cookie.Value)
	if err != nil {
		return nil, err
	}

	return &bunconf.User{
		Username: username,
	}, nil
}

var jwtSigningMethod = jwt.SigningMethodHS256

func encodeUserToken(secretKey string, username string, ttl time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)

	key := []byte(secretKey)
	return token.SignedString(key)
}

func decodeUserToken(secretKey string, jwtToken string) (string, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid JWT token")
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	return claims.Subject, nil
}

//------------------------------------------------------------------------------

// Adapted from https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/
type CloudflareProvider struct {
	conf *bunconf.CloudflareProvider

	verifierMu   sync.Mutex
	verifier     *oidc.IDTokenVerifier
	lastUpdateAt time.Time
}

var _ UserProvider = (*CloudflareProvider)(nil)

func NewCloudflareProvider(conf *bunconf.CloudflareProvider) *CloudflareProvider {
	return &CloudflareProvider{
		conf: conf,
	}
}

func (p *CloudflareProvider) Auth(req bunrouter.Request) (*bunconf.User, error) {
	ctx := req.Context()
	headers := req.Header

	accessJWT := headers.Get("Cf-Access-Jwt-Assertion")
	if accessJWT == "" {
		return nil, errNoUser
	}

	token, err := p.getVerifier(ctx).Verify(ctx, accessJWT)
	if err != nil {
		return nil, fmt.Errorf("verifyCloudflareToken failed: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
	}

	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parseCloudflareToken failed: %w", err)
	}

	return &bunconf.User{
		// TODO: is there a username?
		Username: claims.Email,
	}, nil
}

func (p *CloudflareProvider) getVerifier(ctx context.Context) *oidc.IDTokenVerifier {
	p.verifierMu.Lock()
	defer p.verifierMu.Unlock()

	if time.Since(p.lastUpdateAt) < 24*time.Hour {
		return p.verifier
	}

	certsURL := fmt.Sprintf("%s/cdn-cgi/access/certs", p.conf.TeamURL)
	config := &oidc.Config{
		ClientID: p.conf.Audience,
	}
	keySet := oidc.NewRemoteKeySet(ctx, certsURL)
	verifier := oidc.NewVerifier(p.conf.TeamURL, keySet, config)

	p.verifier = verifier
	p.lastUpdateAt = time.Now()

	return verifier
}
