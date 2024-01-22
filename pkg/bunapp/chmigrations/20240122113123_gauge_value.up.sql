DROP VIEW IF EXISTS ?DB.datapoint_hours_mv ?ON_CLUSTER

--migration:split

CREATE MATERIALIZED VIEW ?DB.datapoint_hours_mv ?ON_CLUSTER
TO ?DB.datapoint_hours
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

  any(string_keys) AS string_keys,
  any(string_values) AS string_values,
  max(annotations) AS annotations
FROM ?DB.datapoint_minutes AS m
GROUP BY m.project_id, m.metric, toStartOfHour(m.time), m.attrs_hash
