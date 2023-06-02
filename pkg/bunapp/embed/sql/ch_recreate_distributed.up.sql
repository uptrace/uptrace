DROP TABLE ?DB.spans_index_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE ?DB.measure_minutes_dist ?ON_CLUSTER

--migration:split

DROP TABLE ?DB.measure_minutes_buffer_dist ?ON_CLUSTER

--migration:split

DROP TABLE ?DB.measure_hours_dist ?ON_CLUSTER

--migration:split

CREATE TABLE ?DB.spans_index_buffer_dist ?ON_CLUSTER AS ?DB.spans_index_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), spans_index_buffer, rand())

--migration:split

CREATE TABLE ?DB.measure_minutes_dist ?ON_CLUSTER AS ?DB.measure_minutes
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_minutes, rand())

--migration:split

CREATE TABLE ?DB.measure_minutes_buffer_dist ?ON_CLUSTER AS ?DB.measure_minutes_buffer
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_minutes_buffer, rand())

--migration:split

CREATE TABLE ?DB.measure_hours_dist ?ON_CLUSTER AS ?DB.measure_hours
ENGINE = Distributed(?CLUSTER, currentDatabase(), measure_hours, rand())
