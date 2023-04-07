package alerting

import (
	"bytes"
	"context"
	"fmt"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type WebhookNotifChannel struct {
	*BaseNotifChannel `bun:",inherit"`

	Params WebhookParams `json:"params"`
}

type WebhookParams struct {
	URL     string `json:"url"`
	Payload any    `json:"payload"`
}

func newWebhookNotifChannel(src *BaseNotifChannel) (*WebhookNotifChannel, error) {
	channel := &WebhookNotifChannel{
		BaseNotifChannel: src,
	}

	dec := json.NewDecoder(bytes.NewReader(src.ParamsData))
	dec.UseNumber()
	if err := dec.Decode(&channel.Params); err != nil {
		return nil, err
	}

	return channel, nil
}

var _ NotifChannel = (*WebhookNotifChannel)(nil)

func (c *WebhookNotifChannel) Base() *BaseNotifChannel {
	data, err := json.Marshal(c.Params)
	if err != nil {
		panic(err)
	}
	c.BaseNotifChannel.ParamsData = data
	return c.BaseNotifChannel
}

func SelectWebhookNotifChannel(
	ctx context.Context, app *bunapp.App, channelID uint64,
) (*WebhookNotifChannel, error) {
	channelAny, err := SelectNotifChannel(ctx, app, channelID)
	if err != nil {
		return nil, err
	}

	channel, ok := channelAny.(*WebhookNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notif channel: %T", channelAny)
	}
	return channel, nil
}
