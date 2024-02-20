package org

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/pgquery"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type AlertType string

const (
	AlertError  AlertType = "error"
	AlertMetric AlertType = "metric"
)

type Alert interface {
	Base() *BaseAlert
	URL() string
}

type BaseAlert struct {
	bun.BaseModel `bun:"alerts,alias:a"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	EventID uint64      `json:"-" bun:",nullzero"`
	Event   *AlertEvent `json:"-" bun:"rel:belongs-to,join:event_id=id"`

	Name      string            `json:"name"`
	Attrs     map[string]string `json:"attrs"`
	AttrsHash uint64            `json:"-"`

	Type AlertType `json:"type"`

	SpanGroupID uint64 `json:"spanGroupId,string" bun:",nullzero"`
	MonitorID   uint64 `json:"monitorId" bun:",nullzero"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
}

func NewBaseAlert(alertType AlertType) *BaseAlert {
	return &BaseAlert{
		Type:  alertType,
		Event: new(AlertEvent),
	}
}

func (a *BaseAlert) URL() string {
	return fmt.Sprintf("/alerting/%d/alerts/%d", a.ProjectID, a.ID)
}

var _ json.Marshaler = (*BaseAlert)(nil)

func (a *BaseAlert) MarshalJSON() ([]byte, error) {
	type StrippedBaseAlert BaseAlert

	type AlertOut struct {
		*StrippedBaseAlert

		Status    AlertStatus    `json:"status"`
		Params    bunutil.Params `json:"params"`
		Time      time.Time      `json:"time"`
		UpdatedAt time.Time      `json:"updatedAt"`
	}

	out := &AlertOut{
		StrippedBaseAlert: (*StrippedBaseAlert)(a),
	}
	if a.Event != nil {
		out.Status = a.Event.Status
		out.Params = a.Event.Params
		out.Time = a.Event.Time
		out.UpdatedAt = a.Event.CreatedAt
	}

	return json.Marshal(out)
}

type ErrorAlertParams struct {
	TraceID   idgen.TraceID `json:"traceId"`
	SpanID    idgen.SpanID  `json:"spanId"`
	SpanCount uint64        `json:"spanCount"`
}

func (p *ErrorAlertParams) Clone() *ErrorAlertParams {
	clone := *p
	return &clone
}

func InsertAlert(ctx context.Context, app *bunapp.App, db bun.IDB, a *BaseAlert) (bool, error) {
	if a.ProjectID == 0 {
		return false, errors.New("project id can't be zero")
	}
	if a.Name == "" {
		return false, errors.New("name can't be empty")
	}
	if a.Type == "" {
		return false, errors.New("type can't be empty")
	}

	a.Name = utf8util.Trunc(a.Name, 1000)

	b := pgquery.NewTSBuilder()
	b.AddTitle(a.Name)

	for k, v := range a.Attrs {
		b.AddAttr(k, v)
	}

	res, err := db.NewInsert().
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

	if n != 1 {
		return false, nil
	}
	return true, nil
}
