SET statement_timeout = 0;

--bun:split

UPDATE notif_channels 
SET params = jsonb_set(
  COALESCE(params, '{}'), 
  '{authMethod}', 
  '"webhook"'
)
WHERE type = 'slack' 
  AND (params IS NULL OR params->'authMethod' IS NULL);

--bun:split
