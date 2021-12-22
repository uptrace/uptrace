DROP VIEW IF EXISTS span_system_minutes_mv;

--migrate:split

DROP TABLE IF EXISTS span_system_minutes;

--migrate:split

DROP VIEW IF EXISTS span_system_hours_mv;

--migrate:split

DROP TABLE IF EXISTS span_system_hours;

--migrate:split

DROP TABLE IF EXISTS spans_index;

--migrate:split

DROP TABLE IF EXISTS spans_data;

--migrate:split

DROP TABLE IF EXISTS spans_index_buffer;

--migrate:split

DROP TABLE IF EXISTS spans_data_buffer;
