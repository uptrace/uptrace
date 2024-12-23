CREATE TABLE events_index ?ON_CLUSTER (
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
  string_keys Array(LowCardinality(String)) Codec(?CODEC),
  string_values Array(String) Codec(?CODEC),

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

  process_pid Int32 Codec(T64, ?CODEC),
  process_command LowCardinality(String) Codec(?CODEC),
  process_runtime_name LowCardinality(String) Codec(?CODEC),
  process_runtime_version LowCardinality(String) Codec(?CODEC),
  process_runtime_description LowCardinality(String) Codec(?CODEC),

  messaging_message_id String Codec(?CODEC),
  messaging_message_type LowCardinality(String) Codec(?CODEC),
  messaging_message_payload_size_bytes Int32 Codec(T64, ?CODEC)
)
ENGINE = ?(REPLICATED)MergeTree()
ORDER BY (project_id, system, group_id, time)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
SETTINGS ttl_only_drop_parts = 1,
         storage_policy = ?SPANS_STORAGE

--migration:split

CREATE TABLE events_data ?ON_CLUSTER (
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
