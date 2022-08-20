DROP TABLE IF EXISTS ?DB.spans_index_buffer ?ON_CLUSTER

--migration:split

ALTER TABLE ?DB.spans_index ?ON_CLUSTER
ADD COLUMN "span.is_event" UInt8 ALIAS "span.event_name" != ''

--migration:split

CREATE TABLE ?DB.spans_index_buffer ?ON_CLUSTER AS ?DB.spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)
