package alerting

import (
	"context"
	"database/sql"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
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
	}

	params := &alert.Params
	params.TraceID = traceID
	params.SpanID = spanID

	if alert.ID == 0 {
		return createAlert(ctx, app, alert)
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
		addSpanAttrs(attrs, span, attrkey.LogSeverity)
	case tracing.EventTypeExceptions:
		addSpanAttrs(attrs, span, attrkey.ExceptionType)
	}

	return attrs
}

func addSpanAttrs(attrs map[string]string, span *tracing.Span, keys ...string) {
	for _, key := range keys {
		switch value := span.Attrs[key].(type) {
		case string:
			attrs[key] = value
		case []string:
			if len(value) > 0 {
				attrs[key] = value[0]
			}
		}
	}
}
