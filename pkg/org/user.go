package org

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	bun.BaseModel `bun:"users,alias:u"`

	ID       uint64 `json:"id" bun:",pk,autoincrement"`
	Username string `json:"username"`
	Password string `json:"-" bun:",nullzero"`

	Email  string `json:"email" bun:",nullzero"`
	Avatar string `json:"avatar" bun:",nullzero"`

	NotifyByEmail bool `json:"notifyByEmail"`
}

func (u *User) Init() error {
	if u.Username == "" {
		u.Username = u.Email
	}
	if u.Username == "" {
		return errors.New("username can't be empty")
	}
	if u.Avatar == "" {
		u.Avatar = u.gravatar()
	}
	return nil
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
	email := u.Email
	if email == "" {
		email = u.Username
	}
	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=identicon", md5s(email))
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

func SelectUserByUsername(ctx context.Context, app *bunapp.App, username string) (*User, error) {
	user := new(User)
	if err := app.PG.NewSelect().
		Model(user).
		Where("username = ?", username).
		Scan(ctx); err != nil {
		return nil, err
	}
	return user, nil
}
