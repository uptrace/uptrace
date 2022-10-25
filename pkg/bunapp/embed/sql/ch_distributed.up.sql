CREATE TABLE ?DB.spans_data_buffer_dist ?ON_CLUSTER AS ?DB.spans_data_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data_buffer, rand())

--migration:split

CREATE TABLE ?DB.spans_index_buffer_dist ?ON_CLUSTER AS ?DB.spans_index_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index_buffer, rand())

--migration:split

CREATE TABLE ?DB.span_system_minutes_dist ?ON_CLUSTER AS ?DB.span_system_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_minutes, rand())

--migration:split

CREATE TABLE ?DB.span_system_hours_dist ?ON_CLUSTER AS ?DB.span_system_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_system_hours, rand())

--migration:split

CREATE TABLE ?DB.span_service_minutes_dist ?ON_CLUSTER AS ?DB.span_service_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_minutes, rand())

--migration:split

CREATE TABLE ?DB.span_service_hours_dist ?ON_CLUSTER AS ?DB.span_service_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_service_hours, rand())

--migration:split

CREATE TABLE ?DB.span_host_minutes_dist ?ON_CLUSTER AS ?DB.span_host_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_minutes, rand())

--migration:split

CREATE TABLE ?DB.span_host_hours_dist ?ON_CLUSTER AS ?DB.span_host_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), span_host_hours, rand())

--migration:split

CREATE TABLE ?DB.measure_minutes_buffer_dist ?ON_CLUSTER AS ?DB.measure_minutes_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_minutes_buffer, rand())

--migration:split

CREATE TABLE ?DB.measure_hours_dist ?ON_CLUSTER AS ?DB.measure_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_hours, rand())
