alter table datapoint_minutes
add column otel_library_name LowCardinality(String) Codec(?CODEC)

--migration:split

alter table datapoint_minutes
add column otel_library_version LowCardinality(String) Codec(?CODEC)

--migration:split

alter table datapoint_hours
add column otel_library_name LowCardinality(String) Codec(?CODEC)

--migration:split

alter table datapoint_hours
add column otel_library_version LowCardinality(String) Codec(?CODEC)

--migration:split

DROP TABLE IF EXISTS datapoint_minutes_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE datapoint_minutes_buffer ?ON_CLUSTER AS datapoint_minutes
ENGINE = Buffer(currentDatabase(), datapoint_minutes,
  5,
  5, 45,
  1_000_000, 1_000_000,
  500_000_000, 500_000_000)
