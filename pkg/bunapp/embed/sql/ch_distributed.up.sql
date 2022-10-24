CREATE TABLE spans_data_buffer_dist ?ON_CLUSTER AS spans_data_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data_buffer)

--migration:split

CREATE TABLE spans_index_buffer_dist ?ON_CLUSTER AS spans_index_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index_buffer)

--migration:split

CREATE TABLE span_system_minutes_dist ?ON_CLUSTER AS span_system_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_minutes)

--migration:split

CREATE TABLE span_system_hours_dist ?ON_CLUSTER AS span_system_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_hours)

--migration:split

CREATE TABLE span_service_minutes_dist ?ON_CLUSTER AS span_service_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_minutes)

--migration:split

CREATE TABLE span_service_hours_dist ?ON_CLUSTER AS span_service_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_hours)

--migration:split

CREATE TABLE span_host_minutes_dist ?ON_CLUSTER AS span_host_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_minutes)

--migration:split

CREATE TABLE span_host_hours_dist ?ON_CLUSTER AS span_host_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_hours)

--migration:split

CREATE TABLE measure_minutes_buffer_dist ?ON_CLUSTER AS measure_minutes_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_minutes_buffer)

--migration:split

CREATE TABLE measure_hours_dist ?ON_CLUSTER AS measure_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_hours)
