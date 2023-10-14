DROP TABLE IF EXISTS spans_data_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS spans_data_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS spans_index_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS spans_index_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS datapoint_minutes_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS datapoint_minutes_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS datapoint_minutes_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS datapoint_hours_dist ?ON_CLUSTER
