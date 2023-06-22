ALTER TABLE ?DB.spans_data ?ON_CLUSTER
ADD INDEX trace_id_idx (trace_id) TYPE bloom_filter GRANULARITY 32;
