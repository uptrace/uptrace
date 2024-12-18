CREATE TABLE spans_index ?ON_CLUSTER (
  id UInt64 Codec(T64, ?CODEC),
  trace_id UUID Codec(?CODEC),
  parent_id UInt64 Codec(?CODEC),

  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  type LowCardinality(String) Codec(?CODEC),
  system LowCardinality(String) Codec(?CODEC),
  group_id UInt64 Codec(Delta, ?CODEC),

  kind LowCardinality(String) Codec(?CODEC),
  name LowCardinality(String) Codec(?CODEC),
  event_name String Codec(?CODEC),
  display_name String Codec(?CODEC),

  time DateTime Codec(Delta, ?CODEC),
  duration Int64 Codec(T64, ?CODEC),
  count Float32 Codec(?CODEC),

  status_code LowCardinality(String) Codec(?CODEC),
  status_message String Codec(?CODEC),

  all_keys Array(LowCardinality(String)) Codec(?CODEC),
  string_keys Array(LowCardinality(String)) Codec(?CODEC),
  string_values Array(String) Codec(?CODEC),

  telemetry_sdk_name LowCardinality(String) Codec(?CODEC),
  telemetry_sdk_language LowCardinality(String) Codec(?CODEC),
  telemetry_sdk_version LowCardinality(String) Codec(?CODEC),
  telemetry_auto_version LowCardinality(String) Codec(?CODEC),

  otel_library_name LowCardinality(String) Codec(?CODEC),
  otel_library_version LowCardinality(String) Codec(?CODEC),

  deployment_environment LowCardinality(String) Codec(?CODEC),

  service_name LowCardinality(String) Codec(?CODEC),
  service_version LowCardinality(String) Codec(?CODEC),
  service_namespace LowCardinality(String) Codec(?CODEC),
  host_name LowCardinality(String) Codec(?CODEC),

  client_address LowCardinality(String) Codec(?CODEC),
  client_socket_address LowCardinality(String) Codec(?CODEC),
  client_socket_port Int32 Codec(?CODEC),

  url_scheme LowCardinality(String) Codec(?CODEC),
  url_full String Codec(?CODEC),
  url_path LowCardinality(String) Codec(?CODEC),

  http_request_method LowCardinality(String) Codec(?CODEC),
  http_response_status_code UInt16 Codec(?CODEC),
  http_route LowCardinality(String) Codec(?CODEC),

  rpc_method LowCardinality(String) Codec(?CODEC),
  rpc_service LowCardinality(String) Codec(?CODEC),

  db_system LowCardinality(String) Codec(?CODEC),
  db_name LowCardinality(String) Codec(?CODEC),
  db_statement String Codec(?CODEC),
  db_operation LowCardinality(String) Codec(?CODEC),
  db_sql_table LowCardinality(String) Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, system, group_id, time)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE spans_data ?ON_CLUSTER (
  project_id UInt32 Codec(Delta, ?CODEC),
  type LowCardinality(String) Codec(?CODEC),
  trace_id UUID Codec(?CODEC),
  id UInt64 Codec(T64, ?CODEC),
  parent_id UInt64 Codec(?CODEC),
  time DateTime64(6) Codec(Delta, ?CODEC),
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

DROP TABLE IF EXISTS spans_index_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE datapoint_minutes ?ON_CLUSTER (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  metric LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(DoubleDelta, ?CODEC),
  attrs_hash UInt64 Codec(Delta, ?CODEC),

  instrument LowCardinality(String) Codec(?CODEC),
  min SimpleAggregateFunction(min, Float64) Codec(?CODEC),
  max SimpleAggregateFunction(max, Float64) Codec(?CODEC),
  sum SimpleAggregateFunction(sum, Float64) Codec(?CODEC),
  count SimpleAggregateFunction(sum, UInt64) Codec(T64, ?CODEC),
  gauge SimpleAggregateFunction(anyLast, Float64) Codec(?CODEC),
  histogram AggregateFunction(quantilesBFloat16(0.5), Float32) Codec(?CODEC),

  string_keys Array(LowCardinality(String)) Codec(?CODEC),
  string_values Array(String) Codec(?CODEC),
  annotations SimpleAggregateFunction(max, String) Codec(?CODEC)
)
ENGINE = ?(REPLICATED)AggregatingMergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, metric, time, attrs_hash)
TTL toDate(time) + INTERVAL ?METRICS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?METRICS_STORAGE

