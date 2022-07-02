# OpenTelemetry distributed tracing tool that monitors performance, errors, and logs

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/matrix/uptrace:matrix.org)](https://matrix.to/#/#uptrace:matrix.org)

Uptrace is an OpenTelemetry distributed tracing tool that monitors performance, errors, and logs. It
uses OpenTelelemetry to collect data and ClickHouse database to store it.

Uptrace comes in 2 versions:

- This open source version which only supports distributed tracing.
- [Cloud](https://uptrace.dev/) version that, in addition to tracing, also supports metrics.

> :star: Looking for a ClickHouse client? Uptrace uses
> [go-clickhouse](https://github.com/uptrace/go-clickhouse).

<p align="center">
  <a href="https://uptrace.dev/open-source/?autoplay">
    <img src="https://uptrace.dev/uptrace-os/poster.png" alt="Distributed tracing, errors, and logs">
  </a>
</p>

**Features**:

- OpenTelemetry tracing and logs via gRPC (`:14317`) and HTTP (`:14318`)
- Zipkin protocol support on `http://uptrace:14318/api/v2/spans`.
- [Vector Logs](example/vector-logs) support.
- Span/Trace grouping
- SQL-like query language
- Errors/logs support
- Percentiles
- Systems, services, and hostnames dashboards
- Multiple users/projects via YAML config
- Sampling/adjusted counts support

**Roadmap**:

- Metrics support
- ClickHouse S3 storage
- Email notifications
- mTLS support

## Getting started

- [Docker example](example/docker) to try Uptrace with a single command
- [Installation](https://get.uptrace.dev/guide/) guide with pre-compiled binaries for Linux, MacOS,
  and Windows

We also provide guides for the most popular frameworks:

- [Gin+GORM example](example/gin-gorm)
- [Django example](example/django)
- [Flask example](example/flask)
- [Rails example](example/rails)

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
