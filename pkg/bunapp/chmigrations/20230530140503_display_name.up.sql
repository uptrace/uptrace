DROP TABLE ?DB.spans_index_buffer ?ON_CLUSTER;

--migration:split

ALTER TABLE ?DB.spans_index ?ON_CLUSTER
ADD COLUMN display_name String Codec(?CODEC)
AFTER event_name;

--migration:split

ALTER TABLE ?DB.spans_index ?ON_CLUSTER
UPDATE display_name = if(event_name != '', event_name, name) WHERE display_name = '';

--migration:split

CREATE TABLE ?DB.spans_index_buffer ?ON_CLUSTER AS ?DB.spans_index
ENGINE = Buffer(currentDatabase(), spans_index, 3, 5, 10, 10000, 1000000, 10000000, 100000000);
