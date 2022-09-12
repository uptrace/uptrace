CREATE TABLE metrics (
  id INTEGER PRIMARY KEY,
  project_id int4 NOT NULL,

  name varchar(1000) NOT NULL,
  description varchar(1000),
  unit varchar(100),
  instrument varchar(100) NOT NULL,

  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--bun:split

CREATE UNIQUE INDEX metrics_project_id_name_unq
ON metrics (project_id, name);

--bun:split

CREATE TABLE dashboards (
  id INTEGER PRIMARY KEY,
  project_id int4 NOT NULL,
  template_id varchar(100),

  name varchar(1000) NOT NULL,
  base_query varchar(1000),

  is_table boolean,
  metrics varchar(1000),
  query varchar(1000),
  columns varchar(1000)
);

--bun:split

CREATE UNIQUE INDEX dashboards_project_id_template_id
ON dashboards (project_id, template_id);

--bun:split

CREATE TABLE dash_entries (
  id INTEGER PRIMARY KEY,
  dash_id int8 NOT NULL REFERENCES dashboards (id) ON DELETE CASCADE,
  project_id int4 NOT NULL,

  name varchar(1000) NOT NULL,
  description varchar(1000),
  weight int4 NOT NULL,
  chart_type varchar(100) NOT NULL DEFAULT 'line',

  metrics varchar(1000) NOT NULL,
  query varchar(1000) NOT NULL,
  columns varchar(1000) NOT NULL
);

--bun:split

CREATE TABLE dash_gauges (
  id INTEGER PRIMARY KEY,

  project_id int4 NOT NULL,
  dash_id int8 NOT NULL REFERENCES dashboards (id) ON DELETE CASCADE,
  dash_kind varchar(100) NOT NULL,

  name varchar(1000) NOT NULL,
  description varchar(1000),
  weight int4 NOT NULL,
  template varchar(1000) NOT NULL,

  metrics varchar(1000) NOT NULL,
  query varchar(1000) NOT NULL,
  columns varchar(1000) NOT NULL
);

--bun:split

CREATE TABLE alerts (
  rule_id INTEGER PRIMARY KEY,
  alerts varchar(10000)
);
