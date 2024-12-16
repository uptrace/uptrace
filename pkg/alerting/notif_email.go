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
	"github.com/wneessen/go-mail"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
)

const fromName = "Uptrace"

type EmailNotifierParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	PG     *bun.DB
	CH     *ch.DB
	PS     *org.ProjectGateway
}

type EmailNotifier struct {
	*EmailNotifierParams

	disabled bool

	mu     sync.Mutex
	client *mail.Client

	emails *template.Template

	from string
}

func NewEmailNotifier(p EmailNotifierParams) *EmailNotifier {
	if !p.Conf.SMTPMailer.Enabled {
		p.Logger.Info("smtp_mailer is disabled in the config")
		return &EmailNotifier{
			disabled: true,
		}
	}

	client, err := bunapp.NewMailer(p.Conf)
	if err != nil {
		p.Logger.Error("mail.NewClient failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	emails, err := template.New("").
		Funcs(sprig.FuncMap()).
		ParseFS(bunapp.FS(), path.Join("email", "*.html"))
	if err != nil {
		p.Logger.Error("template.New failed", zap.Error(err))
		return &EmailNotifier{
			disabled: true,
		}
	}

	return &EmailNotifier{
		EmailNotifierParams: &p,
		client:              client,
		emails:              emails,
		from:                p.Conf.SMTPMailer.From,
	}
}

func (n *EmailNotifier) NotifyHandler(ctx context.Context, eventID uint64, recipients []string) error {
	if n.disabled {
		return nil
	}

	alert, err := selectAlertWithEvent(ctx, n.PG, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := n.PS.SelectByID(ctx, baseAlert.ProjectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		return n.notifyOnErrorAlert(ctx, project, alert, recipients)
	case *MetricAlert:
		return n.notifyOnMetricAlert(ctx, project, alert, recipients)
	default:
		return fmt.Errorf("unknown alert type: %T", alert)
	}
}

func (n *EmailNotifier) notifyOnErrorAlert(
	ctx context.Context,
	project *org.Project,
	alert *ErrorAlert,
	recipients []string,
) error {
	const tplName = "error_alert.html"

	span, err := tracing.SelectSpan(
		ctx, n.CH, alert.ProjectID, alert.Event.Params.TraceID, alert.Event.Params.SpanID,
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
		"projectSettingsURL": n.Conf.SiteURL(project.EmailSettingsURL()),

		"title":     emailErrorFormatter.Format(project, alert),
		"alert":     alert,
		"alertName": alert.Name,
		"alertURL":  n.Conf.SiteURL(alert.URL()),

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
	project *org.Project,
	alert *MetricAlert,
	recipients []string,
) error {
	const tplName = "metric_alert.html"

	var buf bytes.Buffer

	if err := n.emails.ExecuteTemplate(&buf, tplName, map[string]any{
		"projectName":        project.Name,
		"projectSettingsURL": n.Conf.SiteURL(project.EmailSettingsURL()),

		"title":       template.HTML(emailMetricFormatter.Format(project, alert)),
		"longSummary": template.HTML(alert.LongSummary("<br />")),
		"alert":       alert,
		"alertName":   alert.Name,
		"alertURL":    n.Conf.SiteURL(alert.URL()),
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
