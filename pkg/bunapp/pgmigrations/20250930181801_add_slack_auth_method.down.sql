SET statement_timeout = 0;

--bun:split

UPDATE notif_channels 
SET params = params - 'authMethod'
WHERE type = 'slack' 
  AND params ? 'authMethod';

--bun:split
