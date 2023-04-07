package alerting

import (
	"bytes"
	"context"
	"fmt"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type NotifChannel interface {
	Base() *BaseNotifChannel
}

type BaseNotifChannel struct {
	bun.BaseModel `bun:"notif_channels,alias:c"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name  string            `json:"name" bun:",nullzero"`
	Type  NotifChannelType  `json:"type"`
	State NotifChannelState `json:"state" bun:",nullzero"`

	ParamsData []byte         `json:"-" bun:"params,nullzero,type:jsonb"`
	ParamsMap  map[string]any `json:"params" bun:"-"`
}

type NotifChannelType string

const (
	NotifChannelSlack        NotifChannelType = "slack"
	NotifChannelWebhook      NotifChannelType = "webhook"
	NotifChannelAlertmanager NotifChannelType = "alertmanager"
)

type NotifChannelState string

const (
	NotifChannelDraft      NotifChannelState = "draft"
	NotifChannelDelivering NotifChannelState = "delivering"
	NotifChannelPaused     NotifChannelState = "paused"
	NotifChannelDisabled   NotifChannelState = "disabled"
)

var _ NotifChannel = (*BaseNotifChannel)(nil)

func (c *BaseNotifChannel) Base() *BaseNotifChannel {
	return c
}

var _ bun.AfterScanRowHook = (*BaseNotifChannel)(nil)

func (c *BaseNotifChannel) AfterScanRow(ctx context.Context) error {
	if len(c.ParamsData) == 0 {
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(c.ParamsData))
	dec.UseNumber()
	if err := dec.Decode(&c.ParamsMap); err != nil {
		return fmt.Errorf("AfterScanRow failed: %w", err)
	}
	return nil
}

func InsertNotifChannel(ctx context.Context, app *bunapp.App, channel NotifChannel) error {
	_, err := app.PG.NewInsert().
		Model(channel).
		Exec(ctx)
	return err
}

func UpdateNotifChannel(ctx context.Context, app *bunapp.App, channel NotifChannel) error {
	base := channel.Base()
	_, err := app.PG.NewUpdate().
		Model(base).
		Set("name = ?", base.Name).
		Set("params = ?", string(base.ParamsData)).
		Where("id = ?", base.ID).
		Exec(ctx)
	return err
}

func UpdateNotifChannelState(
	ctx context.Context, app *bunapp.App, channel *BaseNotifChannel, state NotifChannelState,
) error {
	if _, err := app.PG.NewUpdate().
		Model(channel).
		Set("state = ?", state).
		Where("id = ?", channel.ID).
		Where("state = ?", channel.State).
		Returning("state").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func SelectNotifChannel(
	ctx context.Context, app *bunapp.App, channelID uint64,
) (NotifChannel, error) {
	channel := new(BaseNotifChannel)

	if err := app.PG.NewSelect().
		Model(channel).
		Where("id = ?", channelID).
		Scan(ctx); err != nil {
		return nil, err
	}

	switch channel.Type {
	case NotifChannelSlack:
		return newSlackNotifChannel(channel)
	case NotifChannelWebhook, NotifChannelAlertmanager:
		return newWebhookNotifChannel(channel)
	default:
		return nil, fmt.Errorf("unsupported notification channel: %q", channel.Type)
	}
}
