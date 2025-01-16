package org

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	bun.BaseModel `bun:"users,alias:u"`

	ID uint64 `json:"id" bun:",pk,autoincrement"`

	Email    string `json:"email" bun:",nullzero"`
	Password string `json:"-" bun:"-"`

	Name   string `json:"name" bun:",nullzero"`
	Avatar string `json:"avatar" bun:",nullzero"`

	NotifyByEmail bool   `json:"notifyByEmail"`
	AuthToken     string `json:"authToken"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

func NewUserFromConfig(src *bunconf.User) (*User, error) {
	dest := &User{
		Email:         src.Email,
		Name:          src.Name,
		Avatar:        src.Avatar,
		NotifyByEmail: src.NotifyByEmail,
		AuthToken:     src.AuthToken,
		CreatedAt:     time.Now(),
	}
	dest.UpdatedAt = dest.CreatedAt

	if err := dest.SetPassword(src.Password); err != nil {
		return nil, err
	}
	return dest, nil
}

func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("user email can't be empty")
	}
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if u.Avatar == "" {
		u.Avatar = u.gravatar()
	}
	return nil
}

func (u *User) Username() string {
	if u.Name != "" {
		return u.Name
	}

	i := strings.IndexByte(u.Email, '@')
	if i == 0 {
		return u.Email
	}
	return u.Email[:i]
}

func (u *User) SetPassword(pass string) error {
	pass, err := hashPassword(pass)
	if err != nil {
		return err
	}
	u.Password = pass
	return nil
}

func hashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *User) gravatar() string {
	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=identicon", md5s(u.Email))
}

func md5s(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

type UserGatewayParams struct {
	fx.In

	Conf *bunconf.Config
}

type UserGateway struct {
	*UserGatewayParams

	users []*User
}

func NewUserGateway(p UserGatewayParams) (*UserGateway, error) {
	var users []*User
	for id, u := range p.Conf.Auth.Users {
		user, err := NewUserFromConfig(&u)
		if err != nil {
			return nil, err
		}
		user.ID = uint64(id + 1)

		if err := user.Validate(); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return &UserGateway{
		UserGatewayParams: &p,
		users:             users,
	}, nil
}

func findUser(users []*User, f func(*User) bool) (*User, error) {
	idx := slices.IndexFunc(users, f)
	if idx == -1 {
		return nil, sql.ErrNoRows
	}
	return users[idx], nil
}

func (g *UserGateway) SelectByID(ctx context.Context, id uint64) (*User, error) {
	return findUser(g.users, func(u *User) bool {
		return u.ID == id
	})
}

func (g *UserGateway) SelectByEmail(ctx context.Context, email string) (*User, error) {
	return findUser(g.users, func(u *User) bool {
		return u.Email == email
	})
}

func (g *UserGateway) SelectByToken(ctx context.Context, token string) (*User, error) {
	return findUser(g.users, func(u *User) bool {
		return u.AuthToken == token
	})
}
