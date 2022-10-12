CREATE TABLE ?DB.spans_index ?ON_CLUSTER (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  group_id UInt64 Codec(Delta, ?CODEC),

  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(?CODEC),
  parent_id UInt64 Codec(?CODEC),
  name LowCardinality(String) Codec(?CODEC),
  event_name String Codec(?CODEC),
  is_event UInt8 ALIAS event_name != '',
  kind LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  duration Int64 Codec(Delta, ?CODEC),
  count Float32 Codec(?CODEC),

  status_code LowCardinality(String) Codec(?CODEC),
  status_message String Codec(?CODEC),

  link_count UInt8 Codec(?CODEC),
  event_count UInt8 Codec(?CODEC),
  event_error_count UInt8 Codec(?CODEC),
  event_log_count UInt8 Codec(?CODEC),

  all_keys Array(LowCardinality(String)) Codec(?CODEC),
  attr_keys Array(LowCardinality(String)) Codec(?CODEC),
  attr_values Array(String) Codec(?CODEC),

  "deployment.environment" LowCardinality(String) Codec(?CODEC),

  service LowCardinality(String) Codec(?CODEC),
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

  INDEX idx_attr_keys attr_keys TYPE bloom_filter(0.01) GRANULARITY 64,
  INDEX idx_duration duration TYPE minmax GRANULARITY 1
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, system, group_id, time, "deployment.environment",)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE ?DB.spans_data ?ON_CLUSTER (
  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(?CODEC),
  parent_id UInt64 Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  data String Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (trace_id, id)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE ?DB.spans_index_buffer ?ON_CLUSTER AS ?DB.spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--migration:split

CREATE TABLE ?DB.spans_data_buffer ?ON_CLUSTER AS ?DB.spans_data
ENGINE = Buffer(currentDatabase(), spans_data, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE ?DB.span_system_minutes ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  "deployment.environment" LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, system, "deployment.environment")
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_system_minutes_mv ?ON_CLUSTER
TO ?DB.span_system_minutes AS
SELECT
  project_id,
  system,
  "deployment.environment",
  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,
  toUInt64(sum(count)) AS count,
  countIf(status_code = 'error') AS error_count
FROM ?DB.spans_index
GROUP BY project_id, "deployment.environment", toStartOfMinute(time), system
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_system_hours ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  "deployment.environment" LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),
  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, "deployment.environment", time, system)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_system_hours_mv ?ON_CLUSTER
TO ?DB.span_system_hours AS
SELECT
  project_id,
  system,
  "deployment.environment",
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM span_system_minutes
GROUP BY project_id, "deployment.environment", toStartOfHour(time), system
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE ?DB.span_service_minutes ?ON_CLUSTER (
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
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_service_minutes_mv ?ON_CLUSTER
TO ?DB.span_service_minutes AS
SELECT
  project_id,
  system,
  service,
  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,
  toUInt64(sum(count)) AS count,
  countIf(status_code = 'error') AS error_count
FROM ?DB.spans_index
GROUP BY project_id, toStartOfMinute(time), system, service
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_service_hours ?ON_CLUSTER (
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
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_service_hours_mv ?ON_CLUSTER
TO ?DB.span_service_hours AS
SELECT
  project_id,
  system,
  service,
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM ?DB.span_service_minutes
GROUP BY project_id, toStartOfHour(time), system, service
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE ?DB.span_host_minutes ?ON_CLUSTER (
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
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_host_minutes_mv ?ON_CLUSTER
TO ?DB.span_host_minutes AS
SELECT
  project_id,
  system,
  "host.name" AS host,
  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,
  toUInt64(sum(count)) AS count,
  countIf(status_code = 'error') AS error_count
FROM ?DB.spans_index
GROUP BY project_id, toStartOfMinute(time), system, host
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_host_hours ?ON_CLUSTER (
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
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 128

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_host_hours_mv ?ON_CLUSTER
TO ?DB.span_host_hours AS
SELECT
  project_id,
  system,
  host,
  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM ?DB.span_host_minutes
GROUP BY project_id, toStartOfHour(time), system, host
SETTINGS prefer_column_name_to_alias = 1
