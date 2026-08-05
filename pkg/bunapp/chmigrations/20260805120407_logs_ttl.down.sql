ALTER TABLE ?DB.logs_index ?ON_CLUSTER
MODIFY SETTING storage_policy = ?SPANS_STORAGE

--migration:split

ALTER TABLE ?DB.logs_index ?ON_CLUSTER
MODIFY TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE

--migration:split

ALTER TABLE ?DB.logs_data ?ON_CLUSTER
MODIFY SETTING storage_policy = ?SPANS_STORAGE

--migration:split

ALTER TABLE ?DB.logs_data ?ON_CLUSTER
MODIFY TTL toDate(time) + INTERVAL ?SPANS_TTL DELETE
