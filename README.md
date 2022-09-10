# Distributed Tracing, Metrics, and Logs using OpenTelemetry and ClickHouse

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)

Uptrace is an open-source APM tool that supports distributed tracing, metrics, and logs. Uptrace
uses OpenTelelemetry to collect data and ClickHouse database to store it. ClickHouse is the only
dependency.

**Features**:

- Spans/logs grouping.
- SQL-like query language to aggregate spans.
- Promql-like language to aggregate and monitor metrics.
- Email/Slack/PagerDuty notifications using AlertManager.
- Pre-built metrics dashboards.
- Multiple users/projects via YAML config.

**Why Uptrace?**

- Single UI for traces, metrics, and logs.
- Efficient ingestion: more than 10K spans / second on a single core.
- Excellent on-disk compression with ZSTD, for example, 1KB span can be compressed down to <40
  bytes.
- S3 storage support with ability to automatically upload cold data to S3-like storage.

![Uptrace Home](https://uptrace.dev/get/uptrace/home.png)

## Quickstart

Spend few minutes to decide if you need Uptrace by running a
[docker-compose example](example/docker). You can also play with public
[Uptrace demo](https://app.uptrace.dev/play).

Then follow [getting started guide](https://uptrace.dev/get/get-started.html) to properly setup
Uptrace by downloading a Go binary or installing a DEB/RPM package.

## Telegram

Have questions? Get help via [Telegram](https://t.me/uptrace) or
[start a discussion](https://github.com/uptrace/uptrace/discussions) on GitHub.

## FAQ

**What is the license?**

The Business Source [License](LICENSE) is identical to Apache 2.0 with the only exception being that
you can't use the code to create a cloud service or, in other words, resell the product to others.

BSL is adopted by MariaDB, Sentry, CockroachDB, Couchbase and many others. In most cases, it is a
more permissive license than, for example, AGPL, because it allows you to make private changes to
the code.

In three years, the code also becomes available under Apache 2.0 license. You can learn more about
BSL [here](https://mariadb.com/bsl-faq-adopting/).

**Can I use Uptrace to monitor commercial or production-grade applications?**

Yes, you can use Uptrace to monitor commercial applications and provide your employees access to the
Uptrace app without obligations to pay anything.

**Is the database schema stable?**

Yes, but we are still making changes to the database schema and plan to switch to
[ClickHouse dynamic subcolumns](https://github.com/ClickHouse/ClickHouse/pull/23932) when that
feature is
[stable](https://github.com/ClickHouse/ClickHouse/issues?q=is%3Aissue+is%3Aopen+label%3Acomp-type-object)
enough.

## Contributing

See [Contributing to Uptrace](https://uptrace.dev/get/contributing.html).
