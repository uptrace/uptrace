package alerting

import (
	"context"
	"database/sql"
	"time"

	"github.com/vmihailenco/taskq/v4"
	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
)

type AlertNotifierParams struct {
	fx.In

	Logger    *otelzap.Logger
	PG        *bun.DB
	CH        *ch.DB
	Projects  *org.ProjectGateway
	MainQueue taskq.Queue
}

type AlertNotifier struct {
	*AlertNotifierParams
}

func NewAlertNotifier(p AlertNotifierParams) *AlertNotifier {
	return &AlertNotifier{&p}
}

func (n *AlertNotifier) ErrorHandler(
	ctx context.Context,
	projectID uint32,
	groupID uint64,
	traceID idgen.TraceID,
	spanID idgen.SpanID,
) error {
	project, err := n.Projects.SelectByID(ctx, projectID)
	if err != nil {
		return err
	}

	span, err := tracing.SelectSpan(ctx, n.CH, projectID, traceID, spanID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	baseAlert := &org.BaseAlert{
		ProjectID: projectID,

		Name:  span.DisplayName,
		Attrs: alertAttrs(project, span),

		Type:        org.AlertError,
		SpanGroupID: groupID,
	}

	alert, err := selectErrorAlert(ctx, n.PG, baseAlert)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if alert == nil {
		alert = &ErrorAlert{
			BaseAlert: *baseAlert,
			Event:     new(ErrorAlertEvent),
		}

		alert.Event.Status = org.AlertStatusOpen
		alert.Event.Time = span.Time

		alert.Event.Params.TraceID = traceID
		alert.Event.Params.SpanID = spanID
		alert.Event.Params.SpanCount = 1

		return createAlert(ctx, n.Logger, n.PG, n.CH, n.MainQueue, alert)
	}

	spanCount, err := countAlertSpans(ctx, n.CH, alert, span)
	if err != nil {
		return err
	}

	spanCountThreshold := nextSpanCountThreshold(alert.Event.Params.SpanCount)

	alert.Name = span.DisplayName
	alert.Event.Time = span.Time
	alert.Event.Params.TraceID = traceID
	alert.Event.Params.SpanID = spanID
	alert.Event.Params.SpanCount = spanCount

	if !shouldNotifyOnError(alert, spanCountThreshold) {
		if _, err := n.PG.NewUpdate().
			Model(alert.Event).
			Set("params = ?", alert.Event.Params).
			Set("time = ?", alert.Event.Time).
			Where("id = ?", alert.Event.ID).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}

	return tryAlertInTx(ctx, n.Logger, n.PG, n.CH, n.MainQueue, alert, func(tx bun.Tx) error {
		event := alert.Event.Clone()
		baseEvent := event.Base()
		baseEvent.Name = org.AlertEventRecurring
		baseEvent.Status = org.AlertStatusOpen
		baseEvent.Time = span.Time
		baseEvent.CreatedAt = span.Time

		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}
		if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
			return err
		}
		return nil
	})
}

func shouldNotifyOnError(alert *ErrorAlert, spanCountThreshold int64) bool {
	if alert.Event.Status == org.AlertStatusClosed {
		return true
	}

	elapsed := alert.Event.Time.Sub(alert.Event.CreatedAt)

	if elapsed >= 10*time.Minute && alert.Event.Params.SpanCount >= spanCountThreshold {
		return true
	}

	var elapsedThreshold time.Duration
	if time.Since(alert.CreatedAt) >= 72*time.Hour {
		elapsedThreshold = 7 * 24 * time.Hour // 1 week
	} else {
		elapsedThreshold = 24 * time.Hour
	}
	if elapsed >= elapsedThreshold {
		return true
	}

	return false
}

func alertAttrs(project *org.Project, span *tracing.Span) map[string]string {
	attrs := make(map[string]string)

	if project.GroupByEnv {
		if str, _ := span.Attrs[attrkey.DeploymentEnvironment].(string); str != "" {
			attrs[attrkey.DeploymentEnvironment] = str
		}
	}

	if !span.IsEvent() {
		attrs[attrkey.SpanKind] = span.Kind
	}

	switch span.Type {
	case tracing.TypeSpanHTTPClient, tracing.TypeSpanHTTPServer:
		addSpanAttrs(attrs, span, attrkey.HTTPRequestMethod, attrkey.HTTPRoute)
	case tracing.TypeSpanDB:
		addSpanAttrs(attrs, span,
			attrkey.DBSystem,
			attrkey.DBName,
			attrkey.DBOperation,
			attrkey.DBSqlTable)
	case tracing.TypeSpanRPC:
		addSpanAttrs(attrs, span, attrkey.RPCSystem, attrkey.RPCService, attrkey.RPCMethod)
	case tracing.TypeSpanMessaging:
		addSpanAttrs(attrs, span, attrkey.MessagingSystem)
	case tracing.TypeSpanFuncs:
		if project.GroupFuncsByService {
			if str, _ := span.Attrs[attrkey.ServiceName].(string); str != "" {
				attrs[attrkey.ServiceName] = str
			}
		}
	case tracing.TypeLog:
		addSpanAttrs(attrs, span,
			attrkey.LogSeverity,
			attrkey.ExceptionType,
			attrkey.TelemetrySDKLanguage)
	}

	return attrs
}

func addSpanAttrs(attrs map[string]string, span *tracing.Span, keys ...string) {
	for _, key := range keys {
		switch value := span.Attrs[key].(type) {
		case string:
			if value != "" {
				attrs[key] = value
			}
		case []string:
			if len(value) > 0 {
				str := value[0]
				if str != "" {
					attrs[key] = str
				}
			}
		}
	}
}

func countAlertSpans(
	ctx context.Context, ch *ch.DB, alert *ErrorAlert, span *tracing.Span,
) (int64, error) {
	timeGTE := alert.CreatedAt
	timeLT := time.Now().Add(-time.Minute).Truncate(time.Minute)

	var spanCount int64

	if err := tracing.NewSpanIndexQuery(ch).
		ColumnExpr("toUInt64(sum(s.count))").
		Where("s.project_id = ?", span.ProjectID).
		Where("s.type = ?", span.Type).
		Where("s.system = ?", span.System).
		Where("s.group_id = ?", span.GroupID).
		Where("s.time >= ?", timeGTE).
		Where("s.time < ?", timeLT).
		Scan(ctx, &spanCount); err != nil {
		return 0, err
	}

	return spanCount, nil
}

func selectErrorAlert(
	ctx context.Context, pg *bun.DB, alert *org.BaseAlert,
) (*ErrorAlert, error) {
	dest := NewErrorAlert()
	if err := selectMatchingAlert(ctx, pg, alert, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func nextSpanCountThreshold(n int64) int64 {
	if n < 1e6 {
		next := int64(100)
		for {
			if next > n {
				return next
			}
			next *= 10
		}
	}

	next := int64(2e6)
	for {
		if next > n {
			return next
		}
		next <<= 1
	}
}
