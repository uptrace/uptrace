alter table service_graph_edges ?ON_CLUSTER
add column client_attr LowCardinality(String) Codec(?CODEC) after time

--migration:split

DROP TABLE IF EXISTS service_graph_edges_buffer ?ON_CLUSTER

--migration:split

CREATE TABLE service_graph_edges_buffer ?ON_CLUSTER AS service_graph_edges
ENGINE = Buffer(currentDatabase(), service_graph_edges,
  2,
  5, 45,
  1_000_000, 1_000_000,
  500_000_000, 500_000_000)
