package alerting

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/uuid"
)

func createErrorAlertHandler(
	ctx context.Context,
	projectID uint32,
	groupID uint64,
	traceID uuid.UUID,
	spanID uint64,
) error {
	app := bunapp.AppFromContext(ctx)

	span := &tracing.Span{
		ProjectID: projectID,
		TraceID:   traceID,
		ID:        spanID,
	}
	if err := tracing.SelectSpan(ctx, app, span); err != nil {
		return err
	}

	baseAlert := &org.BaseAlert{
		ProjectID: projectID,
		DedupHash: groupID,

		Name:  span.EventName,
		State: org.AlertOpen,

		TrackableModel: org.ModelSpanGroup,
		TrackableID:    groupID,
		Attrs:          alertAttrs(span),

		Type: org.AlertError,

		CreatedAt: span.Time,
	}

	alert, err := selectErrorAlert(ctx, app, baseAlert)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if alert == nil {
		alert = &ErrorAlert{
			BaseAlert: baseAlert,
		}
		alert.BaseAlert.Params.Any = &alert.Params

		alert.Params.TraceID = traceID
		alert.Params.SpanID = spanID
		alert.Params.SpanCount = 1

		return createAlert(ctx, app, alert)
	}

	spanCount, spanCountTime, err := countAlertSpans(ctx, app, alert, span)
	if err != nil {
		return err
	}

	spanCountThreshold := nextSpanCountThreshold(alert.Params.SpanCount)
	triggered := spanCount >= spanCountThreshold || alert.State == org.AlertClosed

	newParams := alert.Params.Clone()
	newParams.SpanCount = spanCount
	newParams.SpanCountTime = spanCountTime
	if triggered {
		newParams.TraceID = traceID
		newParams.SpanID = spanID
	}

	q := app.PG.NewUpdate().
		Model(alert).
		Set("params = ?", newParams).
		Where("id = ?", alert.ID).
		Where("state = ?", alert.State).
		Where("params = ?", alert.Params).
		Returning("state, params, updated_at")

	if triggered {
		q = q.Set("state = ?", org.AlertOpen).
			Set("updated_at = ?", span.Time)
	}

	res, err := q.Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	if triggered {
		if err := createAlertEvent(ctx, app, alert, &org.AlertEvent{
			ProjectID: alert.ProjectID,
			AlertID:   alert.ID,
			Name:      org.AlertEventRecurring,
			Params:    bunutil.Params{Any: alert.Params},
		}); err != nil {
			return err
		}
	}

	return nil
}

func alertAttrs(span *tracing.Span) map[string]string {
	attrs := make(map[string]string)

	if str, _ := span.Attrs[attrkey.ServiceName].(string); str != "" {
		attrs[attrkey.ServiceName] = str
	}
	if str, _ := span.Attrs[attrkey.ServiceVersion].(string); str != "" {
		attrs[attrkey.ServiceVersion] = str
	}
	if str, _ := span.Attrs[attrkey.DeploymentEnvironment].(string); str != "" {
		attrs[attrkey.DeploymentEnvironment] = str
	}

	switch span.Type {
	case tracing.SpanTypeHTTP:
		addSpanAttrs(attrs, span, attrkey.HTTPMethod)
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
) (uint64, time.Time, error) {
	timeGTE := alert.Params.SpanCountTime
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
		return 0, time.Time{}, err
	}

	if !alert.Params.SpanCountTime.IsZero() {
		spanCount += alert.Params.SpanCount
	}

	return spanCount, timeLT, nil
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
