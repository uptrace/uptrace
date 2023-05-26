package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"path"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const fromName = "Uptrace"

type EmailNotifier struct {
	disabled bool

	mu     sync.Mutex
	client *mail.Client

	emails *template.Template

	from string
}

func NewEmailNotifier(app *bunapp.App) *EmailNotifier {
	conf := app.Config().SMTPMailer

	if !conf.Enabled {
		app.Logger.Info("smtp_mailer is disabled in the config")
		return &EmailNotifier{
			disabled: true,
		}
	}

	client, err := mail.NewClient(
		conf.Host,
		mail.WithPort(conf.Port),
		mail.WithSMTPAuth(conf.AuthType),
		mail.WithUsername(conf.Username),
		mail.WithPassword(conf.Password),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	)
	if err != nil {
		app.Logger.Error("mail.NewClient failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	emails, err := template.New("").
		Funcs(sprig.FuncMap()).
		ParseFS(bunapp.FS(), path.Join("email", "*.html"))
	if err != nil {
		app.Logger.Error("template.New failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	return &EmailNotifier{
		client: client,
		emails: emails,

		from: conf.From,
	}
}

func (n *EmailNotifier) NotifyHandler(
	ctx context.Context, eventID uint64, recipients []string,
) error {
	if n.disabled {
		return nil
	}

	app := bunapp.AppFromContext(ctx)

	alert, err := selectAlertWithEvent(ctx, app, eventID)
	if err != nil {
		return err
	}
	baseAlert := alert.Base()

	project, err := org.SelectProject(ctx, app, baseAlert.ProjectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		return n.notifyOnErrorAlert(ctx, app, project, alert, recipients)
	case *MetricAlert:
		return n.notifyOnMetricAlert(ctx, app, project, alert, recipients)
	default:
		return fmt.Errorf("unknown alert type: %T", alert)
	}

	return nil
}

func (n *EmailNotifier) notifyOnErrorAlert(
	ctx context.Context,
	app *bunapp.App,
	project *org.Project,
	alert *ErrorAlert,
	recipients []string,
) error {
	const tplName = "error_alert.html"

	span := &tracing.Span{
		ProjectID: alert.ProjectID,
		TraceID:   alert.Params.TraceID,
		ID:        alert.Params.SpanID,
	}
	if err := tracing.SelectSpan(ctx, app, span); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	var buf bytes.Buffer

	keys := attrKeys(span.Attrs)
	if err := n.emails.ExecuteTemplate(&buf, tplName, map[string]any{
		"siteURL": app.SiteURL(""),
		"project": project,
		"alert":   alert,
		"attrs":   span.Attrs,
		"keys":    keys,
	}); err != nil {
		return err
	}

	msg := mail.NewMsg()
	msg.Subject(fmt.Sprintf("[%s] %s", project.Name, alert.Name))
	msg.SetBodyString(mail.TypeTextHTML, buf.String())

	if err := msg.FromFormat(fromName, n.from); err != nil {
		return err
	}
	if err := msg.To(recipients...); err != nil {
		return err
	}

	return n.send(msg)
}

func attrKeys(attrs map[string]any) []string {
	keys := make([]string, 0, len(attrs))

	for key := range attrs {
		if strings.HasPrefix(key, "_") {
			continue
		}
		keys = append(keys, key)
	}

	slices.Sort(keys)
	return keys
}

func (n *EmailNotifier) errorNotificationTpl(
	ctx context.Context,
	app *bunapp.App,
	project *org.Project,
	alert *ErrorAlert,
	span *tracing.Span,
) (string, error) {
	return "", nil
}

func (n *EmailNotifier) notifyOnMetricAlert(
	ctx context.Context,
	app *bunapp.App,
	project *org.Project,
	alert *MetricAlert,
	recipients []string,
) error {
	const tplName = "metric_alert.html"

	var buf bytes.Buffer

	if err := n.emails.ExecuteTemplate(&buf, tplName, map[string]any{
		"siteURL": app.SiteURL(""),
		"project": project,
		"alert":   alert,
	}); err != nil {
		return err
	}

	msg := mail.NewMsg()
	msg.Subject(fmt.Sprintf("[%s] %s", project.Name, alert.Name))
	msg.SetBodyString(mail.TypeTextHTML, buf.String())

	if err := msg.FromFormat(fromName, n.from); err != nil {
		return err
	}
	if err := msg.To(recipients...); err != nil {
		return err
	}

	return n.send(msg)
}

func (n *EmailNotifier) send(msg *mail.Msg) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.client.DialAndSend(msg)
}
