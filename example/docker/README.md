# Uptrace Open Source demo

## Getting started

This example demonstrates how to quickly start Uptrace using Docker. To run Uptrace permanently, we
recommend using a DEB/RPM
[package](https://uptrace.dev/get/opentelemetry-tracing-tool.html#packages) or a pre-compiled
[binary](https://uptrace.dev/get/opentelemetry-tracing-tool.html#binaries).

**Step 1**. Download the example using Git:

```shell
git clone https://github.com/uptrace/uptrace.git
cd example/docker
```

**Step 2**. Start the services using Docker:

```shell
docker-compose up -d
```

**Step 3**. Make sure Uptrace is running:

```shell
docker-compose logs uptrace
```

**Step 4**. Open Uptrace UI at http://localhost:14318

Uptrace will monitor itself using [uptrace-go](https://github.com/uptrace/uptrace-go) OpenTelemetry
distro. To get some test data, just reload the UI few times. It usually takes about 30 seconds for
the data to appear.

To configure OpenTelemetry for your programming language, see
[documentation](https://uptrace.dev/get/opentelemetry-tracing-tool.html).

## Grafana

This example comes with a pre-configured Grafana on http://localhost:3000 (`admin:admin`) that
allows you to:

- View metrics using Prometheus data source.
- View traces using Tempo data source.
- View logs using Loki data source.

See [Uptrace Grafana](https://uptrace.dev/get/grafana.html) documentation for details.

## ClickHouse and OpenTelemetry

To trace the ClickHouse database, you can setup a materialized view to export spans from the
[system.opentelemetry_span_log table](https://clickhouse.com/docs/en/operations/system-tables/opentelemetry_span_log):

```shell
docker-compose exec clickhouse clickhouse-client
```

```sql
CREATE MATERIALIZED VIEW default.zipkin_spans
ENGINE = URL('http://uptrace:14318/api/v2/spans', 'JSONEachRow')
SETTINGS output_format_json_named_tuples_as_objects = 1,
    output_format_json_array_of_rows = 1 AS
SELECT
    lower(hex(trace_id)) AS traceId,
    case when parent_span_id = 0 then '' else lower(hex(parent_span_id)) end AS parentId,
    lower(hex(span_id)) AS id,
    operation_name AS name,
    start_time_us AS timestamp,
    finish_time_us - start_time_us AS duration,
    cast(tuple('clickhouse'), 'Tuple(serviceName text)') AS localEndpoint,
    attribute AS tags
FROM system.opentelemetry_span_log
```

See ClickHouse [documentation](https://clickhouse.com/docs/en/operations/opentelemetry/) for
details.
