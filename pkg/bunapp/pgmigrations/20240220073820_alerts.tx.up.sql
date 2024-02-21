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

--bun:split

alter table alerts
add column muted_until timestamptz;

--bun:split

update dashboards
set template_id = 'uptrace.postgresql.10.databases'
where template_id = 'uptrace.db.postgresql_by_host_database';

--bun:split

update dashboards
set template_id = 'uptrace.postgresql.20.tables'
where template_id = 'uptrace.db.postgresql_by_host_database_table';

--bun:split

update dashboards
set template_id = 'uptrace.postgresql.30.indexes'
where template_id = 'uptrace.db.postgresql_table_indexes';

--bun:split

update dashboards
set template_id = 'uptrace.postgresql.40.bgwriter'
where template_id = 'uptrace.db.postgresql_bgwriter';

--bun:split

update dashboards
set template_id = 'uptrace.kafka.10.topics'
where template_id = 'uptrace.kafka.topics';

--bun:split

update dashboards
set template_id = 'uptrace.kafka.20.partitions'
where template_id = 'uptrace.kafka.partitions';

--bun:split

update dashboards
set template_id = 'uptrace.kafka.30.consumer_groups'
where template_id = 'uptrace.kafka.consumer_groups';

--bun:split

update dashboards
set template_id = 'uptrace.dotnet.10.all'
where template_id = 'uptrace.dotnet.all';

--bun:split

update dashboards
set template_id = 'uptrace.dotnet.20.gc'
where template_id = 'uptrace.dotnet.gc';

--bun:split

update dashboards
set template_id = 'uptrace.dotnet.30.runtime'
where template_id = 'uptrace.dotnet.runtime';

--bun:split

update dashboards
set template_id = 'uptrace.dotnet.40.thread_pool'
where template_id = 'uptrace.dotnet.thread_pool';

--bun:split

update dashboards
set template_id = 'uptrace.dotnet.50.jit'
where template_id = 'uptrace.dotnet.jit';

--bun:split

update dashboards
set template_id = 'uptrace.hostmetrics.10.overview'
where template_id = 'uptrace.system.overview_by_host';

--bun:split

update dashboards
set template_id = 'uptrace.hostmetrics.20.filesystems'
where template_id = 'uptrace.system.filesystem_by_host_device';

--bun:split

update dashboards
set template_id = 'uptrace.hostmetrics.30.disks'
where template_id = 'uptrace.system.disk_by_host_device';

--bun:split

update dashboards
set template_id = 'uptrace.hostmetrics.40.network'
where template_id = 'uptrace.system.network_by_host';

--bun:split

update dashboards
set template_id = 'uptrace.php_fpm.10.pools'
where template_id = 'uptrace.php.fpm_pools';

--bun:split

update dashboards
set template_id = 'uptrace.php_fpm.20.servers'
where template_id = 'uptrace.php.fpm_pool_servers';

--bun:split

update dashboards
set template_id = 'uptrace.k8s.10.containers'
where template_id = 'uptrace.k8s.containers';

--bun:split

update dashboards
set template_id = 'uptrace.k8s.20.nodes'
where template_id = 'uptrace.k8s.nodes';

--bun:split

update dashboards
set template_id = 'uptrace.k8s.30.pods'
where template_id = 'uptrace.k8s.pods';

--bun:split

update dashboards
set template_id = 'uptrace.k8s.40.network'
where template_id = 'uptrace.k8s.nodes_network';

--bun:split

update dashboards
set template_id = 'uptrace.rpc.10.clients'
where template_id = 'uptrace.rpc.client_by_service_method_host';

--bun:split

update dashboards
set template_id = 'uptrace.rpc.20.servers'
where template_id = 'uptrace.rpc.server_by_service_method';

--bun:split

update dashboards
set template_id = 'uptrace.tracing.20.hosts'
where template_id = 'uptrace.spans_by_host';

--bun:split

update dashboards
set template_id = 'uptrace.tracing.10.services'
where template_id = 'uptrace.spans_by_service';

--bun:split

update dashboards
set template_id = 'uptrace.enterprise.20.processing'
where template_id = 'uptrace.internal.processing';

--bun:split

update dashboards
set template_id = 'uptrace.enterprise.10.ingestion'
where template_id = 'uptrace.internal.projects';

--bun:split

update dashboards
set template_id = 'uptrace.billing.10.projects'
where template_id = 'uptrace.internal.billing';

--bun:split

update dashboards
set template_id = 'uptrace.aws.10.ec2_instances'
where template_id = 'uptrace.aws.ec2_instances';

--bun:split

update dashboards
set template_id = 'uptrace.aws.20.ebs'
where template_id = 'uptrace.aws.ec2_ebs';

--bun:split

update dashboards
set template_id = 'uptrace.aws.30.ebs_volumes'
where template_id = 'uptrace.aws.ebs_volumes';

--bun:split

update dashboards
set template_id = 'uptrace.aws.40.rds'
where template_id = 'uptrace.aws.rds_instances';

--bun:split

update dashboards
set template_id = 'uptrace.node_exporter.10.cpu_ram'
where template_id = 'uptrace.node_exporter.cpu_ram';

--bun:split

update dashboards
set template_id = 'uptrace.node_exporter.20.filesystems'
where template_id = 'uptrace.node_exporter.filesystem';
