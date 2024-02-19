CREATE INDEX monitors_project_id_name_idx ON monitors
USING GIN (project_id, name);
