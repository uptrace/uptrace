alter table alerts
add column span_group_id int8;

--bun:split

update alerts
set span_group_id = trackable_id
where trackable_model = 'SpanGroup' and trackable_id is not null;

--bun:split

alter table alerts
drop column dedup_hash;

--bun:split

alter table alerts
drop column trackable_id;

--bun:split

alter table alerts
drop column trackable_model;

--bun:split

CREATE UNIQUE INDEX alerts_project_id_span_group_id_unq
ON alerts (project_id, span_group_id)
WHERE span_group_id IS NOT NULL;
