DROP TABLE IF EXISTS ?DB.measure_minutes_buffer ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.measure_minutes ?ON_CLUSTER

--migration:split

DROP TABLE IF EXISTS ?DB.measure_hours ?ON_CLUSTER

--migration:split

DROP VIEW IF EXISTS ?DB.measure_hours_mv ?ON_CLUSTER
