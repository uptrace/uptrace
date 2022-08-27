# Distributed Tracing and Metrics using OpenTelemetry and ClickHouse

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)

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
- Email/Slack/PagerDuty notifications using AlertManager.
- Span/Trace grouping.
- SQL-like query language to aggregate spans.
- Promql-like language to aggregate metrics.
- Pre-built metrics [dashboards](config/dashboard-templates).
- Multiple users/projects via YAML config.

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

BSL is adopted by MariaDB, Sentry, CockroachDB, Couchbase and many others. In most cases it is a
more permissive license than, for example, AGPL, because it allows private changes to the code.

In three years, the code also becomes available under Apache 2.0 license. You can learn more about
BSL [here](https://mariadb.com/bsl-faq-adopting/).

**Can I use Uptrace to monitor commercial or production-grade applications?**

Yes, you can use Uptrace to monitor commercial applications and grant your employees access to the
Uptrace app without obligations to pay anything.

**Is the database schema stable?**

Yes, but we are still making changes to the database schema and plan to switch to
[ClickHouse dynamic subcolumns](https://github.com/ClickHouse/ClickHouse/pull/23932) when that
feature is
[stable](https://github.com/ClickHouse/ClickHouse/issues?q=is%3Aissue+is%3Aopen+label%3Acomp-type-object)
enough.

## Contributing

See [Contributing to Uptrace](https://uptrace.dev/get/contributing.html).
