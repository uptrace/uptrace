package bunotel

import (
	"context"
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	Meter  = global.Meter("github.com/uptrace/uptrace")
	Tracer = otel.Tracer("github.com/uptrace/uptrace")
)

var projectIDAttr = attribute.Key("project_id")

func ProjectIDAttr(projectID uint32) attribute.KeyValue {
	return projectIDAttr.Int64(int64(projectID))
}

func RunWithNewRoot(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := Tracer.Start(ctx, name, trace.WithNewRoot())
	defer span.End()

	if err := fn(ctx); err != nil {
		otelzap.Ctx(ctx).Error(fmt.Sprintf("%s failed", name), zap.Error(err))
		return err
	}
	return nil
}
