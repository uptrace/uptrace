package tracing

import (
	"time"

	"github.com/uptrace/bun"
)

type SavedView struct {
	bun.BaseModel `bun:"saved_views,alias:v"`

	ID uint64 `json:"id" bun:",pk,autoincrement"`

	UserHash    uint64 `json:"-"`
	ProjectID uint32 `json:"projectId"`

	Name   string         `json:"name"`
	Route  string         `json:"route"`
	Params map[string]any `json:"params"`
	Query  map[string]any `json:"query"`
	Pinned bool           `json:"pinned"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero,notnull,default:now()"`
}
