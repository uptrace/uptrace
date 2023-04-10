CREATE TABLE ?DB.spans_data_buffer_dist ?ON_CLUSTER AS ?DB.spans_data_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data_buffer, rand())

--migration:split

CREATE TABLE ?DB.spans_index_buffer_dist ?ON_CLUSTER AS ?DB.spans_index_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index_buffer, rand())

--migration:split

CREATE TABLE ?DB.measure_minutes_buffer_dist ?ON_CLUSTER AS ?DB.measure_minutes_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_minutes_buffer, rand())

--migration:split

CREATE TABLE ?DB.measure_hours_dist ?ON_CLUSTER AS ?DB.measure_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_hours, rand())
