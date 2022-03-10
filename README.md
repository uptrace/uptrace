# Distributed tracing backend using OpenTelemetry and ClickHouse

Uptrace is a distributed tracing system that uses OpenTelemetry to collect data and ClickHouse
database to store it. ClickHouse is the only dependency.

Uptrace comes in 2 versions:

- This open source version which only supports distributed tracing.
- [Cloud](https://uptrace.dev/) version that supports both tracing and metrics.

> :star: Looking for a ClickHouse client? Uptrace uses
> [go-clickhouse](https://github.com/uptrace/go-clickhouse).

<p align="center">
  <a href="https://uptrace.dev/open-source/?autoplay">
    <img src="https://uptrace.dev/uptrace-os/poster.png" alt="Distributed tracing, errors, and logs">
  </a>
</p>

**Features**:

- OpenTelemetry protocol via gRPC (`:14317`) and HTTP (`:14318`)
- Span/Trace grouping
- SQL-like query language
- Errors/logs support
- Percentiles
- Systems, services, and hostnames dashboards
- Multiple users/projects via YAML config
- Sampling/adjusted counts support

**Roadmap**:

- ClickHouse cluster support in the database schema
- TLS support

## Getting started

- [Docker example](example/docker) allows to run Uptrace with a single command.
- [Installation](https://docs.uptrace.dev/guide/os.html) guide with pre-compiled binaries for Linux,
  MacOS, and Windows.

## Compiling Uptrace manually

To compile and run Uptrace locally, you need Go **1.18** and ClickHouse 21.11+.

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

You can also run Uptrace in debug mode by providing an environment variable:

```shell
DEBUG=2 go run cmd/uptrace/main.go serve
```

TO learn about available commands:

```shell
go run cmd/uptrace/main.go help
```

## Compiling UI manually

You can also start the UI locally:

```shell
cd vue
pnpm install
pnpm serve
```

And open http://localhost:19876

## FAQ

**What is the license?**

The Business Source [License](LICENSE) is identical to Apache 2.0 with the only exception being that
you can't use the code to create a cloud service or, in other words, resell it to others as a
product. It is a more permissive license than, for example, AGPL, because it allows private changes
to the code.

You can learn more about BSL [here](https://mariadb.com/bsl-faq-adopting/).

**Are there 2 versions of Uptrace?**

Yes, having 2 separate versions allows us to have minimal number of dependencies (ClickHouse) and
keep the codebase small and fun to work with.

**Is the database schema stable?**

No, we are still making changes to the database schema and hoping to switch to
[ClickHouse dynamic subcolumns](https://github.com/ClickHouse/ClickHouse/pull/23932) when that
feature is available.
