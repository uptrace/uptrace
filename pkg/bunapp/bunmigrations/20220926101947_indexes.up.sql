CREATE UNIQUE INDEX metrics_project_id_name_unq
ON metrics (project_id, name);

--bun:split

CREATE UNIQUE INDEX dashboards_project_id_template_id
ON dashboards (project_id, template_id);
