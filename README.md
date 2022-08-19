# OpenTelemetry distributed tracing tool that monitors performance, errors, and logs

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://discordapp.com/api/guilds/1000404569202884628/widget.png)](https://discord.gg/YF8tdP8Pmk)

Uptrace is an OpenTelemetry distributed tracing tool that monitors performance, errors, and logs. It
uses OpenTelelemetry to collect data and ClickHouse database to store it.

<p align="center">
  <a href="https://uptrace.dev/open-source/?autoplay">
    <img src="https://uptrace.dev/uptrace-os/poster.png" alt="Distributed tracing, errors, and logs">
  </a>
</p>

**Features**:

- OpenTelemetry [tracing](https://uptrace.dev/opentelemetry/distributed-tracing.html),
  [metrics](https://uptrace.dev/opentelemetry/metrics.html), and logs.
- [Grafana](https://uptrace.dev/get/opentelemetry-metrics-grafana.html) integration to browse
  [OpenTelemetry Metrics](https://uptrace.dev/opentelemetry/metrics.html).
- Email/Slack notifications using Prometheus AlertManager.
- Span/Trace grouping.
- SQL-like query language.
- Charts and Percentiles.
- Systems, services, and hostnames dashboards.
- Multiple users/projects via YAML config.

**Ingestion**:

- OpenTelemetry protocol via gRPC (`:14317`) and HTTP (`:14318`).
- Zipkin protocol support on `http://uptrace:14318/api/v2/spans`.
- [Vector Logs](example/vector-logs) API support.

**Roadmap**:

- ClickHouse S3 storage
- mTLS support

## Getting started

- [Docker example](example/docker) to try Uptrace with a single command
- [Installation](https://uptrace.dev/get/install.html) guide with pre-compiled binaries for Linux,
  MacOS, and Windows

We also provide guides for the most popular frameworks:

- [Gin and GORM](https://uptrace.dev/get/opentelemetry-gin-gorm.html)
- [Django](https://uptrace.dev/get/opentelemetry-django.html)
- [Flask and SQLAlchemy](https://uptrace.dev/get/opentelemetry-flask-sqlalchemy.html)
- [Rails](https://uptrace.dev/get/opentelemetry-rails.html)

## FAQ

**What is the license?**

The Business Source [License](LICENSE) is identical to Apache 2.0 with the only exception being that
you can't use the code to create a cloud service or, in other words, resell the product to others.

BSL is a more permissive license than, for example, AGPL, because it allows private changes to the
code.

In three years, the code also becomes available under Apache 2.0 license. You can learn more about
BSL [here](https://mariadb.com/bsl-faq-adopting/).

**Is the database schema stable?**

Yes, but we are still making changes to the database schema and plan to switch to
[ClickHouse dynamic subcolumns](https://github.com/ClickHouse/ClickHouse/pull/23932) when that
feature is stable enough.

## Contributing

See [Contributing to Uptrace](https://uptrace.dev/get/contributing.html).
