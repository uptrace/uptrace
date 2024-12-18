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
