CREATE TABLE logs_index ?ON_CLUSTER  (
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

  link_count UInt8 Codec(?CODEC),
  event_count UInt8 Codec(?CODEC),
  event_error_count UInt8 Codec(?CODEC),
  event_log_count UInt8 Codec(?CODEC),

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
  db_sql_table LowCardinality(String) Codec(?CODEC),

  log_severity LowCardinality(String) Codec(?CODEC),
  log_message String Codec(?CODEC),

  exception_type LowCardinality(String) Codec(?CODEC),
  exception_message String Codec(?CODEC)
)
ENGINE = MergeTree()
ORDER BY (project_id, trace_id, time)
PARTITION BY toDate(time)
TTL time + INTERVAL ?LOGS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?LOGS_STORAGE;

--migration:split

CREATE TABLE logs_data  (
  project_id UInt32 Codec(Delta, ZSTD(1)),
  type LowCardinality(String) Codec(ZSTD(1)),
  trace_id UUID Codec(ZSTD(1)),
  id UInt64 Codec(T64, ZSTD(1)),
  parent_id UInt64 Codec(ZSTD(1)),
  time DateTime Codec(Delta, ZSTD(1)),
  data String Codec(ZSTD(1))
)
ENGINE = MergeTree()
ORDER BY (project_id, trace_id, time)
PARTITION BY toDate(time)
TTL time + INTERVAL ?LOGS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         index_granularity = 2048,
         storage_policy = ?LOGS_STORAGE;

--migration:split

DROP TABLE IF EXISTS logs_index_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE logs_index_buffer ?ON_CLUSTER AS logs_index
ENGINE = Buffer(currentDatabase(), logs_index,
  5,
  5, 45,
  1_000_000, 1_000_000,
  500_000_000, 500_000_000)

--migration:split

DROP TABLE IF EXISTS logs_data_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE logs_data_buffer ?ON_CLUSTER AS logs_data
ENGINE = Buffer(currentDatabase(), logs_data,
  3,
  5, 45,
  1_000_000, 1_000_000,
  500_000_000, 500_000_000)

--migration:split
