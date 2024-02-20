package alerting

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
)

func createErrorAlertHandler(
	ctx context.Context,
	projectID uint32,
	groupID uint64,
	traceID idgen.TraceID,
	spanID idgen.SpanID,
) error {
	app := bunapp.AppFromContext(ctx)

	project, err := org.SelectProject(ctx, app, projectID)
	if err != nil {
		return err
	}

	span := &tracing.Span{
		ProjectID: projectID,
		TraceID:   traceID,
		ID:        spanID,
	}
	if err := tracing.SelectSpan(ctx, app, span); err != nil {
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

	alert, err := selectErrorAlert(ctx, app, baseAlert)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if alert == nil {
		baseAlert.Event = new(org.AlertEvent)
		alert = NewErrorAlertBase(baseAlert)

		alert.Event.Status = org.AlertStatusOpen
		alert.Event.Time = span.Time

		alert.Params.TraceID = traceID
		alert.Params.SpanID = spanID
		alert.Params.SpanCount = 1

		return createAlert(ctx, app, alert)
	}

	spanCount, err := countAlertSpans(ctx, app, alert, span)
	if err != nil {
		return err
	}

	spanCountThreshold := nextSpanCountThreshold(alert.Params.SpanCount)
	triggered := spanCount >= spanCountThreshold || alert.Event.Status == org.AlertStatusClosed

	alert.Params.TraceID = traceID
	alert.Params.SpanID = spanID
	alert.Params.SpanCount = spanCount

	if !triggered {
		// Update the alert so it is not deleted.
		if _, err := app.PG.NewUpdate().
			Model(alert.Event).
			Set("params = ?", alert.Params).
			Set("time = ?", span.Time).
			Set("created_at = ?", span.Time).
			Where("id = ?", alert.Event.ID).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}

	return createAlertEvent(ctx, app, alert, func(tx bun.Tx) error {
		event := alert.Event.Clone()
		event.Name = org.AlertEventRecurring
		event.Status = org.AlertStatusOpen
		event.Time = span.Time
		event.CreatedAt = span.Time

		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}

		if err := updateAlertEvent(ctx, tx, alert.Base(), event); err != nil {
			return err
		}

		return nil
	})
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
	case tracing.SpanTypeHTTP:
		addSpanAttrs(attrs, span, attrkey.HTTPRequestMethod, attrkey.HTTPRoute)
	case tracing.SpanTypeDB:
		addSpanAttrs(attrs, span,
			attrkey.DBSystem,
			attrkey.DBName,
			attrkey.DBOperation,
			attrkey.DBSqlTable)
	case tracing.SpanTypeRPC:
		addSpanAttrs(attrs, span, attrkey.RPCSystem, attrkey.RPCService, attrkey.RPCMethod)
	case tracing.SpanTypeMessaging:
		addSpanAttrs(attrs, span, attrkey.MessagingSystem)
	case tracing.SpanTypeFuncs:
		if project.GroupFuncsByService {
			if str, _ := span.Attrs[attrkey.ServiceName].(string); str != "" {
				attrs[attrkey.ServiceName] = str
			}
		}
	case tracing.EventTypeLog:
		addSpanAttrs(attrs, span, attrkey.LogSeverity, attrkey.ExceptionType)
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
	ctx context.Context, app *bunapp.App, alert *ErrorAlert, span *tracing.Span,
) (uint64, error) {
	timeGTE := alert.CreatedAt
	timeLT := time.Now().Add(-time.Minute).Truncate(time.Minute)

	var spanCount uint64

	if err := tracing.NewSpanIndexQuery(app).
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
	ctx context.Context, app *bunapp.App, alert *org.BaseAlert,
) (*ErrorAlert, error) {
	dest := NewErrorAlert()
	if err := selectMatchingAlert(ctx, app, alert, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func nextSpanCountThreshold(n uint64) uint64 {
	if n < 1e6 {
		next := uint64(100)
		for {
			if next > n {
				return next
			}
			next *= 10
		}
	}

	next := uint64(2e6)
	for {
		if next > n {
			return next
		}
		next <<= 1
	}
}
