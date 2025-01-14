CREATE TABLE logs_index ?ON_CLUSTER (
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
  count Float32 Codec(?CODEC),

  all_keys Array(LowCardinality(String)) Codec(?CODEC),
  attrs JSON Codec(?CODEC),

  deployment_environment LowCardinality(String) Codec(?CODEC),
  service_namespace LowCardinality(String) Codec(?CODEC),
  service_name LowCardinality(String) Codec(?CODEC),
  service_version LowCardinality(String) Codec(?CODEC),
  host_name LowCardinality(String) Codec(?CODEC),

  telemetry_sdk_name LowCardinality(String) Codec(?CODEC),
  telemetry_sdk_language LowCardinality(String) Codec(?CODEC),
  telemetry_sdk_version LowCardinality(String) Codec(?CODEC),
  telemetry_auto_version LowCardinality(String) Codec(?CODEC),

  otel_library_name LowCardinality(String) Codec(?CODEC),
  otel_library_version LowCardinality(String) Codec(?CODEC),

  log_severity Enum8('' = 0, 'TRACE' = 1, 'TRACE2' = 2, 'TRACE3' = 3, 'TRACE4' = 4, 'DEBUG' = 5, 'DEBUG2' = 6, 'DEBUG3' = 7, 'DEBUG4' = 8, 'INFO' = 9, 'INFO2' = 10, 'INFO3' = 11, 'INFO4' = 12, 'WARN' = 13, 'WARN2' = 14, 'WARN3' = 15, 'WARN4' = 16, 'ERROR' = 17, 'ERROR2' = 18, 'ERROR3' = 19, 'ERROR4' = 20, 'FATAL' = 21, 'FATAL2' = 22, 'FATAL3' = 23, 'FATAL4' = 24) Codec(?CODEC),
  log_file_path LowCardinality(String) Codec(?CODEC),
  log_file_name LowCardinality(String) Codec(?CODEC),
  log_iostream LowCardinality(String) Codec(?CODEC),
  log_source LowCardinality(String) Codec(?CODEC),

  exception_type LowCardinality(String) Codec(?CODEC),
  exception_stacktrace String Codec(?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, system, group_id, time)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE logs_data ?ON_CLUSTER (
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
