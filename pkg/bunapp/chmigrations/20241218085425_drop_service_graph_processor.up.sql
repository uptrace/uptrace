DROP TABLE IF EXISTS ?DB.service_graph_edges ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.service_graph_edges_buffer ?ON_CLUSTER