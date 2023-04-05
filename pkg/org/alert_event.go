package org

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
)

type AlertEventName string

const (
	AlertEventCreated      = "created"
	AlertEventStateChanged = "state-changed"
	AlertEventRecurring    = "recurring"
)

type AlertEvent struct {
	bun.BaseModel `bun:"alert_events,alias:e"`

	ID uint64 `bun:",pk,autoincrement"`

	UserID uint64 `bun:",nullzero"`
	User   *User  `bun:"-"`

	ProjectID uint32
	AlertID   uint64
	Alert     *BaseAlert `bun:"rel:belongs-to,join:alert_id=id"`

	Name   AlertEventName
	Params bunutil.Params `bun:"type:jsonb,nullzero"`

	CreatedAt time.Time `bun:",nullzero"`
}

func InsertAlertEvent(ctx context.Context, app *bunapp.App, event *AlertEvent) error {
	if _, err := app.PG.NewInsert().
		Model(event).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
