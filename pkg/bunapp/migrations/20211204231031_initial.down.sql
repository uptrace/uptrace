DROP VIEW IF EXISTS span_host_minutes_mv;

--migration:split

DROP TABLE IF EXISTS span_host_minutes;

--migration:split

DROP VIEW IF EXISTS span_host_hours_mv;

--migration:split

DROP TABLE IF EXISTS span_host_hours;

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS span_service_minutes_mv;

--migration:split

DROP TABLE IF EXISTS span_service_minutes;

--migration:split

DROP VIEW IF EXISTS span_service_hours_mv;

--migration:split

DROP TABLE IF EXISTS span_service_hours;

--------------------------------------------------------------------------------
--migration:split

DROP VIEW IF EXISTS span_system_minutes_mv;

--migration:split

DROP TABLE IF EXISTS span_system_minutes;

--migration:split

DROP VIEW IF EXISTS span_system_hours_mv;

--migration:split

DROP TABLE IF EXISTS span_system_hours;

--------------------------------------------------------------------------------
--migration:split

DROP TABLE IF EXISTS spans_index;

--migration:split

DROP TABLE IF EXISTS spans_data;

--migration:split

DROP TABLE IF EXISTS spans_index_buffer;

--migration:split

DROP TABLE IF EXISTS spans_data_buffer;
