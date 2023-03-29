package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"
)

var tracer = otel.Tracer("github.com/go-redis/redis/example/otel")

func main() {
	ctx := context.Background()

	uptrace.ConfigureOpentelemetry(
		// copy your project DSN here or use UPTRACE_DSN env var
		uptrace.WithDSN("http://project2_secret_token@localhost:14317/2"),

		uptrace.WithServiceName("myservice"),
		uptrace.WithServiceVersion("v1.0.0"),
	)
	defer uptrace.Shutdown(ctx)

	rdb := redis.NewClient(&redis.Options{
		Addr: ":6666",
	})
	if err := monitorKvrocks(ctx, rdb); err != nil {
		panic(err)
	}
	if err := redisotel.InstrumentTracing(rdb, redisotel.WithDBSystem("kvrocks")); err != nil {
		panic(err)
	}
	if err := redisotel.InstrumentMetrics(rdb, redisotel.WithDBSystem("kvrocks")); err != nil {
		panic(err)
	}

	for i := 0; i < 1e6; i++ {
		ctx, rootSpan := tracer.Start(ctx, "handleRequest")

		if err := handleRequest(ctx, rdb); err != nil {
			rootSpan.RecordError(err)
			rootSpan.SetStatus(codes.Error, err.Error())
		}

		rootSpan.End()

		if i == 0 {
			fmt.Printf("view trace: %s\n", uptrace.TraceURL(rootSpan))
		}

		time.Sleep(time.Second)
	}
}

func handleRequest(ctx context.Context, rdb *redis.Client) error {
	if err := rdb.Set(ctx, "First value", "value_1", 0).Err(); err != nil {
		return err
	}
	if err := rdb.Set(ctx, "Second value", "value_2", 0).Err(); err != nil {
		return err
	}

	var group sync.WaitGroup

	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			val := rdb.Get(ctx, "Second value").Val()
			if val != "value_2" {
				log.Printf("%q != %q", val, "value_2")
			}
		}()
	}

	group.Wait()

	if err := rdb.Del(ctx, "First value").Err(); err != nil {
		return err
	}
	if err := rdb.Del(ctx, "Second value").Err(); err != nil {
		return err
	}

	return nil
}

var re = regexp.MustCompile(`used_disk_percent:\s(\d+)%`)

func monitorKvrocks(ctx context.Context, rdb *redis.Client) error {
	mp := global.MeterProvider()
	meter := mp.Meter("github.com/uptrace/uptrace/example/kvrocks")

	usedDiskPct, err := meter.Float64ObservableGauge(
		"kvrocks.used_disk_percent",
		instrument.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	if _, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			pct, err := getUsedDiskPercent(ctx, rdb)
			if err != nil {
				return err
			}
			o.ObserveFloat64(usedDiskPct, pct, semconv.DBSystemKey.String("kvrocks"))
			return nil
		},
		usedDiskPct,
	); err != nil {
		return err
	}

	return nil
}

func getUsedDiskPercent(ctx context.Context, rdb *redis.Client) (float64, error) {
	info, err := rdb.Info(ctx, "keyspace").Result()
	if err != nil {
		return 0, err
	}

	m := re.FindStringSubmatch(info)
	if m == nil {
		return 0, errors.New("can't find used_disk_percent metric")
	}

	n, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return float64(n) / 100, nil
}
