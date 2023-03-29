package org

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/pgquery"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"github.com/uptrace/uptrace/pkg/uuid"
)

type AlertType string

const (
	AlertError  AlertType = "error"
	AlertMetric AlertType = "metric"
)

type AlertState string

const (
	AlertOpen   AlertState = "open"
	AlertClosed AlertState = "closed"
)

type Alert interface {
	Base() *BaseAlert
	URL() string
	Summary() string
}

type BaseAlert struct {
	bun.BaseModel `bun:"alerts,alias:a"`

	ID        uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name      string     `json:"name"`
	State     AlertState `json:"state"`
	DedupHash uint64     `json:"-"`

	MonitorID      uint64         `json:"monitorId" bun:",nullzero"`
	TrackableModel TrackableModel `json:"trackableModel" bun:",nullzero"`
	TrackableID    uint64         `json:"trackableId,string" bun:",nullzero"`

	Attrs     map[string]string `json:"attrs"`
	AttrsHash uint64            `json:"-"`

	Type   AlertType      `json:"type"`
	Params bunutil.Params `json:"params"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Payload
	Event *AlertEvent `json:"-" bun:"-"`
}

func (a *BaseAlert) Base() *BaseAlert {
	return a
}

func (a *BaseAlert) URL() string {
	return fmt.Sprintf("/alerting/%d/alerts/%d", a.ProjectID, a.ID)
}

type ErrorAlertParams struct {
	TraceID   uuid.UUID `json:"traceId"`
	SpanID    uint64    `json:"spanId,string"`
	SpanCount uint64    `json:"spanCount"`
}

func InsertAlert(ctx context.Context, app *bunapp.App, a *BaseAlert) (bool, error) {
	if a.ProjectID == 0 {
		return false, errors.New("project id can't be zero")
	}
	if a.Name == "" {
		return false, errors.New("name can't be empty")
	}
	if a.Type == "" {
		return false, errors.New("type can't be empty")
	}
	if a.CreatedAt.IsZero() {
		return false, errors.New("created at time can't be zero")
	}
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = a.CreatedAt
	}

	a.Name = utf8util.Trunc(a.Name, 1000)

	if a.Attrs == nil {
		a.Attrs = make(map[string]string)
	}
	a.Attrs[attrkey.AlertType] = string(a.Type)

	b := pgquery.NewTSBuilder()
	b.AddTitle(a.Name)

	for k, v := range a.Attrs {
		b.AddAttr(k, v)
	}

	res, err := app.PG.NewInsert().
		Model(a).
		Value("tsv", "setweight(to_tsvector('english', ?), 'A') || "+
			"setweight(to_tsvector('english', ?), 'B') || "+
			"array_to_tsvector(?)",
			b.Title(), b.Body(), pgdialect.Array(b.Attrs())).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	return true, nil
}
