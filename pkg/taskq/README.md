# Golang asynchronous task/job queue with Redis, SQS, IronMQ, and in-memory backends

![build workflow](https://github.com/vmihailenco/taskq/actions/workflows/build.yml/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/vmihailenco/taskq/v3)](https://pkg.go.dev/github.com/vmihailenco/taskq/v3)
[![Documentation](https://img.shields.io/badge/bun-documentation-informational)](https://taskq.uptrace.dev/)
[![Chat](https://discordapp.com/api/guilds/752070105847955518/widget.png)](https://discord.gg/rWtp5Aj)

> taskq is brought to you by :star: [**uptrace/uptrace**](https://github.com/uptrace/uptrace).
> Uptrace is an open source and blazingly fast
> [distributed tracing tool](https://get.uptrace.dev/compare/distributed-tracing-tools.html) powered
> by OpenTelemetry and ClickHouse. Give it a star as well!

## Features

- Redis, SQS, IronMQ, and in-memory backends.
- Automatically scaling number of goroutines used to fetch (fetcher) and process messages (worker).
- Global rate limiting.
- Global limit of workers.
- Call once - deduplicating messages with same name.
- Automatic retries with exponential backoffs.
- Automatic pausing when all messages in queue fail.
- Fallback handler for processing failed messages.
- Message batching. It is used in SQS and IronMQ backends to add/delete messages in batches.
- Automatic message compression using snappy / s2.

Resources:

- [**Get started**](https://taskq.uptrace.dev/guide/golang-task-queue.html)
- [Examples](https://github.com/vmihailenco/taskq/tree/v3/example)
- [Discussions](https://github.com/uptrace/bun/discussions)
- [Chat](https://discord.gg/rWtp5Aj)
- [Reference](https://pkg.go.dev/github.com/vmihailenco/taskq/v3)

## Getting started

To get started, see [Golang Task Queue](https://taskq.uptrace.dev/) documentation.

**Producer**:

```go
import (
    "github.com/vmihailenco/taskq/v3"
    "github.com/vmihailenco/taskq/v3/redisq"
)

// Create a queue factory.
var QueueFactory = redisq.NewFactory()

// Create a queue.
var MainQueue = QueueFactory.RegisterQueue(&taskq.QueueOptions{
    Name:  "api-worker",
    Redis: Redis, // go-redis client
})

// Register a task.
var CountTask = taskq.RegisterTask("counter", &taskq.TaskOptions{
    Handler: func() error {
        IncrLocalCounter()
        return nil
    },
})

ctx := context.Background()

// And start producing.
for {
	// Call the task without any args.
	err := MainQueue.AddJob(ctx, CountTask.NewJob())
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
}
```

**Consumer**:

```go
// Start consuming the queue.
if err := MainQueue.Start(context.Background()); err != nil {
    log.Fatal(err)
}
```

## See also

- [Golang ORM](https://github.com/uptrace/bun) for PostgreSQL, MySQL, MSSQL, and SQLite
- [Golang PostgreSQL](https://bun.uptrace.dev/postgres/)
- [Golang HTTP router](https://github.com/uptrace/bunrouter)
- [Golang ClickHouse](https://github.com/uptrace/go-clickhouse)

## Contributors

Thanks to all the people who already contributed!

<a href="https://github.com/vmihailenco/taskq/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=vmihailenco/taskq" />
</a>
