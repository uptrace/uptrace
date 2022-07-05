CREATE TABLE measure_minutes (
  project_id UInt32 Codec(DoubleDelta, ?CODEC),
  metric LowCardinality(String) Codec(?CODEC),
  time DateTime Codec(DoubleDelta, ?CODEC),
  attrs_hash UInt64 Codec(Delta, ?CODEC),

  instrument LowCardinality(String) Codec(?CODEC),
  sum SimpleAggregateFunction(sumWithOverflow, Float32) Codec(?CODEC),
  value SimpleAggregateFunction(anyLast, Float32) Codec(?CODEC),
  histogram AggregateFunction(quantilesBFloat16(0.5, 0.9, 0.99), Float32) Codec(?CODEC),

  keys Array(LowCardinality(String)) Codec(?CODEC),
  values Array(LowCardinality(String)) Codec(?CODEC)
)
ENGINE = ?(REPLICATED)AggregatingMergeTree
PARTITION BY toDate(time)
ORDER BY (project_id, metric, time, attrs_hash)
TTL toDate(time) + INTERVAL ?TTL DELETE
SETTINGS ttl_only_drop_parts = 1

--migration:split

CREATE TABLE measure_minutes_buffer AS measure_minutes
ENGINE = Buffer(currentDatabase(), measure_minutes, 8, 10, 30, 10000, 1000000, 10000000, 100000000)
