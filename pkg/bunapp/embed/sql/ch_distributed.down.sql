DROP TABLE IF EXISTS spans_data_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS spans_index_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_system_minutes_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_system_hours_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_service_minutes_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_service_hours_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_host_minutes_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS span_host_hours_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS measure_minutes_dist ?ON_CLUSTER
