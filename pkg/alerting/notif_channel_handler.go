package alerting

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/slack-go/slack"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type NotifChannelHandler struct {
	*bunapp.App

	httpClient *http.Client
}

func NewNotifChannelHandler(app *bunapp.App) *NotifChannelHandler {
	return &NotifChannelHandler{
		App: app,

		httpClient: app.HTTPClient,
	}
}

func (h *NotifChannelHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	channels := make([]*BaseNotifChannel, 0)

	if err := h.PG.NewSelect().
		Model(&channels).
		Where("project_id = ?", project.ID).
		Order("id ASC").
		Limit(100).
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channels": channels,
	})
}

func (h *NotifChannelHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	channel := NotifChannelFromContext(ctx)

	if _, err := h.PG.NewDelete().
		Model(channel).
		Where("id = ?", channel.Base().ID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) Pause(w http.ResponseWriter, req bunrouter.Request) error {
	return h.UpdateNotifChannelState(w, req, NotifChannelPaused)
}

func (h *NotifChannelHandler) Unpause(w http.ResponseWriter, req bunrouter.Request) error {
	return h.UpdateNotifChannelState(w, req, NotifChannelDelivering)
}

func (h *NotifChannelHandler) UpdateNotifChannelState(
	w http.ResponseWriter, req bunrouter.Request, state NotifChannelState,
) error {
	ctx := req.Context()
	channel := NotifChannelFromContext(ctx).Base()

	if err := UpdateNotifChannelState(ctx, h.App, channel, state); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

//------------------------------------------------------------------------------

func (h *NotifChannelHandler) SlackShow(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	channel, err := SlackNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

type SlackNotifChannelIn struct {
	Name   string      `json:"name"`
	Params SlackParams `json:"params"`
}

func (in *SlackNotifChannelIn) Validate(channel *SlackNotifChannel) error {
	if in.Name == "" {
		return errors.New("channel name can't be empty")
	}
	if in.Params.WebhookURL == "" {
		return errors.New("webhook URL can't be empty")
	}

	u, err := url.Parse(in.Params.WebhookURL)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported URL protocol scheme: %q", u.Scheme)
	}

	channel.Name = in.Name
	channel.Params.WebhookURL = in.Params.WebhookURL

	return nil
}

func (h *NotifChannelHandler) SlackCreate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	channel := &SlackNotifChannel{
		BaseNotifChannel: &BaseNotifChannel{
			ProjectID: project.ID,
			Type:      NotifChannelSlack,
		},
	}

	in := new(SlackNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendSlackTestMsg(channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := InsertNotifChannel(ctx, h.App, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) SlackUpdate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	channel, err := SlackNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	in := new(SlackNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendSlackTestMsg(channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := UpdateNotifChannel(ctx, h.App, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) sendSlackTestMsg(channel *SlackNotifChannel) error {
	webhookURL := channel.Params.WebhookURL
	if webhookURL == "" {
		return errors.New("webhook URL can't be empty")
	}

	msg := &slack.WebhookMessage{
		Text: fmt.Sprintf("Test message from Uptrace"),
	}
	if err := slack.PostWebhook(webhookURL, msg); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

func (h *NotifChannelHandler) WebhookShow(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	channel, err := WebhookNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

type WebhookNotifChannelIn struct {
	Type   NotifChannelType `json:"type"`
	Name   string           `json:"name"`
	Params struct {
		WebhookParams
		Payload string `json:"payload"`
	} `json:"params"`
}

func (in *WebhookNotifChannelIn) Validate(channel *WebhookNotifChannel) error {
	switch in.Type {
	case NotifChannelWebhook, NotifChannelAlertmanager:
		channel.Type = in.Type
	default:
		return fmt.Errorf("unsupported notification channel type: %q", in.Type)
	}

	if in.Name == "" {
		return errors.New("channel name can't be empty")
	}
	if in.Params.URL == "" {
		return errors.New("url can't be empty")
	}

	u, err := url.Parse(in.Params.URL)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported URL protocol scheme: %q", u.Scheme)
	}

	var payload any
	if in.Params.Payload != "" {
		if err := json.Unmarshal([]byte(in.Params.Payload), &payload); err != nil {
			return fmt.Errorf("invalid JSON payload: %w", err)
		}
	}

	channel.Name = in.Name
	channel.Params.URL = in.Params.URL
	channel.Params.Payload = payload

	return nil
}

func (h *NotifChannelHandler) WebhookCreate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	channel := &WebhookNotifChannel{
		BaseNotifChannel: &BaseNotifChannel{
			ProjectID: project.ID,
		},
	}

	in := new(WebhookNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendWebhookTestMsg(project, channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := InsertNotifChannel(ctx, h.App, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) WebhookUpdate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	channel, err := WebhookNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	in := new(WebhookNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendWebhookTestMsg(project, channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := UpdateNotifChannel(ctx, h.App, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) sendWebhookTestMsg(
	project *org.Project, channel *WebhookNotifChannel,
) error {
	alert := &MetricAlert{
		BaseAlert: &org.BaseAlert{
			ID:        123,
			ProjectID: project.ID,
			Name:      "Test message",
			Type:      org.AlertMetric,
			State:     org.AlertOpen,
			CreatedAt: time.Now(),

			Event: &org.AlertEvent{
				ID:        uint64(time.Now().UnixNano()),
				Name:      org.AlertEventCreated,
				CreatedAt: time.Now(),
			},
		},
	}

	var msg any

	switch channel.Type {
	case NotifChannelWebhook:
		msg = NewWebhookMessage(h.App, alert, channel.Params.Payload)
	case NotifChannelAlertmanager:
		msg = NewAlertmanagerMessage(h.App, alert)
	default:
		return fmt.Errorf("unsupported webhook type: %q", channel.Type)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	resp, err := h.httpClient.Post(channel.Params.URL, "application/json", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<10))
		if err != nil {
			return err
		}

		var message string

		var out struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &out); err == nil {
			message = out.Message
		} else {
			if len(body) > 100 {
				body = body[:100]
			}
			message = string(body)
		}

		return fmt.Errorf("unexpected response from webhook: %s (%s)", resp.Status, message)
	}

	return nil
}

//------------------------------------------------------------------------------

func (h *NotifChannelHandler) EmailShow(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)
	project := org.ProjectFromContext(ctx)

	data := &org.UserProjectData{
		NotifyOnMetrics:         true,
		NotifyOnNewErrors:       true,
		NotifyOnRecurringErrors: true,
	}

	if err := h.PG.NewSelect().
		Model(data).
		Where("user_id = ?", user.ID).
		Where("project_id = ?", project.ID).
		Scan(ctx); err != nil && err != sql.ErrNoRows {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": data,
	})
}

func (h *NotifChannelHandler) EmailUpdate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)
	project := org.ProjectFromContext(ctx)

	var in struct {
		NotifyOnMetrics         bool `json:"notifyOnMetrics"`
		NotifyOnNewErrors       bool `json:"notifyOnNewErrors"`
		NotifyOnRecurringErrors bool `json:"notifyOnRecurringErrors"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	data := &org.UserProjectData{
		UserID:                  user.ID,
		ProjectID:               project.ID,
		NotifyOnMetrics:         in.NotifyOnMetrics,
		NotifyOnNewErrors:       in.NotifyOnNewErrors,
		NotifyOnRecurringErrors: in.NotifyOnRecurringErrors,
	}

	if _, err := h.PG.NewInsert().
		Model(data).
		On("CONFLICT (user_id, project_id) DO UPDATE").
		Set("notify_on_metrics = EXCLUDED.notify_on_metrics").
		Set("notify_on_new_errors = EXCLUDED.notify_on_new_errors").
		Set("notify_on_recurring_errors = EXCLUDED.notify_on_recurring_errors").
		Exec(ctx); err != nil && err != sql.ErrNoRows {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": data,
	})
}

//------------------------------------------------------------------------------

type WebhookMessage struct {
	ID        uint64             `json:"id,string"`
	EventName org.AlertEventName `json:"eventName"`
	Payload   any                `json:"payload"`
	CreatedAt time.Time          `json:"createdAt"`

	Alert struct {
		ID        uint64         `json:"id,string"`
		URL       string         `json:"url"`
		Name      string         `json:"name"`
		Type      org.AlertType  `json:"type"`
		State     org.AlertState `json:"state"`
		CreatedAt time.Time      `json:"createdAt"`
	} `json:"alert"`
}

func NewWebhookMessage(app *bunapp.App, alert org.Alert, payload any) *WebhookMessage {
	baseAlert := alert.Base()

	msg := new(WebhookMessage)

	msg.ID = baseAlert.Event.ID
	msg.EventName = baseAlert.Event.Name
	msg.Payload = payload
	msg.CreatedAt = baseAlert.Event.CreatedAt

	msg.Alert.ID = baseAlert.ID
	msg.Alert.URL = app.SiteURL(baseAlert.URL())
	msg.Alert.Name = baseAlert.Name
	msg.Alert.Type = baseAlert.Type
	msg.Alert.State = baseAlert.State
	msg.Alert.CreatedAt = baseAlert.CreatedAt

	return msg
}

//------------------------------------------------------------------------------

func NewAlertmanagerMessage(app *bunapp.App, alert org.Alert) models.PostableAlerts {
	baseAlert := alert.Base()

	labels := make(models.LabelSet, len(baseAlert.Attrs)+1)
	for k, v := range baseAlert.Attrs {
		labels[cleanLabelName(k)] = v
	}
	labels["alertname"] = baseAlert.Name
	labels["alerturl"] = app.SiteURL(baseAlert.URL())

	annotations := models.LabelSet{
		"summary": alert.Summary(),
	}

	dest := &models.PostableAlert{
		Alert: models.Alert{
			Labels: labels,
		},
		Annotations: annotations,
		StartsAt:    strfmt.DateTime(baseAlert.CreatedAt),
	}
	if baseAlert.State == org.AlertClosed {
		dest.EndsAt = strfmt.DateTime(baseAlert.Event.CreatedAt)
	}

	return models.PostableAlerts{dest}
}

func cleanLabelName(s string) string {
	if isValidLabelName(s) {
		return s
	}

	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowedLabelNameChar(c) {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return unsafeconv.String(r)
}

func isValidLabelName(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowedLabelNameChar(c) {
			return false
		}
	}
	return true
}

func isAllowedLabelNameChar(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_'
}
