alter table service_graph_edges ?ON_CLUSTER
add column client_attr LowCardinality(String) Codec(?CODEC) after time

--migration:split

CREATE OR REPLACE TABLE service_graph_edges_buffer ?ON_CLUSTER AS service_graph_edges
ENGINE = Buffer(currentDatabase(), service_graph_edges, 3, 5, 10, 10000, 1000000, 10000000, 100000000)
