DROP VIEW IF EXISTS span_host_minutes_mv;

--migrate:split

DROP TABLE IF EXISTS span_host_minutes;

--migrate:split

DROP VIEW IF EXISTS span_host_hours_mv;

--migrate:split

DROP TABLE IF EXISTS span_host_hours;

--------------------------------------------------------------------------------
--migrate:split

DROP VIEW IF EXISTS span_service_minutes_mv;

--migrate:split

DROP TABLE IF EXISTS span_service_minutes;

--migrate:split

DROP VIEW IF EXISTS span_service_hours_mv;

--migrate:split

DROP TABLE IF EXISTS span_service_hours;

--------------------------------------------------------------------------------
--migrate:split

DROP VIEW IF EXISTS span_system_minutes_mv;

--migrate:split

DROP TABLE IF EXISTS span_system_minutes;

--migrate:split

DROP VIEW IF EXISTS span_system_hours_mv;

--migrate:split

DROP TABLE IF EXISTS span_system_hours;

--------------------------------------------------------------------------------
--migrate:split

DROP TABLE IF EXISTS spans_index;

--migrate:split

DROP TABLE IF EXISTS spans_data;

--migrate:split

DROP TABLE IF EXISTS spans_index_buffer;

--migrate:split

DROP TABLE IF EXISTS spans_data_buffer;
