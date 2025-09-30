package alerting

import (
	"bytes"
	"context"
	"fmt"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type SlackNotifChannel struct {
	*BaseNotifChannel `bun:",inherit"`

	Params SlackParams `json:"params"`
}

type SlackParams struct {
	WebhookURL string `json:"webhookUrl"`
	Token      string `json:"token"`
	Channel    string `json:"channel"`
	AuthMethod string `json:"authMethod"`
}

func newSlackNotifChannel(src *BaseNotifChannel) (*SlackNotifChannel, error) {
	channel := &SlackNotifChannel{
		BaseNotifChannel: src,
	}

	dec := json.NewDecoder(bytes.NewReader(src.ParamsData))
	dec.UseNumber()
	if err := dec.Decode(&channel.Params); err != nil {
		return nil, err
	}

	return channel, nil
}

var _ NotifChannel = (*SlackNotifChannel)(nil)

func (c *SlackNotifChannel) Base() *BaseNotifChannel {
	data, err := json.Marshal(c.Params)
	if err != nil {
		panic(err)
	}
	c.BaseNotifChannel.ParamsData = data
	return c.BaseNotifChannel
}

func SelectSlackNotifChannel(
	ctx context.Context, app *bunapp.App, channelID uint64,
) (*SlackNotifChannel, error) {
	channelAny, err := SelectNotifChannel(ctx, app, channelID)
	if err != nil {
		return nil, err
	}

	channel, ok := channelAny.(*SlackNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notif channel: %T", channelAny)
	}
	return channel, nil
}
