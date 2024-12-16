package alerting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-openapi/strfmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/segmentio/encoding/json"
	"github.com/slack-go/slack"
	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type NotifChannelHandlerParams struct {
	fx.In
	Logger     *otelzap.Logger
	Conf       *bunconf.Config
	PG         *bun.DB
	Projects   *org.ProjectGateway
	Users      *org.UserGateway
	HTTPClient *http.Client
}

type NotifChannelHandler struct {
	*NotifChannelHandlerParams
}

func NewNotifChannelHandler(p NotifChannelHandlerParams) *NotifChannelHandler {
	return &NotifChannelHandler{&p}
}

func registerNotifChannelHandler(p bunapp.RouterParams, h *NotifChannelHandler, middleware *Middleware) {
	p.RouterInternalV1.
		Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id/notification-channels", func(g *bunrouter.Group) {
			g.GET("", h.List)

			g.POST("/slack", h.SlackCreate)
			g.POST("/webhook", h.WebhookCreate)
			g.POST("/telegram", h.TelegramCreate)

			g.GET("/email", h.EmailShow)
			g.PUT("/email", h.EmailUpdate)

			g = g.Use(middleware.NotifChannel)

			g.DELETE("/:channel_id", h.Delete)
			g.PUT("/:channel_id/paused", h.Pause)
			g.PUT("/:channel_id/unpaused", h.Unpause)

			g.GET("/slack/:channel_id", h.SlackShow)
			g.PUT("/slack/:channel_id", h.SlackUpdate)

			g.GET("/webhook/:channel_id", h.WebhookShow)
			g.PUT("/webhook/:channel_id", h.WebhookUpdate)

			g.GET("/telegram/:channel_id", h.TelegramShow)
			g.PUT("/telegram/:channel_id", h.TelegramUpdate)
		})
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

	if err := UpdateNotifChannelState(ctx, h.PG, channel, state); err != nil {
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

	if err := InsertNotifChannel(ctx, h.PG, channel); err != nil {
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

	if err := UpdateNotifChannel(ctx, h.PG, channel); err != nil {
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

func (h *NotifChannelHandler) TelegramShow(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	channel, err := TelegramNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

type TelegramNotifChannelIn struct {
	Name   string `json:"name"`
	Params struct {
		ChatID int64 `json:"chatId"`
	} `json:"params"`
}

func (in *TelegramNotifChannelIn) Validate(channel *TelegramNotifChannel) error {
	if in.Name == "" {
		return errors.New("channel name can't be empty")
	}
	if in.Params.ChatID == 0 {
		return errors.New("chat id can't be empty")
	}

	channel.Name = in.Name
	channel.Params.ChatID = in.Params.ChatID

	return nil
}

func (h *NotifChannelHandler) TelegramCreate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	channel := &TelegramNotifChannel{
		BaseNotifChannel: &BaseNotifChannel{
			ProjectID: project.ID,
			Type:      NotifChannelTelegram,
		},
	}

	in := new(TelegramNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendTelegramTestMsg(channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := InsertNotifChannel(ctx, h.PG, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) sendTelegramTestMsg(channel *TelegramNotifChannel) error {
	bot, err := channel.TelegramBot(h.Conf)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(channel.Params.ChatID, "Test message from Uptrace")
	if _, err := bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (h *NotifChannelHandler) TelegramUpdate(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	channel, err := TelegramNotifChannelFromContext(ctx)
	if err != nil {
		return err
	}

	in := new(TelegramNotifChannelIn)
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(channel); err != nil {
		return httperror.Wrap(err)
	}
	if err := h.sendTelegramTestMsg(channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := UpdateNotifChannel(ctx, h.PG, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
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
	if err := h.sendWebhookTestMsg(ctx, project, channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := InsertNotifChannel(ctx, h.PG, channel); err != nil {
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
	if err := h.sendWebhookTestMsg(ctx, project, channel); err != nil {
		return httperror.Wrap(err)
	}

	if err := UpdateNotifChannel(ctx, h.PG, channel); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"channel": channel,
	})
}

func (h *NotifChannelHandler) sendWebhookTestMsg(
	ctx context.Context, project *org.Project, channel *WebhookNotifChannel,
) error {
	alert := &MetricAlert{
		BaseAlert: org.BaseAlert{
			ID:        123,
			ProjectID: project.ID,
			Name:      "Test message",
			Type:      org.AlertMetric,
			CreatedAt: time.Now(),
		},
		Event: &MetricAlertEvent{
			BaseAlertEvent: org.BaseAlertEvent{
				ID:        uint64(time.Now().UnixNano()),
				Name:      org.AlertEventCreated,
				Status:    org.AlertStatusOpen,
				CreatedAt: time.Now(),
			},
		},
	}

	return h.notifyByWebhookChannel(ctx, project, alert, channel)
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
		ID        uint64          `json:"id,string"`
		URL       string          `json:"url"`
		Name      string          `json:"name"`
		Type      org.AlertType   `json:"type"`
		Status    org.AlertStatus `json:"status"`
		CreatedAt time.Time       `json:"createdAt"`
	} `json:"alert"`
}

func NewWebhookMessage(conf *bunconf.Config, alert org.Alert, payload any) *WebhookMessage {
	baseAlert := alert.Base()
	baseEvent := alert.GetEvent().Base()

	msg := new(WebhookMessage)

	msg.ID = baseEvent.ID
	msg.EventName = baseEvent.Name
	msg.Payload = payload
	msg.CreatedAt = baseEvent.CreatedAt

	msg.Alert.ID = baseAlert.ID
	msg.Alert.URL = conf.SiteURL(baseAlert.URL())
	msg.Alert.Name = baseAlert.Name
	msg.Alert.Type = baseAlert.Type
	msg.Alert.Status = baseEvent.Status
	msg.Alert.CreatedAt = baseAlert.CreatedAt

	return msg
}

//------------------------------------------------------------------------------

func NewAlertmanagerMessage(
	conf *bunconf.Config, project *org.Project, alert org.Alert,
) (models.PostableAlerts, error) {
	baseAlert := alert.Base()

	labels := make(models.LabelSet, len(baseAlert.Attrs)+1)
	for k, v := range baseAlert.Attrs {
		labels[cleanLabelName(k)] = v
	}
	labels["alertname"] = baseAlert.Name
	labels["alerturl"] = conf.SiteURL(baseAlert.URL())

	var summary string
	switch alert := alert.(type) {
	case *ErrorAlert:
		summary = telegramErrorFormatter.Format(project, alert)
	case *MetricAlert:
		summary = telegramMetricFormatter.Format(project, alert)
	default:
		return nil, fmt.Errorf("unknown alert type: %T", alert)
	}

	dest := &models.PostableAlert{
		Alert: models.Alert{
			Labels: labels,
		},
		Annotations: models.LabelSet{
			"summary": summary,
		},
		StartsAt: strfmt.DateTime(baseAlert.CreatedAt),
	}
	if baseAlert.Event.Status == org.AlertStatusClosed {
		dest.EndsAt = strfmt.DateTime(baseAlert.Event.CreatedAt)
	} else {
		endsAt := baseAlert.CreatedAt.Add(30 * 24 * time.Hour)
		dest.EndsAt = strfmt.DateTime(endsAt)
	}

	return models.PostableAlerts{dest}, nil
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
