CREATE TABLE ?DB.spans_data_dist ?ON_CLUSTER AS ?DB.spans_data
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data, rand())

--migration:split

CREATE TABLE ?DB.spans_data_buffer_dist ?ON_CLUSTER AS ?DB.spans_data_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_data_buffer, rand())

--migration:split

CREATE TABLE ?DB.spans_index_dist ?ON_CLUSTER AS ?DB.spans_index
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index, rand())

--migration:split

CREATE TABLE ?DB.spans_index_buffer_dist ?ON_CLUSTER AS ?DB.spans_index_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index_buffer, rand())

--migration:split

CREATE TABLE ?DB.datapoint_minutes_dist ?ON_CLUSTER AS ?DB.datapoint_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), datapoint_minutes, rand())

--migration:split

CREATE TABLE ?DB.datapoint_minutes_buffer_dist ?ON_CLUSTER AS ?DB.datapoint_minutes_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), datapoint_minutes_buffer, rand())

--migration:split

CREATE TABLE ?DB.datapoint_hours_dist ?ON_CLUSTER AS ?DB.datapoint_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), datapoint_hours, rand())
