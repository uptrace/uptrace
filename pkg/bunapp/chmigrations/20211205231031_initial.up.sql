CREATE TABLE ?DB.spans_index ?ON_CLUSTER (
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
  time DateTime64(9) Codec(Delta, ?CODEC),
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
