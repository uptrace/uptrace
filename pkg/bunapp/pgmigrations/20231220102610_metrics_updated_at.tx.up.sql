ALTER TABLE metrics
ALTER COLUMN updated_at SET DEFAULT now();

--bun:split

UPDATE metrics SET updated_at = created_at
WHERE updated_at IS NULL;

--bun:split

ALTER TABLE metrics
ALTER COLUMN updated_at SET NOT NULL;
