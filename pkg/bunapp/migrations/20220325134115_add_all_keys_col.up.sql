DROP TABLE IF EXISTS ?DB.spans_index_buffer ?ON_CLUSTER

--migration:split

ALTER TABLE ?DB.spans_index ?ON_CLUSTER
ADD COLUMN all_keys Array(LowCardinality(String)) AFTER "span.event_log_count"

--migration:split

CREATE TABLE ?DB.spans_index_buffer ?ON_CLUSTER AS ?DB.spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)
