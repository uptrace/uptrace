package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "http://project2_secret_token@localhost:14318/2",
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		panic(err)
	}
	defer sentry.Flush(3 * time.Second)

	eventId := sentry.CaptureException(errors.New("Yeah, it works!"))
	fmt.Println("exception id:", *eventId)

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
	span.Finish()

	fmt.Println("trace id:", span.TraceID)
}
