CREATE TABLE dashboards (
  id int8 PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,

  project_id int4 NOT NULL,
  template_id varchar(500),

  name varchar(1000) NOT NULL,
  base_query varchar(500),

  is_table boolean,
  metrics jsonb,
  query varchar(500),
  columns jsonb
);

--bun:split

CREATE UNIQUE INDEX dashboards_project_id_template_id
ON dashboards (project_id, template_id);