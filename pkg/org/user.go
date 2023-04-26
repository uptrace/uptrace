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
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	bun.BaseModel `bun:"users,alias:u"`

	ID uint64 `json:"id" bun:",pk,autoincrement"`

	Email    string `json:"email" bun:",nullzero"`
	Password string `json:"-" bun:",nullzero"`

	Name   string `json:"name" bun:",nullzero"`
	Avatar string `json:"avatar" bun:",nullzero"`

	NotifyByEmail bool `json:"notifyByEmail"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
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

func GetOrCreateUser(ctx context.Context, app *bunapp.App, user *User) error {
	if user.Email == "" {
		return errors.New("user email can't be empty")
	}
	user.Email = strings.ToLower(user.Email)
	if user.Avatar == "" {
		user.Avatar = user.gravatar()
	}

	if _, err := app.PG.NewInsert().
		Model(user).
		On("CONFLICT (email) DO UPDATE").
		Set("name = coalesce(EXCLUDED.name, u.name)").
		Set("password = coalesce(EXCLUDED.password, u.password)").
		Set("avatar = EXCLUDED.avatar").
		Set("updated_at = now()").
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
