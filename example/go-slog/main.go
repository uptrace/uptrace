package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()

	uptrace.ConfigureOpentelemetry()
	defer uptrace.Shutdown(ctx)

	// Configure slog with Otel handler.
	slog.SetDefault(otelslog.NewLogger("app_or_package_name"))

	tracer := otel.Tracer("slog-example")

	ctx, span := tracer.Start(ctx, "operation-name")
	defer span.End()

	slog.ErrorContext(ctx, "Hello world!", "locale", "en_US")
	fmt.Println(uptrace.TraceURL(trace.SpanFromContext(ctx)))
}
