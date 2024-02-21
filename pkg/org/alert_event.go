package org

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunutil"
)

type AlertStatus string

const (
	AlertStatusOpen   AlertStatus = "open"
	AlertStatusClosed AlertStatus = "closed"
)

type AlertEventName string

const (
	AlertEventCreated       AlertEventName = "created"
	AlertEventStatusChanged AlertEventName = "status-changed"
	AlertEventRecurring     AlertEventName = "recurring"
)

type AlertEvent interface {
	Validate() error
	Base() *BaseAlertEvent
	Clone() AlertEvent
}

type BaseAlertEvent struct {
	bun.BaseModel `bun:"alert_events,alias:e"`

	ID uint64 `bun:",pk,autoincrement"`

	UserID uint64 `bun:",nullzero"`
	User   *User  `bun:"rel:belongs-to,join:user_id=id"`

	ProjectID uint32
	AlertID   uint64
	Alert     *BaseAlert `bun:"rel:belongs-to,join:alert_id=id"`

	Name   AlertEventName
	Status AlertStatus
	Params bunutil.Params `bun:"type:jsonb,nullzero"` // immutable

	Time      time.Time `bun:",nullzero"`
	CreatedAt time.Time `bun:",nullzero"`
}

func (e *BaseAlertEvent) Validate() error {
	if e.ProjectID == 0 {
		return errors.New("alert event project id can't be zero")
	}
	if e.AlertID == 0 {
		return errors.New("event alert id can't be zero")
	}
	if e.Name == "" {
		return errors.New("alert event name can't be empty")
	}
	if e.Status == "" {
		debug.PrintStack()
		return errors.New("alert event status can't be empty")
	}
	if e.Time.IsZero() {
		return errors.New("alert event time can't be zero")
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = e.Time
	}
	return nil
}

func InsertAlertEvent(ctx context.Context, db bun.IDB, event AlertEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	if _, err := db.NewInsert().
		Model(event).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
