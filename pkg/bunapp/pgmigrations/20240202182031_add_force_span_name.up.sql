ALTER TABLE projects
ADD COLUMN force_span_name text[];

--bun:split

alter table users
add column auth_token varchar(500);
