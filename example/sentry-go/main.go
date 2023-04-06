package main

import (
	"context"
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

	sentry.CaptureMessage("It works!")

	ctx := context.Background()

	span := sentry.StartSpan(ctx, "doWork",
		sentry.TransactionName(fmt.Sprintf("doWork: %s", "hello")))
	defer span.Finish()

	{
		ctx := span.Context()
		span := sentry.StartSpan(ctx, "suboperation1")

		{
			span := sentry.StartSpan(span.Context(), "suboperation3")
			span.Finish()
		}

		span.Finish()

		span = sentry.StartSpan(ctx, "suboperation2")
		span.Finish()
	}
}
