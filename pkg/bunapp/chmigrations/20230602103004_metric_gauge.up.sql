DROP TABLE ?DB.measure_minutes_buffer ?ON_CLUSTER

--migration:split

DROP VIEW ?DB.measure_hours_mv ?ON_CLUSTER

--migration:split

ALTER TABLE ?DB.measure_minutes ?ON_CLUSTER
RENAME COLUMN value TO gauge

--migration:split

ALTER TABLE ?DB.measure_hours ?ON_CLUSTER
RENAME COLUMN value TO gauge

--migration:split

CREATE MATERIALIZED VIEW ?DB.measure_hours_mv ?ON_CLUSTER
TO ?DB.measure_hours
AS SELECT
  project_id,
  metric,
  toStartOfHour(time) AS time,
  attrs_hash,

  anyLast(instrument) AS instrument,
  min(min) AS min,
  max(max) AS max,
  sum(sum) AS sum,
  sum(count) AS count,

  anyLast(gauge) AS gauge,
  quantilesBFloat16MergeState(0.5)(histogram) AS histogram,

  anyLast(string_keys) AS string_keys,
  anyLast(string_values) AS string_values,
  max(annotations) AS annotations
FROM ?DB.measure_minutes
GROUP BY project_id, metric, toStartOfHour(time), attrs_hash
SETTINGS prefer_column_name_to_alias = 1

--migration:split

CREATE TABLE ?DB.measure_minutes_buffer ?ON_CLUSTER AS ?DB.measure_minutes
ENGINE = Buffer(?DB, measure_minutes, 3, 5, 10, 10000, 1000000, 10000000, 100000000)
