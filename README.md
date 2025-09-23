Languages: **English** | [简体中文](README.zh.md)

# Open source APM: OpenTelemetry traces, metrics, and logs

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)
[![Slack](https://img.shields.io/badge/slack-uptrace.svg?logo=slack)](https://join.slack.com/t/uptracedev/shared_invite/zt-3e35d4b0m-zfAew95ymE5Fv31LwvyuoQ)

Uptrace is an [open source APM](https://uptrace.dev/get/hosted/open-source-apm)
that supports distributed tracing, metrics, and logs. You can use it to monitor
applications and troubleshoot issues.

Uptrace comes with an intuitive query builder, rich dashboards, alerting rules,
notifications, and integrations for most languages and frameworks.

Uptrace can process billions of spans and metrics on a single server and allows
you to monitor your applications at 10x lower cost.

Uptrace uses OpenTelemetry framework to collect data and ClickHouse database to
store it. It also requires PostgreSQL database to store metadata such as metric
names and alerts.

**Features**:

- Single UI for traces, metrics, and logs.
- 50+ pre-built dashboards that are automatically created once metrics start
  coming in.
- Service graph and
  [chart annotations](https://uptrace.dev/features/annotations).
- Spans/logs/metrics [monitoring](https://uptrace.dev/features/alerting) with
  notifications via Email, Slack, WebHook, and AlertManager.
- SQL-like query language to
  [aggregate spans](https://uptrace.dev/features/querying/spans).
- Promql-like language to
  [aggregate metrics](https://uptrace.dev/features/querying/metrics).
- Data ingestion using
  [OpenTelemetry](https://uptrace.dev/ingest/opentelemetry),
  [Prometheus](https://uptrace.dev/ingest/prometheus),
  [Vector](https://uptrace.dev/ingest/vector),
  [FluentBit](https://uptrace.dev/ingest/logs/fluentbit),
  [CloudWatch](https://uptrace.dev/ingest/cloudwatch), and more.
- [Grafana](https://uptrace.dev/features/grafana) compatibility. You can
  configure Grafana to use Uptrace as a Tempo/Prometheus datasource.
- Managing users/projects via YAML config.
- Single sign-on (SSO) using OpenID Connect:
  [Keycloak](https://uptrace.dev/features/sso/keycloak),
  [Google Cloud](https://uptrace.dev/features/sso/google), and
  [Cloudflare](https://uptrace.dev/features/sso/cloudflare).
- Efficient processing: more than 10K spans / second on a single core.
- Excellent on-disk compression: 1KB span can be compressed down to ~40 bytes.

**System overview**

![System overview](./example/docker/images/home.png)

**Faceted filters**

![Faceted filters](./example/docker/images/facets.png)

**Metrics**

![Metrics](./example/docker/images/metrics.png)

**Alerts**

![Alerts](./example/docker/images/alerts.png)

## Quickstart

You can try Uptrace in just a few minutes by visiting the
[cloud demo](https://app.uptrace.dev/play) (no login required) or by
[running](https://github.com/uptrace/uptrace/tree/master/example/docker) it
locally with Docker.

Then follow the [getting started](https://uptrace.dev/get) guide.

## Help

Have questions? Get help via [Telegram](https://t.me/uptrace),
[Slack](https://join.slack.com/t/uptracedev/shared_invite/zt-3e35d4b0m-zfAew95ymE5Fv31LwvyuoQ),
or [start a discussion](https://github.com/uptrace/uptrace/discussions) on
GitHub.

## Contributing

<a href="https://github.com/uptrace/uptrace/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=uptrace/uptrace" />
</a>
