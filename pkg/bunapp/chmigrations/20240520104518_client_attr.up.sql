alter table service_graph_edges ?ON_CLUSTER
add column client_attr LowCardinality(String) Codec(?CODEC) after time

--migration:split

DROP TABLE IF EXISTS service_graph_edges_buffer ?ON_CLUSTER
