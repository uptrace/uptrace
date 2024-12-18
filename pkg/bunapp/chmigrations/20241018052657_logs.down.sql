DROP TABLE IF EXISTS ?DB.logs_index ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.logs_data ?ON_CLUSTER