--migration:split

DROP TABLE IF EXISTS datapoint_minutes_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE datapoint_hours ?ON_CLUSTER (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  metric LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(DoubleDelta, ?CODEC),
  attrs_hash UInt64 Codec(Delta, ?CODEC),

  instrument LowCardinality(String) Codec(?CODEC),
  min SimpleAggregateFunction(min, Float64) Codec(?CODEC),
  max SimpleAggregateFunction(max, Float64) Codec(?CODEC),
  sum SimpleAggregateFunction(sum, Float64) Codec(?CODEC),
  count SimpleAggregateFunction(sum, UInt64) Codec(T64, ?CODEC),
  gauge SimpleAggregateFunction(anyLast, Float64) Codec(?CODEC),
  histogram AggregateFunction(quantilesBFloat16(0.5), Float32) Codec(?CODEC),

  string_keys Array(LowCardinality(String)) Codec(?CODEC),
  string_values Array(String) Codec(?CODEC),
  annotations SimpleAggregateFunction(max, String) Codec(?CODEC)
)
ENGINE = ?(REPLICATED)AggregatingMergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, metric, time, attrs_hash)
TTL toDate(time) + INTERVAL ?METRICS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?METRICS_STORAGE

--migration:split

CREATE MATERIALIZED VIEW datapoint_hours_mv ?ON_CLUSTER
TO datapoint_hours
AS SELECT
  project_id,
  metric,
  toStartOfHour(time) AS time,
  attrs_hash,

  anyLast(instrument) AS instrument,
  min(min) AS min,
  max(max) AS max,
  sum(sum) AS sum,
  sum(count) AS count,
  anyLast(gauge) AS gauge,
  quantilesBFloat16MergeState(0.5)(histogram) AS histogram,

  any(string_keys) AS string_keys,
  any(string_values) AS string_values,
  max(annotations) AS annotations
FROM datapoint_minutes AS m
GROUP BY m.project_id, m.metric, toStartOfHour(m.time), m.attrs_hash

--------------------------------------------------------------------------------
--migration:split

CREATE TABLE service_graph_edges ?ON_CLUSTER (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  type LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(T64, ?CODEC),

  client_name LowCardinality(String) Codec(?CODEC),
  server_name LowCardinality(String) Codec(?CODEC),
  server_attr LowCardinality(String) Codec(?CODEC),

  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_namespace LowCardinality(String) Codec(?CODEC),

  client_duration_min SimpleAggregateFunction(min, Float32) Codec(?CODEC),
  client_duration_max SimpleAggregateFunction(max, Float32) Codec(?CODEC),
  client_duration_sum SimpleAggregateFunction(sumWithOverflow, Float32) Codec(?CODEC),

  server_duration_min SimpleAggregateFunction(min, Float32) Codec(?CODEC),
  server_duration_max SimpleAggregateFunction(max, Float32) Codec(?CODEC),
  server_duration_sum SimpleAggregateFunction(sumWithOverflow, Float32) Codec(?CODEC),

  count SimpleAggregateFunction(sumWithOverflow, UInt32) Codec(Delta, ?CODEC),
  error_count SimpleAggregateFunction(sumWithOverflow, UInt32) Codec(Delta, ?CODEC)
)
ENGINE = ?(REPLICATED)AggregatingMergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, time, type, client_name, server_name, deployment_environment, service_namespace)
PRIMARY KEY (project_id, time, type, client_name, server_name)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
  storage_policy = ?SPANS_STORAGE

