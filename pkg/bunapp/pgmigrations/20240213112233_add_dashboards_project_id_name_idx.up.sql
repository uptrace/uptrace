CREATE EXTENSION IF NOT EXISTS pg_trgm;

--migration:split

CREATE INDEX dashboards_project_id_name_idx ON dashboards
USING GIN (project_id, name);
