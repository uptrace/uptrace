DROP TABLE IF EXISTS spans_index_buffer

--migration:split

ALTER TABLE spans_index
ADD COLUMN "span.is_event" UInt8 ALIAS "span.event_name" != ''

--migration:split

CREATE TABLE spans_index_buffer AS spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 5, 10, 30, 10000, 1000000, 10000000, 100000000)
