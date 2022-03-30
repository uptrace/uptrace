# Changelog

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
- Added `SAMPLE BY` to `spans_index` table.
- Added query limits to `spans_index` queries to better support large datasets.
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
