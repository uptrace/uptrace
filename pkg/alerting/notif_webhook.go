package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/segmentio/encoding/json"
	"go.uber.org/zap"

	"github.com/uptrace/uptrace/pkg/org"
)

func (h *NotifChannelHandler) notifyByWebhookHandler(ctx context.Context, eventID, channelID uint64) error {
	alert, err := selectAlertWithEvent(ctx, h.PG, h.Users, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := h.Projects.SelectByID(ctx, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectWebhookNotifChannel(ctx, h.PG, channelID)
	if err != nil {
		return err
	}

	return h.notifyByWebhookChannel(ctx, project, alert, channel)
}

func (h *NotifChannelHandler) notifyByWebhookChannel(
	ctx context.Context,
	project *org.Project,
	alert org.Alert,
	channel *WebhookNotifChannel,
) error {
	if channel.State != NotifChannelDelivering {
		return nil
	}

	var msg any

	switch channel.Type {
	case NotifChannelWebhook:
		msg = NewWebhookMessage(h.Conf, alert, channel.Params.Payload)
	case NotifChannelAlertmanager:
		var err error
		msg, err = NewAlertmanagerMessage(h.Conf, project, alert)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported webhook type: %s", channel.Type)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", channel.Params.URL, &buf)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	req.Header.Set("User-Agent", "Uptrace/1.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<10))
		if err != nil {
			return err
		}

		var out struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &out); err == nil {
			h.Logger.Error("http.Post failed", zap.String("message", out.Message))
		} else {
			if len(body) > 100 {
				body = body[:100]
			}
			h.Logger.Error("http.Post failed", zap.String("message", string(body)))
		}

		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return nil
}
