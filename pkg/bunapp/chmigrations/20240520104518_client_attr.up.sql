alter table service_graph_edges ?ON_CLUSTER
add column client_attr LowCardinality(String) Codec(?CODEC) after time