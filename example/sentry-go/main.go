package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	dsn := os.Getenv("UPTRACE_DSN")
	if dsn == "" {
		dsn = "http://project1_secret@localhost:14318/1"
	}
	fmt.Println("using DSN:", dsn)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            true,
		AttachStacktrace: true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		panic(err)
	}
	defer sentry.Flush(3 * time.Second)

	ctx := context.Background()
	doWork(ctx)
}

func doWork(ctx context.Context) {
	span := sentry.StartSpan(ctx, "doWork")
	span.SetData("code.function", "main.doWork")
	defer span.Finish()

	ctx = span.Context()

	{
		span := sentry.StartSpan(ctx, "SELECT")
		span.SetData("db.system", "postgresql")
		span.SetData("db.statement", "SELECT * FROM articles LIMIT 100")

		{
			ctx := span.Context()
			span := sentry.StartSpan(ctx, "GET /foo/bar")
			span.SetData("http.method", "GET")
			span.SetData("http.route", "/foo/bar")
			span.SetData("http.url", "https://mydomain.com/foo/bar?q=123")
			span.Finish()
		}

		span.Finish()
	}

	span = sentry.StartSpan(ctx, "AuthService.Auth")
	span.SetData("rpc.system", "grpc")
	span.SetData("rpc.service", "AuthService.Auth")
	span.SetData("rpc.method", "Auth")
	eventID := captureExceptionOnSpan(span.Context(), span, errors.New("Yeah, it works!"))
	if eventID != nil {
		fmt.Println("exception id:", *eventID)
	}
	span.Finish()

	fmt.Println("trace id:", span.TraceID)
}

func captureExceptionOnSpan(ctx context.Context, span *sentry.Span, err error) *sentry.EventID {
	span.Status = sentry.SpanStatusInternalError
	span.SetData("exception.type", fmt.Sprintf("%T", err))
	span.SetData("exception.message", err.Error())

	traceContext := sentry.Context{
		"trace_id": span.TraceID,
		"span_id":  span.SpanID,
		"op":       span.Op,
		"status":   span.Status,
	}
	if span.ParentSpanID != (sentry.SpanID{}) {
		traceContext["parent_span_id"] = span.ParentSpanID
	}
	if span.Description != "" {
		traceContext["description"] = span.Description
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	var eventID *sentry.EventID
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetContext("trace", traceContext)
		eventID = hub.CaptureException(err)
	})
	if eventID != nil {
		span.SetData("sentry.event_id", string(*eventID))
	}
	return eventID
}
