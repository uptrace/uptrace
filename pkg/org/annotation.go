package org

import (
	"time"

	"github.com/uptrace/bun"
)

type Annotation struct {
	bun.BaseModel `bun:"alias:e"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`
	Hash      uint64 `json:"-" bun:",nullzero"`

	Name        string            `json:"name"`
	Description string            `json:"description"`
	Color       string            `json:"color"`
	Attrs       map[string]string `json:"attrs"`
	CreatedAt   time.Time         `json:"createdAt" bun:",nullzero"`
}
