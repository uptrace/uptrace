DROP TABLE IF EXISTS ?DB.spans_index ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_data ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_tracing_events_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_tracing_spans_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_event_count_mv ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP TABLE IF EXISTS ?DB.datapoint_minutes ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.datapoint_hours ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.datapoint_hours_mv ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP TABLE IF EXISTS ?DB.service_graph_edges ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_service_graph_client_duration_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_service_graph_server_duration_mv ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.metrics_uptrace_service_graph_failed_requests_mv ?ON_CLUSTER
