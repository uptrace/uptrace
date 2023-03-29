package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
)

func notifyByWebhookHandler(ctx context.Context, eventID, channelID uint64) error {
	app := bunapp.AppFromContext(ctx)

	alert, err := selectAlertWithEvent(ctx, app, eventID)
	if err != nil {
		return err
	}
	baseAlert := alert.Base()

	project, err := org.SelectProject(ctx, app, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectWebhookNotifChannel(ctx, app, channelID)
	if err != nil {
		return err
	}

	return notifyByWebhookChannel(ctx, app, project, alert, channel)
}

func notifyByWebhookChannel(
	ctx context.Context,
	app *bunapp.App,
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
		msg = NewWebhookMessage(app, alert, channel.Params.Payload)
	case NotifChannelAlertmanager:
		msg = NewAlertmanagerMessage(app, alert)
	default:
		return fmt.Errorf("unsupported webhook type: %s", channel.Type)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", channel.Params.URL, &buf)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Uptrace/1.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.HTTPClient.Do(req)
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
			app.Zap(ctx).Error("http.Post failed", zap.String("message", out.Message))
		} else {
			if len(body) > 100 {
				body = body[:100]
			}
			app.Zap(ctx).Error("http.Post failed", zap.String("message", string(body)))
		}

		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return nil
}
