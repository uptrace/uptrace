package pgq_test

import (
	"context"
	"database/sql"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
	"github.com/vmihailenco/taskq/pgq/v4"
	"github.com/vmihailenco/taskq/taskqtest/v4"
	"github.com/vmihailenco/taskq/v4"
)

func queueName(s string) string {
	version := strings.Split(runtime.Version(), " ")[0]
	version = strings.Replace(version, ".", "", -1)
	return "test-" + s + "-" + version
}

func newFactory(t *testing.T) taskq.Factory {
	dsn := "postgres://test:test@localhost:5432/test?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.FromEnv("BUNDEBUG")))

	f, err := os.Open("migration.sql")
	require.NoError(t, err)

	err = migrate.Exec(context.Background(), db, f, false)
	require.NoError(t, err)

	return pgq.NewFactory(db)
}

func TestConsumer(t *testing.T) {
	taskqtest.TestConsumer(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("consumer"),
	})
}

func TestUnknownTask(t *testing.T) {
	taskqtest.TestUnknownTask(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("unknown-task"),
	})
}

func TestFallback(t *testing.T) {
	taskqtest.TestFallback(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("fallback"),
	})
}

func TestDelay(t *testing.T) {
	taskqtest.TestDelay(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("delay"),
	})
}

func TestRetry(t *testing.T) {
	taskqtest.TestRetry(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("retry"),
	})
}

func TestNamedJob(t *testing.T) {
	taskqtest.TestNamedJob(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("named-message"),
	})
}

func TestCallOnce(t *testing.T) {
	taskqtest.TestCallOnce(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("call-once"),
	})
}

func TestLen(t *testing.T) {
	taskqtest.TestLen(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("queue-len"),
	})
}

func TestRateLimit(t *testing.T) {
	taskqtest.TestRateLimit(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("rate-limit"),
	})
}

func TestErrorDelay(t *testing.T) {
	taskqtest.TestErrorDelay(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("delayer"),
	})
}

func TestBatchConsumerSmallJob(t *testing.T) {
	taskqtest.TestBatchConsumer(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("batch-consumer-small-message"),
	}, 100)
}

func TestBatchConsumerLarge(t *testing.T) {
	taskqtest.TestBatchConsumer(t, newFactory(t), &taskq.QueueConfig{
		Name: queueName("batch-processor-large-message"),
	}, 64000)
}
