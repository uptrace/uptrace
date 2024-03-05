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

CREATE OR REPLACE TABLE datapoint_minutes_buffer ?ON_CLUSTER AS datapoint_minutes
ENGINE = Buffer(currentDatabase(), datapoint_minutes, 3, 5, 10, 10000, 1000000, 10000000, 100000000)
