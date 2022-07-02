CREATE TABLE spans_data_buffer_dist AS spans_data_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data_buffer)

--migration:split

CREATE TABLE spans_index_dist AS spans_index
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index)

--migration:split

CREATE TABLE span_system_minutes_dist AS span_system_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_minutes)

--migration:split

CREATE TABLE span_system_hours_dist AS span_system_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_hours)

--migration:split

CREATE TABLE span_service_minutes_dist AS span_service_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_minutes)

--migration:split

CREATE TABLE span_service_hours_dist AS span_service_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_hours)

--migration:split

CREATE TABLE span_host_minutes_dist AS span_host_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_minutes)

--migration:split

CREATE TABLE span_host_hours_dist AS span_host_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_hours)
