# Changelog

To get started with Uptrace, see https://uptrace.dev/get/get-started.html

## v1.2.0 - Unreleased

- Added 2 quick filters by `deployment.environment` and `service.name` attributes. More to come in
  the next release.

- Tweaked spans grouping and added 2 related options:

  - `project.group_by_env` - group spans by `deployment.environment` attribute.
  - `project.group_funcs_by_service` - group funcs spans by `service.name` attribute.

- Changed ClickHouse schema to not use column names with dots in them which causes issues with
  migrations, for example, such columns could not be renamed.

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
