package org

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
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
	}
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

func SelectUser(ctx context.Context, app *bunapp.App, id uint64) (*User, error) {
	user := new(User)
	if err := app.PG.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func SelectUserByEmail(ctx context.Context, app *bunapp.App, email string) (*User, error) {
	user := new(User)
	if err := app.PG.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func SelectUserByToken(ctx context.Context, app *bunapp.App, token string) (*User, error) {
	user := new(User)
	if err := app.PG.NewSelect().
		Model(user).
		Where("auth_token = ?", token).
		Scan(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func GetOrCreateUser(ctx context.Context, app *bunapp.App, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if _, err := app.PG.NewInsert().
		Model(user).
		On("CONFLICT (email) DO UPDATE").
		Set("name = coalesce(EXCLUDED.name, u.name)").
		Set("avatar = EXCLUDED.avatar").
		Set("updated_at = now()").
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
