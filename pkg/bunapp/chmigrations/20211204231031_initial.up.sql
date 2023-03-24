CREATE TABLE ?DB.spans_index ?ON_CLUSTER (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  type LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  group_id UInt64 Codec(Delta, ?CODEC),

  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(T64, ?CODEC),
  parent_id UInt64 Codec(?CODEC),
  name LowCardinality(String) Codec(?CODEC),
  event_name String Codec(?CODEC),
  kind LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  duration Int64 Codec(T64, ?CODEC),
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

  deployment_environment LowCardinality(String) Codec(?CODEC),

  service_name LowCardinality(String) Codec(?CODEC),
  host_name LowCardinality(String) Codec(?CODEC),

  db_system LowCardinality(String) Codec(?CODEC),
  db_statement String Codec(?CODEC),
  db_operation LowCardinality(String) Codec(?CODEC),
  db_sql_table LowCardinality(String) Codec(?CODEC),

  log_severity LowCardinality(String) Codec(?CODEC),
  log_message String Codec(?CODEC),

  exception_type LowCardinality(String) Codec(?CODEC),
  exception_message String Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, system, group_id, time)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE ?DB.spans_data ?ON_CLUSTER (
  project_id UInt32 Codec(Delta, ?CODEC),
  type LowCardinality(String) Codec(?CODEC),
  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(T64, ?CODEC),
  parent_id UInt64 Codec(?CODEC),
  time DateTime Codec(Delta, ?CODEC),
  data String Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (trace_id, id)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 2048,
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
  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, service_name, system)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_system_minutes_mv ?ON_CLUSTER
TO ?DB.span_system_minutes AS
SELECT
  project_id,
  deployment_environment,
  service_name,
  system,

  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,

  toUInt64(sum(count)) AS count,
  toUInt64(sumIf(count, status_code = 'error')) AS error_count
FROM ?DB.spans_index
GROUP BY project_id, toStartOfMinute(time), deployment_environment, service_name, system
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_system_hours ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, service_name, system)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_system_hours_mv ?ON_CLUSTER
TO ?DB.span_system_hours AS
SELECT
  project_id,
  deployment_environment,
  service_name,
  system,

  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,

  sum(count) AS count,
  sum(error_count) AS error_count
FROM ?DB.span_system_minutes
GROUP BY project_id, toStartOfHour(time), deployment_environment, service_name, system
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE ?DB.span_service_minutes ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  deployment_environment LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, system, service_name)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_service_minutes_mv ?ON_CLUSTER
TO ?DB.span_service_minutes AS
SELECT
  project_id,
  deployment_environment,
  system,
  service_name,

  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,

  toUInt64(sum(count)) AS count,
  toUInt64(sumIf(count, status_code = 'error')) AS error_count
FROM ?DB.spans_index
GROUP BY project_id, toStartOfMinute(time), deployment_environment, system, service_name
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_service_hours ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  deployment_environment LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, system, service_name)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_service_hours_mv ?ON_CLUSTER
TO ?DB.span_service_hours AS
SELECT
  project_id,
  deployment_environment,
  system,
  service_name,

  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,

  sum(count) AS count,
  sum(error_count) AS error_count
FROM ?DB.span_service_minutes
GROUP BY project_id, toStartOfHour(time), deployment_environment, system, service_name
SETTINGS prefer_column_name_to_alias = 1

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE ?DB.span_host_minutes ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  host_name LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, service_name, system, host_name)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_host_minutes_mv ?ON_CLUSTER
TO ?DB.span_host_minutes AS
SELECT
  project_id,
  deployment_environment,
  service_name,
  system,
  host_name,

  toStartOfMinute(time) AS time,
  quantilesTDigestWeightedState(0.5, 0.9, 0.99)(toFloat32(duration), toUInt32(count)) AS tdigest,

  toUInt64(sum(count)) AS count,
  toUInt64(sumIf(count, status_code = 'error')) AS error_count
FROM ?DB.spans_index
GROUP BY project_id, toStartOfMinute(time), deployment_environment, service_name, system, host_name
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.span_host_hours ?ON_CLUSTER (
  project_id UInt32 Codec(?CODEC),
  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  host_name LowCardinality(String) Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  tdigest AggregateFunction(quantilesTDigestWeighted(0.5, 0.9, 0.99), Float32, UInt32) Codec(?CODEC),

  count UInt64 Codec(Delta, ?CODEC),
  error_count UInt64 Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (project_id, time, deployment_environment, service_name, system, host_name)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 256

--migration:split

CREATE MATERIALIZED VIEW ?DB.span_host_hours_mv ?ON_CLUSTER
TO ?DB.span_host_hours AS
SELECT
  project_id,
  deployment_environment,
  service_name,
  system,
  host_name,

  toStartOfHour(time) AS time,
  quantilesTDigestWeightedMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,

  sum(count) AS count,
  sum(error_count) AS error_count
FROM ?DB.span_host_minutes
GROUP BY project_id, deployment_environment, service_name, toStartOfHour(time), system, host_name
SETTINGS prefer_column_name_to_alias = 1
