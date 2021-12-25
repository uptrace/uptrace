# Distributed tracing backend using OpenTelemetry and ClickHouse

Uptrace is a distributed tracing system that uses OpenTelemetry to collect data and ClickHouse
database to store it. ClickHouse is the only dependency.

_Screenshot goes here_

Features:

- OpenTelemetry protocol via gRPC (`:14317`) and HTTP (`:14318`)
- Span/Trace grouping
- SQL-like query language
- Percentiles
- Systems dashboard

Roadmap:

- Errors/logs support
- More dashboards for services and hosts
- ClickHouse cluster support
- TLS support
- Improved SQL support using CockroachDB SQL parser

## Getting started

- [Docker example](example/docker) allows to run Uptrace with a single command.
- [Installation](https://docs.uptrace.dev/guide/os.html) guide with pre-compiled binaries for Linux,
  MacOS, and Windows.

## Running Uptrace locally

To run Uptrace locally, you need Go **1.18** and ClickHouse.

**Step 1**. Create `uptrace` ClickHouse database:

```shell
clickhouse-client -q "CREATE DATABASE uptrace"
```

**Step 2**. Reset ClickHouse database schema:

```shell
go run cmd/uptrace/main.go ch reset
```

**Step 3**. Start Uptrace:

```shell
go run cmd/uptrace/main.go serve
```

**Step 4**. Open Uptrace UI at http://localhost:14318

Uptrace will monitor itself using [uptrace-go](https://github.com/uptrace/uptrace-go) OpenTelemetry
distro. To get some test data, just reload the UI few times.

## Running UI locally

You can also start the UI locally:

```shell
cd vue
pnpm install
pnpm serve
```

And open http://localhost:19876
