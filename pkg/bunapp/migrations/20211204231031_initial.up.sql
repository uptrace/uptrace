CREATE TABLE spans_index (
  project_id UInt32 Codec(DoubleDelta, Default),
  "span.system" LowCardinality(String),
  "span.group_id" UInt64 Codec(Delta, Default),

  "span.trace_id" UUID,
  "span.id" UInt64,
  "span.parent_id" UInt64,
  "span.name" LowCardinality(String),
  "span.event_name" String,
  "span.kind" LowCardinality(String),
  "span.time" DateTime Codec(Delta, Default),
  "span.duration" Int64 Codec(Delta, Default),
  "span.count" Float32,

  "span.status_code" LowCardinality(String),
  "span.status_message" String,

  "span.link_count" UInt8,
  "span.event_count" UInt8,
  "span.event_error_count" UInt8,
  "span.event_log_count" UInt8,

  all_keys Array(LowCardinality(String)),
  attr_keys Array(LowCardinality(String)),
  attr_values Array(String),

  "service.name" LowCardinality(String),
  "host.name" LowCardinality(String),

  "db.system" LowCardinality(String),
  "db.statement" String,
  "db.operation" LowCardinality(String),
  "db.sql.table" LowCardinality(String),

  "log.severity" LowCardinality(String),
  "log.message" String,

  "exception.type" LowCardinality(String),
  "exception.message" String,

  INDEX idx_attr_keys attr_keys TYPE bloom_filter(0.01) GRANULARITY 64
)
ENGINE = MergeTree()
ORDER BY (project_id, "span.system", "span.group_id", "span.time")
PARTITION BY toDate("span.time")
TTL toDate("span.time") + INTERVAL ?TTL DELETE

--migration:split

CREATE TABLE spans_data (
  trace_id UUID,
  id UInt64,
  parent_id UInt64,
  time DateTime Codec(Delta, Default),
  data String
)
ENGINE = MergeTree()
ORDER BY (trace_id, id)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128

--migration:split

CREATE TABLE spans_index_buffer AS spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--migration:split

CREATE TABLE spans_data_buffer AS spans_data
ENGINE = Buffer(currentDatabase(), spans_data, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE span_system_minutes (
  project_id UInt32,
  system LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
  project_id UInt32,
  system LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
  project_id UInt32,
  system LowCardinality(String),
  service LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, service)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
  project_id UInt32,
  system LowCardinality(String),
  service LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, service)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
  project_id UInt32,
  system LowCardinality(String),
  host LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, host)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
  project_id UInt32,
  system LowCardinality(String),
  host LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, host)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

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
