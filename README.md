# Distributed tracing backend using OpenTelemetry and ClickHouse

Uptrace is a distributed tracing system that uses OpenTelemetry to collect data and ClickHouse
database to store it. ClickHouse is the only dependency.

<p align="center">
  <a href="https://uptrace.dev/open-source/?autoplay">
    <img src="https://uptrace.dev/uptrace-os/poster.png" alt="Distributed tracing, errors, and logs">
  </a>
</p>

**Features**:

- OpenTelemetry protocol via gRPC (`:14317`) and HTTP (`:14318`)
- Span/Trace grouping
- SQL-like query language
- Percentiles
- Systems dashboard
- Multiple users/projects via YAML config

**Roadmap**:

- Errors/logs support
- More dashboards for services and hosts
- ClickHouse cluster support
- TLS support
- Sampling/adjusted counts support
- Improved SQL support using CockroachDB SQL parser

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

You can run Uptrace in debug mode by providing an environment variable:

```shell
DEBUG=2 go run cmd/uptrace/main.go serve
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

The Business Source License [license](LICENSE) is identical to Apache 2.0 with the only exception
being that you can't use the code to create a cloud service. It is a more permissive license than,
for example, AGPL, because it allows private changes to the code.

You can learn more about BSL [here](https://mariadb.com/bsl-faq-adopting/).

**Are there 2 versions of Uptrace?**

Yes, having 2 separate versions allows us to have minimal number of dependencies (ClickHouse) and
keep the codebase small and fun to work with.
