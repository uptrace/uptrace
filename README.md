Languages: **English** | [简体中文](README.zh.md)

# Open source APM: OpenTelemetry traces, metrics, and logs

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)
[![Slack](https://img.shields.io/badge/slack-uptrace.svg?logo=slack)](https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw)

Uptrace is an [open source APM](https://uptrace.dev/get/open-source-apm.html) that supports
distributed tracing, metrics, and logs. You can use it to monitor applications and troubleshoot
issues.

Uptrace comes with an intuitive query builder, rich dashboards, alerting rules, notifications, and
integrations for most languages and frameworks.

Uptrace can process billions of spans and metrics on a single server and allows you to monitor your
applications at 10x lower cost.

Uptrace uses OpenTelemetry framework to collect data and ClickHouse database to store it. It also
requires PostgreSQL database to store metadata such as metric names and alerts.

**Features**:

- Single UI for traces, metrics, and logs.
- 50+ pre-built dashboards that are automatically created once metrics start coming in.
- Service graph and [chart annotations](https://uptrace.dev/get/annotations.html).
- Spans/logs/metrics [monitoring](https://uptrace.dev/get/alerting.html) with notifications via
  Email, Slack, WebHook, and AlertManager.
- SQL-like query language to [aggregate spans](https://uptrace.dev/get/querying-spans.html).
- Promql-like language to [aggregate metrics](https://uptrace.dev/get/querying-metrics.html).
- Data ingestion using [OpenTelemetry](https://uptrace.dev/get/ingest/opentelemetry.html),
  [Prometheus](https://uptrace.dev/get/ingest/prometheus.html),
  [Vector](https://uptrace.dev/get/ingest/vector.html),
  [FluentBit](https://uptrace.dev/get/ingest/fluent-bit.html),
  [CloudWatch](https://uptrace.dev/get/ingest/aws-cloudwatch.html), and more.
- [Grafana](https://uptrace.dev/get/grafana.html) compatibility. You can configure Grafana to use
  Uptrace as a Tempo/Prometheus datasource.
- Managing users/projects via YAML config.
- Single sign-on (SSO) using OpenID Connect: [Keycloak](https://uptrace.dev/get/sso/keycloak.html),
  [Google Cloud](https://uptrace.dev/get/sso/google.html), and
  [Cloudflare](https://uptrace.dev/get/sso/cloudflare.html).
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

You can try Uptrace in just a few minutes by visiting the [cloud demo](https://app.uptrace.dev/play)
(no login required) or by [running](https://github.com/uptrace/uptrace/tree/master/example/docker)
it locally with Docker.

Then follow the [getting started](https://uptrace.dev/get/get-started.html) guide.

## Help

Have questions? Get help via [Telegram](https://t.me/uptrace),
[Slack](https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw), or
[start a discussion](https://github.com/uptrace/uptrace/discussions) on GitHub.

## Contributing

See [Contributing to Uptrace](https://uptrace.dev/get/contributing.html).

<a href="https://github.com/uptrace/uptrace/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=uptrace/uptrace" />
</a>

### Infrastructure model
![Infrastructure main model](.infragenie/infrastructure_main_model.svg)
- [prometheus component model](.infragenie/prometheus_component_model.svg)

---
