# Changelog

To get started with Uptrace, see https://uptrace.dev/get/get-started.html

## v1.6.0 UNRELEASED

This release is partially backwards incompatible with v1.5. PostgreSQL database will be migrated
automatically preserving your settings and dashboards, but you will need to reset the ClickHouse
database schema to upgrade:

```shell
uptrace ch reset
```

The recommended upgrade path is to start a separate Uptrace v1.6 instance in parallel with v1.5,
write data to both instances, and switch when you have enough data in the new instance. This is more
or less what we do when deploying major changes to Uptrace Cloud.

The license is changed from BSL to AGPL v3.0 which is a license approved by Open Software
Foundation.

#### Breaking

- ClickHouse database schema is changed.

- Metric names and attributes are automatically changed to comply with Prometheus/Loki restrictions,
  for example, `service.name` becomes `service_name`.

- Upgraded to
  [v1.21](https://github.com/open-telemetry/opentelemetry-specification/blob/main/schemas/1.21.0)
  OpenTelemetry semantic conventions which introduced some breaking changes to attribute names. Most
  notably:

  - `http.method` is renamed to `http.request.method`.
  - `http.status_code` is renamed to `http.response.status_code`.

  Uptrace will automatically rename attributes, but you may need to update your favorite queries.

#### Features

- Added [service graphs](https://app.uptrace.dev/overview/1/service-graph). Service Graphs provide a
  visual representation of service interactions, dependencies, and performance metrics. Service
  graphs are built by analyzing span relationships and require certain span attributes.

- Added ability to group dashboard charts into rows, for example, charts can be grouped by category
  such as CPU, RAM, Network metrics.

- Added support for [Prometheus remote write](https://uptrace.dev/get/ingest/prometheus.html), which
  allows Prometheus to send its collected metrics data directly to a long-term storage solution such
  as Uptrace.

- Added ability to use Uptrace as a
  [Prometheus datasource in Grafana](https://uptrace.dev/get/grafana.html). Uptrace uses the
  original Prometheus engine so all Prometheus queries should be supported and you should be able to
  use existing Grafana dashboards with the Uptrace data source.

- Added [chart annotations](https://uptrace.dev/get/annotations.html) support. Annotations are
  labels or notes added to a chart to provide additional information or context. Annotations help
  clarify the data presented in the chart and help the viewer understand key points or trends.

#### Improvements

- Simplified UI for building dashboards and metric monitors.
- Added support for `per_min(sum(attr_key))` expressions in tracing query language.
- Added support for simple expressions like `sum(.duration) / .count` which is the same as
  `avg(.duration)`.
- Uptrace DSN now contains gRPC port so the same DSN can be used by all OpenTelemetry distributions
  provided by Uptrace such as uptrace-go, uptrace-js, etc.
- Improved templates for emails and Slack notifications.
- Added more metrics dashboards for HTTP checks and Kubernetes.
- Added more indexed semantic attributes in ClickHouse.
- Spans to metrics conversion requires a recent ClickHouse version that supports
  `allow_experimental_analyzer = 1`.

#### Changes

- Changed the default spans TTL to 7 days, metrics TTL to 30 days.

- There is now a single TLS config shared by HTTP and GRPC servers and defined in `listen.tls`
  section. The old configuration still works, but the first TLS config will be used.

## v1.5.2 - July 6 2023

- When authenticating via email and password, Uptrace ignores users in the PostgreSQL database and
  only allows users in the YAML config. This is needed when you want to disable email:password
  authentication completely.

## v1.5.0 - June 16 2023

This release is backwards compatible with v1.4.x, but contains a ClickHouse mutation to add new
columns.

#### Features

- Add support for `display.name` attribute. You can now use `display.name contains "get"` to search
  for spans and events/logs at the the same time. You can also use `display.name` attribute to
  [override](https://uptrace.dev/get/grouping.html) default span/event/log names.
- Ported Timeseries tab from [Uptrace Cloud](https://uptrace.dev/cloud) version.
- Allow to explore attribute values, for example, you can click on the `enduser.name` attribute to
  see the list of affected users.
- Metric dashboard templates can now automatically create monitors.
- Added ability to receive metrics and logs from
  [AWS CloudWatch](https://uptrace.dev/get/ingest/aws-cloudwatch.html).
- Uptrace can be used as a [Tempo data source](https://uptrace.dev/get/grafana.html) in Grafana.
  Grafana TraceQL is NOT supported, but Search is fully supported.

#### Improvements

- System picker now shows number of matching groups. Also added a preset for Spans / Logs / Events
  groups.
- `span.` prefix is replaced with `.`, for example, instead of `p50(span.duration)` you can write
  `p50(.duration)`. Old names are deprecated, but still supported.
- Added more metric dashboards and improved existing ones.
- Uptrace config now has sensible defaults so you can start with an empty YAML config file and add
  changes as you go.
- Added `uptrace config dump` to view the current Uptrace config in YAML format.
- Added a Docker image for linux/arm64 in addition to linux/amd64.

#### Fixes

- Preserve filters when navigating from "Overview" to "Tracing" tabs.
- Fixed numeric aggregations over custom attributes, for example, `sum(custom.attribute)`.
- Fixed attributes normalizations, for example, `X-Request-Id` is normalized to `x_request_id`.

#### Other changes

- Renamed `logs.LEVEL` to `logging.LEVEL` so it is not confused with logs processing.

## v1.4.0 - Apr 21 2023

#### Breaking changes

- PostgreSQL database is mandatory now. PostgreSQL is only used to store metadata such as users,
  dashboards, metric names, alerts etc. PostgreSQL DB usually only takes few megabytes of disk
  space.

  SQLite support is removed.

- Alerting rules in the config are ignored. Use Uptrace UI to create
  [monitors](https://uptrace.dev/get/alerting.html) instead.

- AlertManager client config is ignored. Use Uptrace UI to create AlertManager notification channel.

- `auth.users.username` is replaced with `auth.users.email` so users can receive email
  notifications. As the result, emails are used to authenticate users. To set user's name, use
  `auth.users.name` field.

#### Features

- Port Alerts, Monitors, and Notifications Channels from
  [Uptrace Enterprise](https://uptrace.dev/compare) edition.
- Accept errors and spans from [Sentry SDK](https://uptrace.dev/get/ingest/sentry.html).
- Documented [FluentBit](https://uptrace.dev/get/ingest/fluent-bit.html) integration.
- Add filter facets for metrics.

#### Improvements

- Improve UI for switching between table/grid metric views.
- Allow to quickly change group by in the grid.
- Allow to add/edit text gauges.
- Support table visualization in the grid view.
- Support heatmaps in the grid view.
- Allow to customize grid size
- Allow to customize colors for timeseries (just like units).
- Allow to edit dashboards using YAML.
- Add Kafka metrics dashboard.
- Add `site.path` setting to host Uptrace behind a proxy.

#### Fixes

- Fix cumulative to delta metrics conversion.
- Fix exponential histograms handling.
- Fix project token syncing.
- Respect `site.addr` when building OTLP/gRPC and OTLP/HTTP endpoints and DSNs.

### Migrating from previous versions

Uptrace v1.4.0 contains backwards incompatible changes that require resetting database schema.

To migrate to v1.4.0:

1. Stop Uptrace: `sudo systemctl stop uptrace`.
1. [Install](https://uptrace.dev/get/install.html) new version.
1. Configure PostgreSQL database in `uptrace.yml` config.

   ```yaml
   pg:
     addr: localhost:5432
     user: uptrace
     password: uptrace
     database: uptrace
   ```

1. Reset PostgreSQL and ClickHouse databases:

   ```shell
   uptrace pg reset
   uptrace ch reset
   ```

1. Start Uptrace: `sudo systemctl start uptrace`.

## v1.3.0 - Jan 20 2023

- Added ability to parse logs as spans using Vector remap language. See
  [documentation](https://uptrace.dev/get/ingest/vector.html#converting-logs-to-spans) and
  [PostgreSQL](https://dev.to/uptrace/monitoring-postgresql-15-logs-with-vector-and-uptrace-4165)
  example for details.

- Added support for Summary metrics type.

- Zipkin ingestion API now requires an Uptrace DSN in one of the following locations:

  - `uptrace-dsn` HTTP header.
  - `Authorization` HTTP header.
  - `dsn` URL query, for example, `/api/v2/spans?dsn=[dsn]`.

- Allow to configure [uptrace-go](https://uptrace.dev/get/opentelemetry-go.html) to collect Uptrace
  telemetry data:

```yaml
##
## uptrace-go client configuration.
## Uptrace sends internal telemetry here. Defaults to listen.grpc.addr.
##
uptrace_go:
  # dsn: http://project1_secret_token@localhost:14317/1
  # tls:
  #   cert_file: config/tls/uptrace.crt
  #   key_file: config/tls/uptrace.key
  #   insecure_skip_verify: true
```

- Added [sprig](http://masterminds.github.io/sprig/) functions to alerting
  [templates](https://uptrace.dev/get/alerting.html#templating).

- Allow to configure logging level via `logs.level` configuration option, for example:

```yaml
logs:
  level: ERROR
```

## v1.2.0 - Nov 8 2022

- Added ability to filter spans using facets.
- Added ability to select multiple systems.
- Added 2 quick filters by `deployment.environment` and `service.name` attributes on the Overview
  page.
- Added support for [creating metrics from spans](https://uptrace.dev/get/span-metrics.html) so they
  can be monitored like usual metrics, for example:

```yaml
# First, create a metric from incoming spans.
metrics_from_spans:
  - name: uptrace.tracing.spans
    description: Spans count (excluding events)
    instrument: counter
    unit: 1
    value: span.count
    attrs:
      - span.system as system
      - service.name as service
      - host.name as host
      - span.status_code as status
    where: not span.is_event

# Then, monitor that metric.
alerting:
  rules:
    - name: Service has high error rate
      metrics:
        - uptrace.tracing.spans as $spans
      query:
        - $spans{status="error"} / $spans > 0.1 group by service.name
      for: 5m
```

- Alerting rules annotations now support templating just like Prometheus, for example:

```yaml
alerting:
  rules:
    - name: Filesystem usage >= 90%
      metrics:
        - system.filesystem.usage as $fs_usage
      query:
        - group by host.name
        - group by device
        - where device !~ "loop"
        - $fs_usage{state="used"} / $fs_usage >= 0.9
      for: 5m
      annotations:
        summary:
          'FS usage is {{ $values.fs_usage }} on {{ $labels.host_name }} and {{ $labels.device }}'
```

- Tweaked spans grouping and added 2 related options:

  - `project.group_by_env` - group spans by `deployment.environment` attribute.
  - `project.group_funcs_by_service` - group `funcs` spans by `service.name` attribute.

- Added project settings page where you can check available settings and project DSN.

- Changed ClickHouse schema to not use column names with dots in them which causes various issues
  with migrations, for example, such columns can't be renamed.

  If you have an existing ClickHouse database, you will have to reset it with:

```shell
uptrace ch reset
```

## v1.1.0 - Oct 1 2022

- Added additional ways to authenticate users via
  [Keycloak](https://uptrace.dev/get/auth-keycloak.html),
  [Google Cloud](https://uptrace.dev/get/auth-keycloak.html), and
  [Cloudflare](https://uptrace.dev/get/auth-cloudflare.html). Contributed by
  [@aramperes](https://github.com/aramperes).

- Added gauges support to Metrics UI. Only used in Redis Enterprise
  [example](example/redis-enterprise) for now.

- Renamed Logs tab to Events and moved all events there.

- Added support for PostgreSQL database instead of SQLite. This requires resetting SQLite database
  with:

```shell
uptrace db reset
```

- [Docker example](example/docker) is updated to work on Windows.
- Added [Redis Enterprise](example/redis-enterprise) example and metrics dashboards.

## v1.0.2 - Sep 8 2022

- Rename `alertmanager` YAML section to `alertmanager_client` so users don't get confused.

## v1.0.0 - Sep 6 2022

### Added

- Accept and store [OpenTelemetry Metrics](https://uptrace.dev/opentelemetry/metrics.html) in
  ClickHouse.
- Added metrics monitoring using [alerting rules](https://uptrace.dev/get/alerting.html).
- Added ability to send notifications via email/Slack/Telegram using
  [AlertManager](https://uptrace.dev/get/alerting.html#AlertManager).
- Added ability to configure [TLS](https://uptrace.dev/get/config.html#tls).
- Expand env vars in the YAML config, for example:

```yaml
ch:
  dsn: 'clickhouse://${CLICKHOUSE_USER}:@${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/${CLICKHOUSE_DB}?sslmode=disable'
```

### Upgrading

To upgrade, grab the latest
[uptrace.yml](https://github.com/uptrace/uptrace/blob/master/config/uptrace.yml) config, reset
ClickHouse database, and restart Uptrace:

```shell
uptrace ch reset
sudo systemctl restart uptrace
```

## v0.2.15 - Jun 8 2022

### 💡 Enhancements 💡

- Added support for accepting Zipkin spans at `http://uptrace:14318/api/v2/spans`.
- Added support for accepting Vector logs. See the [example](example/vector-logs).
- Added "Slowest groups" to the Overview tab
- Added new config option `ch_schema.compression`. You can now set ClickHouse compression via
  Uptrace config.
- Added new config option `ch_schema.replicated` if you want to use ClickHouse replication.
- Renamed the config option `retention.ttl` to `ch_schema.ttl`.
- Added new config option `spans.buffer_size`.
- Added new config option `spans.batch_size`.

## v0.2.14 - Apr 19 2022

- Fix incorrect `ORDER BY` when focusing on spans.
- Parse HTTP user agent into smaller parts.
- Always show `service.name` attribute when viewing traces.

## v0.2.13 - Apr 7 2022

- Fix incorrect `ORDER BY` when viewing spans.
- Improve config file auto-discovery.
- Update msgpack library.

## v0.2.12 - Mar 30 2022

- Automatically run ClickHouse migrations when Uptrace is started.
- Added 15 and 30 minutes periods.

## v0.2.11 - Mar 29 2022

- Added ability to filter spans by clicking on a chip.
- Added explore menu for each span attribute.
- Better handle situations when `service.name` or `host.name` attributes are not available.
- Support ZSTD and snappy decompression in OTLP.

## v0.2.8 - Mar 15 2022

- Fixed duration filter.
- Added chart resizing when window size is changed.
- Added GRPC metrics stub to remove errors from logs.

## v0.2.5 - Mar 09 2022

- Added list of spans.
- Fixed links to services and hostnames.
- Fixed SQL grouping.

## v0.2.4 - Feb 24 2022

### 💡 Enhancements 💡

- Make sure projects have unique tokens.
- Make user authentication optional by commenting out users section in the YAML config.
- Fixed jumping to a trace to account for the project id.
- Added missing trace find route.

## v0.2.2 - Feb 22 2022

### 💡 Enhancements 💡

- Added log out button.
- Added more concise syntax support to Uptrace query language, for example,
  `{p50,p90,p99}(span.duration)` instead of
  `p50(span.duration), p90(span.duration), p99(span.duration)`.
- Improved Uptrace query parsing.

## v0.2.0 - Jan 25 2022

### 💡 Enhancements 💡

- Added support for exceptions and in-app logs. See
  [Zap](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelzap) and
  [logrus](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otellogrus) integrations.
- Added services and hostnames overview.
- Added SQL query formatting when viewing spans.
- Require user authentication. Users are defined in the YAML config.
- Added support for having multiple isolated projects in the same database. Projects are defined in
  the YAML config.
- Added ability to filter query results, for example,
  `group by span.group_id | p50(span.duration) | where p50(span.duration) > 10ms`.
- Improved error handling on invalid Uptrace queries.
- Use faster and more compact MessagePack encoding to store spans in `spans_data` table.
- Add more attributes to ClickHouse index.

To upgrade, reset ClickHouse schema with the following command (existing data will be lost):

```go
# Using binary
./uptrace --config=/etc/uptrace/uptrace.yml ch reset

# Using sources
go run cmd/uptrace/main.go --config=config/uptrace.yml ch reset
```

## v0.1.0 - Dec 27 2021

Initial release.
