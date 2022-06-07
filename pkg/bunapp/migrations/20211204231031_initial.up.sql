CREATE TABLE spans_index (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  "span.system" LowCardinality(String) Codec(?CODEC),
  "span.group_id" UInt64 Codec(Delta, ?CODEC),

  "span.trace_id" UUID Codec(?CODEC),
  "span.id" UInt64 Codec(?CODEC),
  "span.parent_id" UInt64 Codec(?CODEC),
  "span.name" LowCardinality(String) Codec(?CODEC),
  "span.event_name" String Codec(?CODEC),
  "span.kind" LowCardinality(String) Codec(?CODEC),
  "span.time" DateTime Codec(Delta, ?CODEC),
  "span.duration" Int64 Codec(Delta, ?CODEC),
  "span.count" Float32 Codec(?CODEC),

  "span.status_code" LowCardinality(String) Codec(?CODEC),
  "span.status_message" String Codec(?CODEC),

  "span.link_count" UInt8 Codec(?CODEC),
  "span.event_count" UInt8 Codec(?CODEC),
  "span.event_error_count" UInt8 Codec(?CODEC),
  "span.event_log_count" UInt8 Codec(?CODEC),

  attr_keys Array(LowCardinality(String)) Codec(?CODEC),
  attr_values Array(String) Codec(?CODEC),

  "service.name" LowCardinality(String) Codec(?CODEC),
  "host.name" LowCardinality(String) Codec(?CODEC),

  "db.system" LowCardinality(String) Codec(?CODEC),
  "db.statement" String Codec(?CODEC),
  "db.operation" LowCardinality(String) Codec(?CODEC),
  "db.sql.table" LowCardinality(String) Codec(?CODEC),

  "log.severity" LowCardinality(String) Codec(?CODEC),
  "log.message" String Codec(?CODEC),

  "exception.type" LowCardinality(String) Codec(?CODEC),
  "exception.message" String Codec(?CODEC),

  INDEX idx_attr_keys attr_keys TYPE bloom_filter(0.01) GRANULARITY 8,
  INDEX idx_duration "span.duration" TYPE minmax GRANULARITY 1
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, "span.system", "span.group_id", "span.time")
PARTITION BY toDate("span.time")
TTL toDate("span.time") + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1

--migration:split

CREATE TABLE spans_data (
  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(?CODEC),
  parent_id UInt64 Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  data String Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (trace_id, id)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE TABLE spans_index_buffer AS spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--migration:split

CREATE TABLE spans_data_buffer AS spans_data
ENGINE = Buffer(currentDatabase(), spans_data, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE span_system_minutes (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_system_minutes_mv
TO span_system_minutes AS
SELECT
  project_id,
  "span.system" AS system,
  toStartOfMinute("span.time") AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32("span.duration"), toUInt32("span.count")) AS tdigest,
  toUInt64(sum("span.count")) AS count,
  countIf("span.status_code" = 'error') AS error_count
FROM spans_index
GROUP BY project_id, time, system
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE span_system_hours (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_system_hours_mv
TO span_system_hours AS
SELECT
  project_id,
  system,
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM span_system_minutes
GROUP BY project_id, toStartOfHour(time), system
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE span_service_minutes (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  service LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, service)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_service_minutes_mv
TO span_service_minutes AS
SELECT
  project_id,
  "span.system" AS system,
  "service.name" AS service,
  toStartOfMinute("span.time") AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32("span.duration"), toUInt32("span.count")) AS tdigest,
  toUInt64(sum("span.count")) AS count,
  countIf("span.status_code" = 'error') AS error_count
FROM spans_index
GROUP BY project_id, time, system, service
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE span_service_hours (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  service LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, service)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_service_hours_mv
TO span_service_hours AS
SELECT
  project_id,
  system,
  service,
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM span_service_minutes
GROUP BY project_id, toStartOfHour(time), system, service
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE span_host_minutes (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  host LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, host)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_host_minutes_mv
TO span_host_minutes AS
SELECT
  project_id,
  "span.system" AS system,
  "host.name" AS host,
  toStartOfMinute("span.time") AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32("span.duration"), toUInt32("span.count")) AS tdigest,
  toUInt64(sum("span.count")) AS count,
  countIf("span.status_code" = 'error') AS error_count
FROM spans_index
GROUP BY project_id, time, system, host
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE span_host_hours (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  host LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, host)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW span_host_hours_mv
TO span_host_hours AS
SELECT
  project_id,
  system,
  host,
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM span_host_minutes
GROUP BY project_id, toStartOfHour(time), system, host
SETTINGS prefer_column_name_to_alias = 1
