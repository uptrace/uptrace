DROP TABLE IF EXISTS ?DB.spans_index ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_data ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_index_buffer ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_data_buffer ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_tracing_events_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_tracing_spans_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_event_count_mv ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP TABLE IF EXISTS ?DB.datapoint_minutes_buffer ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.datapoint_minutes ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.datapoint_hours ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.datapoint_hours_mv ?ON_CLUSTER
