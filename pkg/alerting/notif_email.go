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
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
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

func NewEmailNotifier(logger *otelzap.Logger, conf *bunconf.Config) *EmailNotifier {
	if !conf.SMTPMailer.Enabled {
		logger.Info("smtp_mailer is disabled in the config")
		return &EmailNotifier{
			disabled: true,
		}
	}

	client, err := bunapp.NewMailer(conf)
	if err != nil {
		logger.Error("mail.NewClient failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	emails, err := template.New("").
		Funcs(sprig.FuncMap()).
		ParseFS(bunapp.FS(), path.Join("email", "*.html"))
	if err != nil {
		logger.Error("template.New failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	return &EmailNotifier{
		client: client,
		emails: emails,
		from:   conf.SMTPMailer.From,
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
		if err == sql.ErrNoRows {
			return nil
		}
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
}

func (n *EmailNotifier) notifyOnErrorAlert(
	ctx context.Context,
	app *bunapp.App,
	project *org.Project,
	alert *ErrorAlert,
	recipients []string,
) error {
	const tplName = "error_alert.html"

	span, err := tracing.SelectSpan(
		ctx, app, alert.ProjectID, alert.Event.Params.TraceID, alert.Event.Params.SpanID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	var buf bytes.Buffer

	if err := n.emails.ExecuteTemplate(&buf, tplName, map[string]any{
		"projectName":        project.Name,
		"projectSettingsURL": app.SiteURL(project.EmailSettingsURL()),

		"title":     emailErrorFormatter.Format(project, alert),
		"alert":     alert,
		"alertName": alert.Name,
		"alertURL":  app.SiteURL(alert.URL()),

		"spanAttrs": span.Attrs,
		"attrKeys":  attrKeys(span.Attrs),
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
		"projectName":        project.Name,
		"projectSettingsURL": app.SiteURL(project.EmailSettingsURL()),

		"title":       template.HTML(emailMetricFormatter.Format(project, alert)),
		"longSummary": template.HTML(alert.LongSummary("<br />")),
		"alert":       alert,
		"alertName":   alert.Name,
		"alertURL":    app.SiteURL(alert.URL()),
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
