# Distributed tracing using OpenTelemetry and ClickHouse

Uptrace Open Source version is a distributed tracing system that uses OpenTelemetry to collect data
and ClickHouse database to store it. ClickHouse is the only dependency.

_Screenshot goes here_

Features:

- OpenTelemetry traces via OTLP
- Span/Trace grouping
- SQL-like query language
- Percentiles
- Systems dashboard

Roadmap:

- Errors/logs support
- More dashboards for services and hosts
- ReplicatedMergeTree support
- TLS support
- Improved SQL support using CockroachDB SQL parser (if license permits)

## Getting started

- [Docker example](example/docker) allows to run Uptrace with a single command.
- [Installation]() guide with pre-compiled binaries for Linux, MacOS, and Windows.

## Running locally

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

```
$ go run cmd/uptrace/main.go serve
reading config from ./uptrace.yml
serving on http://localhost:15678/ UPTRACE_DSN=http://localhost:4317 OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

**Step 4**. Open Uptrace UI at http://localhost:15678

Uptrace will monitor itself using [uptrace-go](https://github.com/uptrace/uptrace-go) OpenTelemetry
distro. To get some test data, just reload the UI few times.
