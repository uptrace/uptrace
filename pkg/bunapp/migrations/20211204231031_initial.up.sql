CREATE TABLE spans_index (
  "span.system" LowCardinality(String),
  "span.group_id" UInt64 Codec(Delta, Default),

  "span.id" UInt64,
  "span.trace_id" UUID,
  "span.name" LowCardinality(String),
  "span.kind" LowCardinality(String),
  "span.time" DateTime Codec(Delta, Default),
  "span.duration" Int64,

  "span.status_code" LowCardinality(String),
  "span.status_message" String,

  "span.event_count" UInt8,
  "span.event_error_count" UInt8,
  "span.event_log_count" UInt8,

  attr_keys Array(LowCardinality(String)),
  attr_values Array(String),

  "service.name" LowCardinality(String),
  "host.name" LowCardinality(String),

  INDEX idx_attr_keys attr_keys TYPE bloom_filter(0.01) GRANULARITY 64
)
ENGINE = MergeTree()
ORDER BY ("span.system", "span.group_id", "span.time")
PARTITION BY toDate("span.time")
TTL toDate("span.time") + INTERVAL ?TTL DELETE;

--migrate:split

CREATE TABLE spans_data (
  id UInt64,
  parent_id UInt64,
  trace_id UUID,
  time DateTime,
  attrs String,
  events String,
  links String
)
ENGINE = MergeTree()
ORDER BY (trace_id)
PARTITION BY toDate(time)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 64;

--migrate:split

CREATE TABLE spans_index_buffer AS spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--migrate:split

CREATE TABLE spans_data_buffer AS spans_data
ENGINE = Buffer(currentDatabase(), spans_data, 5, 10, 30, 10000, 1000000, 10000000, 100000000)

--migrate:split

CREATE TABLE span_system_minutes (
  system LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigest(0.5, 0.9, 0.99), Int64),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

--migrate:split

CREATE MATERIALIZED VIEW span_system_minutes_mv
TO span_system_minutes AS
SELECT
  "span.system" AS system,
  toStartOfMinute("span.time") AS time,
  quantilesTDigestState(0.5, 0.9, 0.99)("span.duration") AS tdigest,
  count() AS count,
  countIf("span.status_code" = 'error') AS error_count
FROM spans_index AS s
GROUP BY time, system

--migrate:split

CREATE TABLE span_system_hours (
  system LowCardinality(String),
  time DateTime Codec(Delta, Default),
  tdigest AggregateFunction(quantilesTDigest(0.5, 0.9, 0.99), Int64),
  count UInt64 Codec(Delta, Default),
  error_count UInt64 Codec(Delta, Default)
)
ENGINE = SummingMergeTree()
PARTITION BY toDate(time)
ORDER BY (time, system)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS index_granularity = 128;

--migrate:split

CREATE MATERIALIZED VIEW span_system_hours_mv
TO span_system_hours AS
SELECT
  system,
  toStartOfHour(time) AS time,
  quantilesTDigestMergeState(0.5, 0.9, 0.99)(tdigest) AS tdigest,
  sum(count) AS count,
  sum(error_count) AS error_count
FROM span_system_minutes AS s
GROUP BY time, system
