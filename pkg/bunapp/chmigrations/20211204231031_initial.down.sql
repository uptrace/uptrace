DROP VIEW IF EXISTS ?DB.span_host_minutes_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_host_minutes ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.span_host_hours_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_host_hours ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS ?DB.span_service_minutes_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_service_minutes ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.span_service_hours_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_service_hours ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS ?DB.span_system_minutes_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_system_minutes ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.span_system_hours_mv ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.span_system_hours ?ON_CLUSTER

--------------------------------------------------------------------------------
--migration:split

DROP TABLE IF EXISTS ?DB.spans_index ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_data ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_index_buffer ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.spans_data_buffer ?ON_CLUSTER