--migration:split

DROP TABLE IF EXISTS service_graph_edges_buffer ?ON_CLUSTER

--migration:split

CREATE MATERIALIZED VIEW metrics_uptrace_service_graph_client_duration_mv ?ON_CLUSTER
TO datapoint_minutes AS
SELECT
  e.project_id,
  'uptrace_service_graph_client_duration' AS metric,
  e.time,
  xxHash64(
    e.project_id,
    e.type,
    e.client_name,
    e.server_name,
    e.deployment_environment,
    e.service_namespace
  ) AS attrs_hash,

  'summary' AS instrument,
  min(e.client_duration_min) AS min,
  min(e.client_duration_max) AS max,
  sum(e.client_duration_sum) AS sum,
  sum(e.count) AS count,

  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
  ) AS all_keys,
  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
   ) AS string_keys,
  arrayConcat(
    [e.type, e.client_name, e.server_name],
    if(e.deployment_environment != '', [e.deployment_environment], []),
    if(e.service_namespace != '', [e.service_namespace], [])
  ) AS string_values
FROM service_graph_edges AS e
WHERE e.count > 0 AND e.client_duration_sum > 0
GROUP BY
  e.project_id,
  e.type,
  e.time,
  e.client_name,
  e.server_name,
  e.deployment_environment,
  e.service_namespace

--migration:split

CREATE MATERIALIZED VIEW metrics_uptrace_service_graph_server_duration_mv ?ON_CLUSTER
TO datapoint_minutes AS
SELECT
  e.project_id,
  'uptrace_service_graph_server_duration' AS metric,
  e.time,
  xxHash64(
    e.project_id,
    e.type,
    e.client_name,
    e.server_name,
    e.deployment_environment,
    e.service_namespace
  ) AS attrs_hash,

  'summary' AS instrument,
  min(e.server_duration_min) AS min,
  min(e.server_duration_max) AS max,
  sum(e.server_duration_sum) AS sum,
  sum(e.count) AS count,

  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
  ) AS all_keys,
  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
   ) AS string_keys,
  arrayConcat(
    [e.type, e.client_name, e.server_name],
    if(e.deployment_environment != '', [e.deployment_environment], []),
    if(e.service_namespace != '', [e.service_namespace], [])
  ) AS string_values
FROM service_graph_edges AS e
WHERE e.count > 0 AND e.server_duration_sum > 0
GROUP BY
  e.project_id,
  e.type,
  e.time,
  e.client_name,
  e.server_name,
  e.deployment_environment,
  e.service_namespace

--migration:split

CREATE MATERIALIZED VIEW metrics_uptrace_service_graph_failed_requests_mv ?ON_CLUSTER
TO datapoint_minutes AS
SELECT
  e.project_id,
  'uptrace_service_graph_failed_requests' AS metric,
  e.time,
  xxHash64(
    e.project_id,
    e.type,
    e.client_name,
    e.server_name,
    e.deployment_environment,
    e.service_namespace
  ) AS attrs_hash,

  'counter' AS instrument,
  sum(e.error_count) AS sum,

  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
  ) AS all_keys,
  arrayConcat(
    ['type', 'client', 'server'],
    if(e.deployment_environment != '', ['deployment_environment'], []),
    if(e.service_namespace != '', ['service_namespace'], [])
   ) AS string_keys,
  arrayConcat(
    [e.type, e.client_name, e.server_name],
    if(e.deployment_environment != '', [e.deployment_environment], []),
    if(e.service_namespace != '', [e.service_namespace], [])
  ) AS string_values
FROM service_graph_edges AS e
WHERE e.error_count > 0
GROUP BY
  e.project_id,
  e.type,
  e.time,
  e.client_name,
  e.server_name,
  e.deployment_environment,
  e.service_namespace
