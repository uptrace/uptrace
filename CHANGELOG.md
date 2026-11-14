# Changelog

To get started with Uptrace, see https://uptrace.dev/get

## v2.1.0 - November 14 2026

  v2.1.0 is the first stable v2.1 release. It contains everything from v2.1.0-rc.1, rc.2 and rc.3
  below, plus the changes listed here.

  ### Upgrading

  Follow [Upgrading to v2.1.0](https://uptrace.dev/get/hosted/upgrade#upgrading-to-v210). Back up
  PostgreSQL and ClickHouse first, then run `uptrace preflight` and `uptrace migrate up` as the
  [upgrade guide](https://uptrace.dev/get/hosted/upgrade) describes.

  - **From v2.1.0-rc.1, rc.2 or rc.3**, install the new binary and run
    `uptrace migrate up --restart` (systemd), or the Docker and Helm steps of the guide. If you set
    `ch_cluster.kafka`, upgrade from rc.2 or later (see the rc.3 notes).
  - **From v2.0.0-beta.3 to v2.0.3, or a v2.1 beta**, first read
    [Upgrading from v2.0 or a v2.1 beta](https://uptrace.dev/get/hosted/upgrade#upgrading-from-classic):
    ClickHouse 26.3+ and PostgreSQL 15+ are required, `service.secret` must be set, and
    `uptrace migrate up` replaces the old `pg`/`ch` migration commands. Then read the breaking changes
    of each v2.1.0 release candidate below.
  - **v1.x and v2.0.0-beta.1, beta.2 and rc.2** cannot be upgraded. Install v2.1.0 on new databases.

  This upgrade moves incident causes to hypotheses. From the restart until `migrate up` finishes
  the backfill, the incident screens and routes wait (`read_alerts` hold). Run `migrate up` right
  after the restart; with `--restart` it carries on by itself.

  ### Breaking Changes

  - **Incident hypotheses replace the suspected-cause mark.** An incident now stores its possible
    causes as hypotheses. Each one cites the alerts, annotations or telemetry reads behind it and has
    a status: supported, unverified or disproved. A person can confirm or refute a hypothesis, and
    later analyses keep that ruling. The incident shows its origin alert from a confirmed hypothesis
    instead of a "suspected" mark. `migrate up` clears the existing marks and drops
    `incident_alerts.suspected`.
  - **`ingest.request_timeout` replaces `ingest.produce_wait_timeout`.** One deadline (default 10s,
    at most 10m) now covers the whole request from its arrival: the wait for a read slot, for room in
    the buffer, and for the Kafka ack. A shorter client deadline (`grpc-timeout`, `Connect-Timeout-Ms`,
    `X-Envoy-Expected-Rq-Timeout-Ms`) narrows it. `ingest.read_queue_timeout` now defaults to 1s, and
    `kafka.linger` must be shorter than half of the request timeout. The old key still works and logs
    a warning; `uptrace config fix --write` renames it.
  - **Prometheus remote write answers overload with 503**, not 429, because Prometheus drops the data
    of a 429 by default. Other receivers still answer a retryable 429 with `Retry-After`.

  ### API Changes

  In the beta `/internal/v1` API:

  - The incident analysis response no longer has `hypotheses`, `suspectedCauseAlertIds` and
    `alertsById`. `GET .../hypotheses` lists the hypotheses, and `PUT` / `DELETE
    .../hypotheses/{hypothesis_id}/ruling` confirms or refutes one.
  - `PUT .../alerts/{alert_id}/suspected` and `IncidentAlertGrouping.suspected` are removed in favor
    of `IncidentAlertGrouping.origin`.
  - The incident list returns `hypothesisCountsById`.
  - New MCP tool: `list_incident_hypotheses`.

  ### New Features

  - **Kafka acks per pipeline.** `pipelines.<signal>.acks.required` (`all`, the default, or `leader`)
    sets which replicas must hold a record before the broker acknowledges it, and
    `pipelines.<signal>.acks.await: false` lets a pipeline answer once the producer accepts its
    records, without waiting for the ack. Both trade durability for latency; see
    [Kafka acknowledgements](https://uptrace.dev/get/hosted/scale#kafka-acks).
  - **Single-zone Kafka topics.** A `kafka.placement` topic rule can keep each partition inside one
    zone (a broker's zone is its `broker.rack` up to the first `-`), so a follower in another
    datacenter no longer slows down every ack of its partition. Set `single_zone: true` on the rule;
    `uptrace kafka rebalance` places the replicas, and `uptrace kafka check-placement` lists
    partitions that span zones and prints the fix. See
    [Kafka placement](https://uptrace.dev/get/hosted/scale#kafka-placement).
  - **Self-monitoring per signal.** `self_monitoring.traces`, `metrics` and `logs` each take
    `disabled: true`, so you can stop exporting one signal and keep the others.
  - **Preflight checks ClickHouse migration blockers.** `uptrace preflight` fails on a failed
    mutation, a stuck `ON CLUSTER` DDL entry, too little disk space for the pending mutations, and a
    compact part that lacks a column its index reads. `uptrace migrate up --plan` names the rows and
    bytes each pending ClickHouse mutation rewrites, per host and disk.
  - **Threshold monitors with `in_total` and `on_average` window rules** now explain an alert by the
    window's total or average ("3 in total over the last 1h"), not by its newest point.

  ### Bug Fixes

  - **Ingestion.** A trace or Sentry request no longer releases its buffer room before the ack of
    its logs or events arrives, which let the buffer admit more than it could hold. A late Kafka ack
    answers a 429 whose text names the broker.
  - **gRPC.** The server pings idle connections every 30s, inside the 60s idle timeout of HAProxy and
    most load balancers, so idle OTel Arrow streams are no longer dropped by the proxy.
  - **Query editor.** A typed `order by` that names no column of the table is refused with an error,
    instead of sorting by an attribute that no record has. Pasting a query on Traces keeps the root
    `_parent_id` filter. The editor reads queries as the server does, so highlighting, completion and
    the filter chips agree with the result. A key that starts with a dot and a digit is quoted.
  - **Errors.** Client mistakes (4xx other than 401, 403 and 408) no longer record an exception on
    the server span, so they no longer open error-log alerts in self-monitoring. A failed call to an
    upstream service answers 5xx instead of 400.
  - **UI.** Alert policy forms are locked for a viewer who cannot edit them, and the alert note form
    and New monitor say why they are disabled. Facet drawers keep their row heights through a
    rebuild.

## v2.1.0-rc.3 - November 12 2026

  ### Upgrading

  Back up PostgreSQL and ClickHouse, then upgrade with `uptrace migrate up` as described in the
  v2.1.0-rc.1 notes.

  - **Indexes are built after the upgrade, not during it.** `migrate up` no longer builds
    ClickHouse skip indexes or projections, so an upgrade no longer waits days on a large table.
    A worker job checks every 10 minutes and builds each new index one partition at a time, newest
    partition first. Failed partitions are retried after an hour. On the first check after the
    upgrade, it goes through every partition once. That pass is cheap where a part already has
    the index, and it builds the missing ones, such as the new `display_name_ngram` index. Use
    `uptrace ch materialize_indexes status` to watch progress, `retry` to queue failed partitions
    again, and `run` to build now (`--table`, `--index`, `--partitions`, `--dry-run`). To stop the
    job, run `uptrace settings set index_build.enabled false`.
  - **ClickHouse writes through Kafka.** If you set `ch_cluster.kafka`, upgrade from v2.1.0-rc.2 or
    later. This release no longer reads the span, log and event data topics that older releases
    wrote.

  ### Breaking Changes

  - **`contains` and `like` on `_display_name` are literal again**
    ([#625](https://github.com/uptrace/uptrace/issues/625)). Since v2.1.0-rc.1, `contains` on
    `_display_name` matched whole words in any order, so `contains "[ERROR]"` found every line with
    the word "error". It is a case-insensitive substring match again, and `contains` and `like` mean
    the same thing on every column. A `|` in a `contains` needle separates alternatives
    (`contains "get|post"`). Use `search` to match whole words in any order. A new n-gram skip index,
    `display_name_ngram`, speeds up both operators. Review monitors that filter on `_display_name`.
  - **Negated compound filters skip records that lack the attribute.**
    `where not (host = "h1" and region = "us")` no longer matches a span, log or metric series that
    has no `host` or no `region`. This now matches how `!=` and `not in` already behaved. Queries
    that relied on the old result return fewer rows.
  - **Kafka 3.3 or newer** is required. Uptrace refuses to connect to an older broker.
  - **Source map uploads need a user token.** `POST /api/v1/sourcemaps`, which took the project token
    (DSN), is removed, because browser SDKs make the DSN public. Upload with
    `POST /api/v1/sourcemaps/{project_id}` and a user token that can edit the project, kept as a CI
    secret.
  - **More than one ClickHouse host requires replication.** Uptrace refuses to start when
    `ch_cluster` lists several hosts and `ch_cluster.replicated` is off, because migrations then
    created tables on the first host only. Set `ch_cluster.replicated` or list one host.
  - **The ClickHouse Kafka engine is no longer used.** With `ch_cluster.kafka.enabled`, the Uptrace
    worker reads the `click_*` topics from `kafka.brokers` and inserts into ClickHouse itself.
    `migrate up` drops the Kafka engine tables, and the `uptrace ch kafka_*` commands are removed.
    Per-table settings move from the flat `ch_schema.<table>.kafka_*` keys to a
    `ch_schema.<table>.kafka` block, and `ch_cluster.kafka.brokers` is ignored. See
    [Writing ClickHouse through Kafka](https://uptrace.dev/get/hosted/scale#writing-clickhouse-through-kafka).
  - **Configuration changes.** Run `uptrace config fix --write`. It moves the single model provider
    (`genai.protocol`, `api_key`, `base_url`, `model`, `headers`, `session_header`) into the new
    `genai.providers` list, `genai.assistant.data_notice` to `genai.data_notice`, and the ClickHouse
    Kafka keys above. Until you run it, Uptrace accepts the old keys and logs a warning.
  - **`uptrace ch init_replica` is now `uptrace ch clone_replica`.**

  ### API Changes

  - The assistant settings routes are replaced by `/internal/v1/genai/{project_id}` routes for
    providers, features and data notices.
  - The dashboard section and grid-layout routes are replaced by `/edit`, `/edit/preview`,
    `/edit/yaml` and the `/revisions` routes.
  - Attribute value listings no longer return `noData`.
  - The incident list returns the sort it applied. `sort_by` is a plain string on the incident and
    SLO subject lists.
  - Alert responses include the state of the alert's monitor (`monitor`, and `monitorsById` on lists).
  - The anomaly, outlier and forecast detail reads return an empty series, not 400, when the selected
    series is not in the window.
  - New MCP tools: `install_dashboard_templates` and `reset_dashboard_templates`.

  ### New Features

  - **Kafka ingestion no longer needs a license.** The Kafka pipeline runs in every build and with
    every license when `kafka.brokers` is set. See
    [Kafka-based ingestion](https://uptrace.dev/get/hosted/scale#kafka) for how to configure it.
  - **Dashboard layouts.** A new grid editor supports groups, rows and tabs, drag and drop with undo,
    and an automatic width that adapts to the screen. Table dashboards become linked layouts that each
    table row opens. Edit mode has a side panel with a YAML editor and a history of revisions that you
    can view and restore. `dashboard.max_revisions` (default 500) and
    `dashboard.keep_all_revisions_window` (default 168h) control how many revisions are kept.
    Dashboard YAML now uses a `layout` block. Documents in older formats still import.
  - **Model providers.** `genai.providers` names several providers, and each AI feature tries its
    chain of providers in order, moving to the next one when a provider fails. A project admin
    accepts the data notice (`genai.data_notice`) once before a feature calls a provider. Existing
    projects keep the features they had chosen. `genai.assistant.turn_limits` (default
    `[60/h, 300/24h]` per person) limits assistant use.
  - **AI features.** New alerts can be classified by symptom and affected part of the system, and you
    can filter the alert list by `alert.symptom` or `alert.category`. Incident analysis reads spans,
    logs, metrics and annotations through tools, then states each hypothesis with a confidence and its
    sources. A query helper in Explore and a metric monitor guide draft queries and monitors for you.
  - **SLO corrections UI.** You can schedule one-time, daily, weekly and monthly corrections, with a
    preview of their effect. SLO charts shade the excluded periods.
  - **Query text.** The query box also accepts SQL order, such as
    `select count() where status >= 400 group by service_name`. You can copy and paste a query
    between Spans and Traces. The new `_size` field is the stored size of each record in bytes, so
    `sum(_size)` shows what a service stores. CSV exports call it `_size` (was `_bytes`).
  - **Profiling** (`pipelines.profiling.enabled`): profiles over time, comparing two profiles, and
    filtering by attribute.
  - **Alerts from paused monitors are marked.** An open alert whose monitor is paused or disabled says
    so in the alert list and on the alert card, so a stale "firing" no longer looks live.
  - **ClickHouse operations.** `uptrace ch clone_progress` reports how far a cloned replica has got,
    with an ETA. `uptrace ch match_locality` moves a clone's parts to the same disks as its source.
    Replicas can name a `zone`, and `ch_cluster.zones` keeps the reads of each project in one zone.
  - **Kafka operations.** `uptrace kafka rebalance` balances disk use across brokers, `--drain`
    empties brokers, and `--partitions` adds partitions. `kafka.placement` limits which brokers may
    hold which topics, and `uptrace kafka check-placement` checks the cluster against it.
    `uptrace kafka lag --threshold` sets the lag it accepts. `kafka.replication_factor` sets the
    replica count of the topics Uptrace creates.
  - **Seed tools.** `uptrace db seed --dry-run` prints what a seed would change.
    `uptrace db assign-key` links a `seed_data` key to an existing row.

  ### Improvements

  - **Faster trace and log queries.** Root-span filters read the smaller group tables. More filters
    run in PREWHERE, and new key-set bloom filter and time min-max indexes skip more data.
    Attribute and facet listings stop at a row budget (`trace.read_budgets`), not at a timeout, and
    the UI says when a listing is partial.
  - **Ingestion visibility.** New counters show data that was dropped silently before:
    `uptrace.pipeline.consumer_dropped_records`, `uptrace.pipeline.cumtodelta_dropped_datapoints`,
    `uptrace.pipeline.suspended_requests` and `uptrace.pipeline.fitted_records`. The processing
    dashboard has a new Dropped data group. `uptrace.pipeline.insert_dropped_batches` is replaced by
    `uptrace.pipeline.insert_skipped_batches` and `uptrace.pipeline.insert_dropped_rows`.
  - **Operations.** `migrate status` shows the running migration and how long each step took.
    `/status` reports whether a worker has joined its Kafka consumer group. `uptrace doctor` checks
    that the ClickHouse cluster hosts match `ch_cluster.shards` and that the Kafka consumer group has
    members. PostgreSQL reads fail over to another replica within the same request. A `$VAR` in the
    config that names an unset variable logs a warning, and config errors no longer print values.
  - **Source maps** are keyed by debug ID, and the newest three builds at a URL are kept, so a browser
    still running an older build resolves against the right map.
  - **License page** shows the ingestion allowances and the premium features a license grants.
  - **Rate limits** count per account, not per IP address, and every 429 answer includes `Retry-After`.

  ### Bug Fixes

  - **Ingestion no longer acknowledges data it drops**
    ([#630](https://github.com/uptrace/uptrace/issues/630)). When ClickHouse was slow or stalled,
    Uptrace could answer 2xx for logs it then discarded. A request that cannot be buffered now gets a
    retryable 429 with `Retry-After` (gRPC `RESOURCE_EXHAUSTED`). `ingest.produce_wait_timeout`
    (default 5s) sets how long a request waits for room. With Kafka, a request succeeds only after the
    broker acknowledges its records. Otherwise it answers 429 and the client sends it again. When
    `kafka.delivery_timeout` is unset, ingest records expire after 2 × `ingest.produce_wait_timeout`
    (10s by default) and other records never expire. `0` means no limit for every producer. The new
    `uptrace.pipeline.delivery_failed_records` counts records that failed after the transport accepted
    them. `uptrace.pipeline.buffer_full_dropped_records` now counts records answered with 429. The
    ClickHouse setting `wait_for_async_insert` now defaults to 1, so a server restart loses no rows. It
    costs about 10% of insert throughput; set it to 0 in a replica's `query_settings` to opt out.
  - **Ingestion.** An over-long span name, status message or attribute is now cut instead of dropping
    the whole record. A histogram whose bucket counts do not match its bounds is refused and counted,
    not stored without buckets. A map under `message` in an OTLP log becomes a structured body. Clients
    that disconnect mid-body get 408, not 5xx. OTel Arrow streams are no longer cut at
    `listen.grpc.max_connection_age`. A worker no longer panics after `ch_max_insert_size` is
    lowered.
  - **Alerting.** Monitors with thousands of open alerts no longer fail with "max query size
    exceeded" and leave alerts unresolved. Monitors over cross-series queries (`share`, `topk`, joins)
    get the same value in every part of a check. `ignoring()` now matches labels the way PromQL does.
    Checks during `service.warm_up_period` are postponed instead of dropped. After a rolling restart,
    work spreads across processes again, and a backlog no longer runs stale monitor checks.
  - **Notifications.** A channel gets a "Resolved" only for a firing it received. Status emails name
    the event in the subject. Note and assignment emails thread under the alert. Alert emails show the
    monitor description again. Test-send errors no longer show the stored URL.
  - **Authentication.** `uptrace users change_password` refuses an empty password. A password reset
    confirms an unconfirmed account. Confirming an email address ends earlier sign-in links. API tokens
    cannot change an account's email. Token secrets are masked in SSO sessions.
  - **UI.** Deleting a widget, a dashboard or a log-grouping rule asks for confirmation. Query
    controls are readable in the dark theme. Lists return to the first page when their filter or sort
    changes. The Settings panel can set the theme back to the system theme. The in-app changelog page
    is removed.

## v2.1.0-rc.2 - October 22 2026

  ### Upgrading

  From v2.1.0-rc.1, install the new binary and run `uptrace migrate up --restart` (systemd), or use
  the Docker and Helm steps from the rc.1 notes. From v2.0.x or a v2.1 beta, follow
  [First upgrade from v2.0.x or a v2.1 beta](#first-upgrade-from-v20x-or-a-v21-beta) in the rc.1
  notes.

  This upgrade moves the state of each alert onto the alert row. From the restart until
  `migrate up` finishes the backfill, monitors are not checked, no alert notifications are sent, and
  the incident screens do not read their member alerts. Run `migrate up` right after the restart (with
  `--restart` it carries on by itself). Until the backfill finishes, `uptrace migrate status` prints a
  `Release holds` line and `uptrace doctor` warns (`capability_holds`). The backfill deletes alerts
  that have no events.

  ### Breaking Changes

  - **Ingest read slots are on by default.** All ingest receivers (OTLP over HTTP and gRPC, OTel
    Arrow, Prometheus remote write, CloudWatch, Vector, Sentry) now share a limit on request bodies
    read at the same time: `ingest.max_concurrent_reads`, by default 8 × GOMAXPROCS. A request waits
    up to `ingest.read_queue_timeout` (5s) for a slot, then gets a retryable 429 (gRPC:
    `ResourceExhausted`). Set `ingest.read_slots: false` to keep counting the slots without enforcing
    them. One gRPC connection can now open at most `listen.grpc.max_concurrent_streams` (100) streams.

  - **Metric attribute limits apply to every source.** A datapoint with more than
    `pipelines.metrics.max_attrs` (128) attributes, or with a key longer than
    `pipelines.metrics.max_attr_key_len` (256), is dropped and reported as rejected. Values longer
    than `pipelines.metrics.max_attr_value_len` (1000) are cut. These limits now apply to OTLP, OTel
    Arrow, Prometheus remote write and CloudWatch alike. `uptrace.pipeline.attr_limited_datapoints`
    counts each case by reason.

  - **Trace query (TQL) changes.**
    - A bare name after `=`, `!=`, `<`, `<=`, `>` or `>=` is now a second attribute, not a string:
      `where resp_size > req_size`. Quote strings: `where env = "prod"`. When no record carries such
      an attribute, `=` and `!=` on a string are refused with a hint to quote the value.
    - A duration or size literal now converts to the unit of the other side, so
      `wait_seconds > 1s` compares with 1, not 1000.
    - `sum`, `avg`, `median`, `quantile` and the `p50`–`p99` aggregates now weight each record by its
      sampling count, so results on sampled data change.
    - A filter on an array attribute no longer matches records without that attribute, including
      under `!=` and `not`.

  - **API tokens are created from a browser session only.** An API token can no longer create a
    personal token or a service account token, so a deleted token leaves no working tokens behind.
    Scripts that create tokens with a token get a 403.

  - **`pipelines.<signal>.max_ingest_per_min` is removed.** It answered success while dropping
    data. Overload is handled by the fair-share gate, which answers 429 with a retry hint. The old
    key loads with a warning, and `uptrace config fix --write` removes it.

  ### API Changes

  - `PUT /internal/v1/alerts/{project_id}/{alert_id}/assignees` and
    `PUT /internal/v1/incidents/{project_id}/{incident_id}/assignees` take
    `{"add": [...], "remove": [...]}` instead of `{"assigneeIds": [...]}`. The MCP tool
    `assign_incident` takes `add` and `remove`.
  - A chart widget write must include `tooltip` (`mode`, `sortOrder`) and is refused with 400
    without it. Chart legends are validated on write, and reads always return the full legend and
    tooltip.
  - The alert list takes `time_start` and `time_end`, so the MCP `list_alerts` tool returns exact
    counts for a time window.
  - Organization and project updates accept an optional `updatedAt`. When the record has changed
    since that read, they answer 409 `record_changed`.
  - Project responses no longer include `mfaRequired`. Read it from the organization.
  - Trace reads refuse `max_spans`, `max_logs` and `max_events` outside 10–100000 and
    `node_name_depth` outside 1–4 with 400, instead of clamping them.
  - Incident titles and summaries written by a person are refused past 1000 characters instead of
    being cut silently.

  ### New Features

  - **AI assistant.** Each project gets a chat that answers questions using the project's MCP tools
    and links each read it makes. A project admin turns it on in the new Assistant tab of the project
    settings, either by accepting the data notice of the deployment's provider or by configuring a
    provider for the project. The deployment's provider needs a `genai.protocol` that supports tool
    calls (`anthropic` or `openai`). New keys: `genai.assistant.model`, `genai.assistant.thinking`,
    `genai.assistant.context_tokens`, `genai.assistant.data_notice`, `genai.incidents.model`,
    `genai.incidents.thinking`, `genai.headers` and `genai.session_header`.
  - **Pie chart widget**, with two rings when it has two groupings.
  - **Chart tooltips.** Each chart widget can choose the tooltip mode (all, single or hidden) and
    the sort order. Click a chart to pin its tooltip, and scroll the rows when there are many series.
    A metric table or pie built from a capped read (10,000 groups) now says so.
  - **Span details.** New Dashboards, Compare and Explore tabs, and breadcrumb filters. Stack frames
    are now parsed from `exception.stacktrace` text for Go (including `pkg/errors`), browser
    JavaScript and Node.js, and library frames are folded. JavaScript and Node errors without
    `exception.type` regroup once by the type from the stack.
  - **Trace queries (TQL):** compare two attributes; `ipInRange`, `length`, `bucket`, `ipv4Subnet`,
    `timeOfDay` and `dayOfWeek` (with an optional IANA time zone); the `argMax`, `argMin`,
    `uniqValues` and `stddev` aggregates; the `d` and `w` duration units; and a leading `not` in the
    editor.
  - **Metric queries:** `count_over_time`, `quantile_over_time`, `share_gt_over_time`,
    `share_le_over_time`, `predict_linear`, `zscore_over_time`, `duration_over_time`,
    `topk_*`/`bottomk_*` (`avg`, `max`, `min`, `median`), `quantile`, `share`, `zscore`, `limitk`,
    `running_*`, `range_*`, `smooth_exponential`, and `label_set`, `label_del`, `label_keep`,
    `label_replace`.
  - **Prometheus HA pairs are deduplicated.** Labels in `prometheus.ha_replica_labels` (default
    `__replica__`), or the project's Replica labels setting, identify the replica. Uptrace stores one
    series and keeps one replica's samples, failing over after
    `pipelines.metrics.ha_failover_timeout` (30s).
  - **API token controls.** Tokens can have an expiration date, show when they were last used, and
    record who created them. Owners get an email 7 days before a token expires. Organization admins
    can list and block the tokens that reach their organization and cap the API token lifetime.
    Organizations can require SSO and set the SSO session length. Accounts get an email when 2FA is
    turned on or off and when recovery codes are regenerated or used.
  - **Webhook channels** show their URL again. Turn on "hide URL" for URLs that carry a secret, and
    add basic or bearer authentication.
  - **Monitor emails.** Projects and users can turn off the emails sent when a monitor is paused or
    resumed, or disabled. The person who made the change no longer gets the email.
  - **Logs.** ANSI color codes are removed from log and exception messages, so colored and plain
    lines group as one. A new project switch, "Extract attributes from messages", stops storing
    `key=value` pairs from messages that no grouping rule matches.
  - **Ingestion.** Prometheus remote write accepts zstd bodies and answers 415 for unsupported
    encodings. CloudWatch metric streams in the OpenTelemetry 1.0.0 format are accepted.
  - **Operations.**
    - The Go memory limit follows the cgroup limit (`service.memory_limit_ratio`, default 0.9;
      `GOMEMLIMIT` wins), and default buffer budgets are capped at a quarter of it.
    - `service.shutdown.delay` keeps serving while `/health/ready` fails, so load balancers drain
      first.
    - Config values can be read from files with `${file:<path>}`.
    - `uptrace migrate up` and `down` take `--batch-pause-ratio`.
    - `uptrace license show` prints the license.
    - `uptrace doctor` warns about leftover migration fences (`clickhouse_fence`) and active upgrade
      holds.
    - A new "Uptrace: Ingest Read Slots" dashboard and a monitor for long fair-share throttling help
      size ingestion.
  - **UI.**
    - Pinned dashboards at the top of the navigation panel, and search in the panel.
    - Duplicate in the query clause menu, and pinned filters in every filter builder.
    - Bulk owner assignment for alerts.
    - Reload and a conflict notice on settings forms: edits are kept when the record changes under
      them.
    - Respect for the system "reduce motion" setting.

  ### Improvements

  - **Incidents.** `genai.incidents.thinking` defaults to `medium`, which merged no unrelated alerts
    in testing. Endpoints that refuse the reasoning field are retried without it. Set it to `""` for
    the model's default.
  - **Anomaly detection.** Fewer false alerts on daily-seasonal series, seasonality is kept after a
    level shift, and previews and backtests match what the saved monitor does.
  - **Performance.** Faster anomaly detectors (vector kernels on amd64 and arm64), JSON handling and
    compression, and faster alert-state backfills during upgrades. New PostgreSQL indexes speed up
    deleting users and projects, the monitor list, and notification logs. Email lookups at login use
    the index and ignore letter case.

  ### Bug Fixes

  - **Alerting.**
    - Resuming a firing monitor no longer resets it or sends a false email.
    - The monitor list sorts by alert counts again.
    - Weighted error counts no longer drop fractional points.
    - Resolved SLO alerts keep their priority, and reminders no longer fire early.
    - Calendar SLOs refuse rule windows longer than the period, and subjects from a wider grouping no
      longer show the whole group's budget.
    - An outlier hold can be turned off at 10s resolution.
    - A crash in one SLO maintenance visit no longer stops the worker.
    - Notification templates bound `js`, `html` and `urlquery` output.
  - **Notifications.**
    - Telegram group-to-supergroup migrations are treated as permanent errors.
    - Google Chat links in status emails open the channel.
    - Names are cut by characters, not bytes.
    - A deleted user in an assignment no longer fails delivery.
    - Vendor incidents reopened during a send are closed.
  - **Traces and logs queries.**
    - `x / max(_dur_ms)` and error-rate expressions are no longer wrong on the pre-aggregated
      tables.
    - `_count` filters and `top3`/`top10`, `avg` and `min` over `_count` return consistent results.
    - The span histogram and the Compare tab show the same percentiles.
    - Time zone arguments work on images without zoneinfo.
    - Log timestamps in Go `time.String()`, PostgreSQL and `date(1)` formats with numeric zones are
      recognized.
    - Log grouping rules with certain captures or units load again.
    - An empty CSV export keeps its header.
  - **Metric queries.**
    - `uptime > 1d` parses.
    - `sort`, `topk` and `bottomk` rank on the visible range.
    - `mad`, `topk` and similar functions group correctly.
    - `deriv` and the count-like rollups report correct gaps and units.
    - A result of exactly the series limit is not marked truncated.
    - Exponential histograms whose bucket bounds overflow are rejected.
  - **Dashboards.**
    - Stat widget sparkline and bar follow the display mode.
    - Widgets stored with values that newer validation refuses can be saved, cloned and exported
      again.
    - Overlapping widgets in .NET templates are fixed.
    - Template sync continues past a failing project.
    - Over-long names answer 400, not 500.
    - Tags and overviews link only the project's own dashboards.
  - **Ingestion.**
    - A panic in an OTLP handler answers 500 or gRPC `Internal` instead of crashing the process.
    - Requests whose gRPC deadline has passed are not stored twice.
    - Corrupt compressed bodies answer 400.
    - Remote write accepts gzipped snappy bodies again and reports accurate written counts.
    - Sentry child spans follow the transaction's sampling, spans without a start time no longer get
      a 55-year duration, and breadcrumbs keep their keys.
    - Profiles honor `max_data_age` and `max_clock_skew`.
  - **Auth.**
    - Password reset links expire after one hour, as the email says.
    - In organizations that require 2FA, tokens created from a session without 2FA are refused.
      Older tokens keep working, and you can verify a token with a fresh 2FA.
    - Owners must enroll in 2FA before they can require it.
    - Users whose projects need SSO are sent to the SSO sign-in.
  - **Operations.**
    - TLS certificates load before the servers accept connections.
    - Shutdown no longer waits on idle connections, and OTel Arrow streams end cleanly.
    - The UI works under a `site.url` path named `internal` or `api`, and unknown API paths answer
      404.
    - An explicit `site.api_url` is kept when a license sets the site URL.
    - `service.profiling` works on the hosted build.
    - Prepaid licenses get their 3-day grace at startup.
    - Config fixes: an empty config file no longer panics, `consumer_disabled` turns the consumer
      off, `config fix` no longer empties a file, and commented-out `${file:}` references are
      ignored.
    - The DEB and RPM migration check reads variables from `/etc/uptrace/uptrace.conf`.
    - `migrate up --restart` no longer fails when systemd restarts a worker early.
    - ClickHouse steps on installs without distributed tables are fenced against concurrent
      inserts, which caused `Missing columns` during the rc.1 upgrade.
    - Kafka consumers no longer get stuck rejoining when heartbeats are slow.
    - Releases include `checksums.txt` for every artifact.

  ### Known Issues

  Both are fixed in v2.1.0-rc.3.

  - **`contains` on `_display_name` matches whole words in any order**
    ([#625](https://github.com/uptrace/uptrace/issues/625)). Since rc.1,
    `_display_name contains "[ERROR]"` matches any line with the word "error", and
    `contains "Driver ODBC"` matches "ODBC Driver". Searches and monitors that use it can match too
    much and fire on healthy systems.
  - **Acknowledged data can be lost when ClickHouse stalls**
    ([#630](https://github.com/uptrace/uptrace/issues/630)). When ClickHouse slowed down or paused
    and the ingest buffer filled, Uptrace could answer HTTP 200 (or an OTLP partial success, which
    clients do not retry) and drop the records. In testing, rc.2 lost most acknowledged logs during a
    long ClickHouse pause. With Kafka, records could also expire after the default 2-minute
    `kafka.delivery_timeout`.

## v2.1.0-rc.1 - October 07 2026

  ### Upgrading

  You can upgrade directly from v2.0.0–v2.0.3 and from every v2.1.0 beta (beta to beta.8). You do not
  need to install the releases in between. Upgrading from v2.0.0-beta.1, v2.0.0-beta.2, v2.0.0-rc.2,
  and v1.x is not supported.

  Read [Upgrading Uptrace](https://uptrace.dev/get/hosted/upgrade) before you start, especially
  [Upgrading from v2.0 or a v2.1 beta](https://uptrace.dev/get/hosted/upgrade#upgrading-from-classic).
  Back up PostgreSQL and ClickHouse first.

  ### System Requirements

  - **ClickHouse**: v26.3 or later (was v25.8). v2.0.x and the v2.1 betas keep working on ClickHouse
    26.3, so upgrade ClickHouse first.
  - **PostgreSQL**: v15 or later (was v14).
  - **Redis**: v6 or later.

  `uptrace preflight` checks both database versions before you upgrade.

  ### Breaking Changes

  - **New migration command.** `uptrace migrate up` replaces `pg init`, `pg migrate`, `ch init`,
    `ch migrate`, `ch check`, and the other `pg`/`ch` migration commands. It migrates PostgreSQL and
    ClickHouse together, in phases (expand, restart, backfill, finalize), so Uptrace keeps accepting
    data during an upgrade. The old commands print the replacement and exit with an error. Update your
    scripts, init containers, and entrypoints:

    ```shell
    uptrace --config=/etc/uptrace/config.yml migrate up
    uptrace --config=/etc/uptrace/config.yml db seed
    ```

    On the first run, `migrate up` takes over the migration records of v2.0.x and the v2.1 betas and
    does not run those migrations again.

  - **`serve` refuses an unmigrated database.** `uptrace serve` and `uptrace worker` exit with an
    error that names `uptrace migrate up` when the databases lack the schema of the release. The DEB
    and RPM packages start or restart the service only when no migration is pending, so the old
    version keeps serving until you migrate.

  - **`service.secret` must be set.** Uptrace refuses to start when `service.secret` is the `FIXME`
    placeholder that older example configs shipped with. Set it to a random value, for example the
    output of `openssl rand -hex 32`. A new secret logs out every user once. The DEB and RPM packages
    replace the placeholder for you.

  - **Configuration changes.** Run `uptrace config fix --write` to rewrite the options that moved or
    were renamed (for example top-level `spans` to `pipelines.spans`, and `ch_schema.metrics` to
    `ch_schema.timeseries` and `ch_schema.datapoints`) and to delete the options that are no longer
    used (for example `ch_schema.<table>.ttl_delete`). Until you do, Uptrace accepts the old options
    and logs a warning for each one.

    Several switches now have positive names and default to `true`: `auth.sign_up`,
    `auth.sign_in`, `listen.http2`, `listen.http3`, `sourcemaps.upload`, and
    `pipelines.*.kafka_consumer` replace the old `*_disabled` options, and `config fix` rewrites
    them. `pipelines.*.fairshare_disabled` is replaced by `fairshare` and has no automatic rewrite:
    Uptrace refuses to start while the old key is present.

  - **Outbound requests to private addresses are blocked.** Notification channels, webhooks, and
    source map downloads can no longer reach private, loopback, or link-local addresses by default.
    If a channel must reach an internal host, set `outbound_http.ssrf_protection` to
    `egress_proxy` or `off`.

  - **Query language changes.**
    - `count()` now always counts spans. Use `countDistinct()` to count distinct spans.
    - A `search` clause in a pipe-joined query ends at the next pipe. Quotes no longer extend it.
    - Trace search no longer supports the deprecated regular-expression matchers.
    - Cross-join aggregate references follow one per-trace rule, so some trace queries can return
      different values.

  - **Saved filter links.** Alert, incident, and monitor filter URLs use named fields. Links and
    bookmarks made with an older version open without a filter.

  - **GitHub and Google sign-in require a verified email address** at the provider.

  - **Per-project metric granularity is removed.** Metric datapoints keep the timestamp that the
    client sends.

  - **Renamed commands.** `sync_dashboards` is now `dashboard sync`, `ch move_partitions` is now
    `ch move_parts`, and `retention check` is now `storage retention`.

  ### First upgrade from v2.0.x or a v2.1 beta

  Processes of v2.0.x and the v2.1 betas cannot receive the restart request of
  `uptrace migrate up --restart`. While one is still connected, `migrate up` stops before the backfill
  and names it. Restart the processes yourself between the two runs:

  ```shell
  uptrace migrate up --before-release-change
  sudo systemctl restart uptrace
  uptrace migrate up
  ```

  Later upgrades can use `uptrace migrate up --restart`.

  **Docker:** pull the new `example/docker` files, set `service.secret` in `uptrace.yml`, then run
  `docker compose pull` and `docker compose up -d`. The container runs `uptrace migrate up` and
  `uptrace db seed` before it starts the server.

  **Kubernetes:** the Helm chart runs `migrate up --before-release-change` in a pre-upgrade Job and
  `migrate up` in a post-upgrade Job. Use `helm upgrade --wait`.

  ### API Changes

  - Incident settings moved from the project to `/incidents/{project_id}/settings`.
  - The project field `incidentsEnabled` is replaced by `incidentAutoGrouping`.
  - The incident membership response no longer has `attachedBy`; read `attachReason`.
  - Alert, incident, and monitor list filters use named fields. Clients must send the new format.
  - The MCP dashboard tools are now `create_dashboard` and `update_dashboard`.

  ### New Features

  - **Incidents.** Related alerts are grouped into incidents. Each incident has its own priority,
    owners, status, timeline and notes. Grouping can be automatic or manual, and manual use needs
    no model provider. You can also archive or merge incidents, file issues in external trackers,
    and get AI analysis with recommendations to review.
  - **SLOs.** Track reliability targets and error budgets over rolling windows. Burn-rate and
    budget-spent rules can alert per group or for the whole SLO. SLOs have their own section and
    can be imported and exported as YAML.
  - **Anomaly, outlier and forecast monitors** for metrics. They support seasonality, sensitivity
    presets, K-of-N confirmation and level-shift detection. You can preview and backtest them on
    interactive charts, and alerts explain how each decision was made.
  - **Alert policies** replace alert rules. A policy can set or remove attributes, set the
    priority, or archive matching alerts. Monitors and notification channels can be paused for a
    set time.
  - **New notification channels:** Email, incident.io and Pushover. Opsgenie channels can target
    Jira Service Management, and Alertmanager channels support authentication. Monitors can have
    descriptions, such as runbook links, and these appear in notifications.
  - **Traces and Explore.** A new [Traces](https://uptrace.dev/features/querying/traces) page
    filters whole traces with cross-join queries. Trace queries support compound expressions,
    `quantile()`, and arithmetic in `having`. The Tracing section is now Explore and includes
    Metrics and Annotations.
  - **Service graph** with server-side filters and per-node metrics.
  - **Full-text search** over spans and logs.
  - **Sentry SDK support.** Each project token has a Sentry DSN, so Sentry SDKs can send data to
    Uptrace. Span details show breadcrumbs and structured stack traces with nested causes. You
    can upload source maps to symbolicate JavaScript errors.
  - **Issue trackers.** Create a GitHub or Forgejo issue from any span or error group. You can
    edit the description before filing, and the issue links back to the span or group.
  - **MCP server.** Connect AI assistants to your project data and dashboards. The Project tab
    shows how to configure it.
  - **Metrics.** Prometheus remote write 2.0 is supported, and native histograms are accepted over
    remote write 1.0 and 2.0. Metric queries add PromQL-compatible rollup functions, `stddev`,
    `stdvar` and `having`.
  - **Global search.** Press ⌘K or Ctrl+K to move around the app and find project data.
  - **Debuggers.** Test transformation pipelines on real spans, logs and events. Test log
    grouping rules on sample or recent log lines.
  - **Setup and onboarding.** Each project gets a setup checklist, an integration picker and
    per-integration guides. Overview dashboards are suggested from the templates you can use.
  - **Retention.** Each project can set retention per signal, within limits set by the
    deployment.
  - **Operations.** PostgreSQL read replicas can serve read traffic, and the OTel Arrow receiver
    is now configurable. The `uptrace` binary adds the `kafka` (including `kafka rebalance`),
    `storage` and `dashboard` command groups.
  - **Authentication.** Members who must use two-factor authentication are guided through setup.
    SSO providers can be disabled without deleting them, and SAML requests are signed with
    SHA-256.

## v2.0.3 - May 13 2026

Same as v2.0.2, but rebuilt on a fresh Alpine image to address recent CVEs.

## v2.1.0-beta.4 - March 13 2026

- Restored deprecated `ch check` command. It does nothing.

## v2.1.0-beta.3 - March 11 2026

- Restored `max_insert_size` configuration option.

## v2.1.0-beta.2 - March 09 2026

No breaking changes.

### New Features

- Added [Traces](https://uptrace.dev/features/querying/traces) filtering.
- Added MCP server — see your Project tab for details.
- Added Incident.io notification channel.

### Upgrading

Make sure to run `pg init` and `ch init` commands during deployment.

## v2.1.0-beta - January 21 2026

### Upgrading

If you're already running an Uptrace installation, make sure to upgrade to v2.0.2 before upgrading
to v2.1.0.

### System Requirements

- **ClickHouse**: v25.8 or later (required for JSON data type support)
- **PostgreSQL**: v14 or later
- **Redis**: v6 or later

### Breaking Changes

- `uptrace serve` no longer automatically runs database migrations. Migrations must now be executed
  as a separate step before starting the server. Both the Dockerfile and Kubernetes Helm Charts have
  been updated to run migrations in a dedicated init container.

  To migrate manually, run the following commands in order:

  ```shell
  uptrace --config=/etc/uptrace/config.yml pg init
  uptrace --config=/etc/uptrace/config.yml pg migrate
  uptrace --config=/etc/uptrace/config.yml ch check
  uptrace --config=/etc/uptrace/config.yml ch init
  uptrace --config=/etc/uptrace/config.yml ch migrate
  uptrace --config=/etc/uptrace/config.yml ch dist
  uptrace --config=/etc/uptrace/config.yml db seed
  ```

  This change improves deployment reliability by ensuring migrations complete successfully before
  the application starts, and allows for better error handling during upgrades.

### Improvements

TODO

## v2.0.0 - July 24 2025

[Release notes](https://uptrace.dev/blog/uptrace-v20)

### Breaking Changes

- **Database Schema Rewrite**: ClickHouse and PostgreSQL schemas have been completely rewritten for
  improved performance and scalability.
- **ClickHouse JSON Type**: Uptrace now uses the ClickHouse
  [JSON](https://clickhouse.com/docs/sql-reference/data-types/newjson) data type to index span and
  log attributes. This change requires ClickHouse v25.3 or later.
- **Redis Dependency**: Redis is now required for in-memory caching operations.
- **Configuration Changes**: Some configuration options have been modified. Please review the
  migration guide below.

### Migration Guide

The recommended upgrade path is to run a separate Uptrace v2.x instance in parallel with your
existing v1.x instance. This approach allows you to:

1. Write data to both instances simultaneously
2. Validate data integrity and functionality in the new instance
3. Switch over when sufficient data has been collected in the new instance

This parallel deployment strategy mirrors our approach for deploying major changes to Uptrace Cloud,
ensuring minimal downtime and data loss.

### New Deployment Guides

- [Ansible Deployment Guide](https://uptrace.dev/get/hosted/ansible) - Automated deployment using
  Ansible playbooks
- [Kubernetes Deployment Guide](https://uptrace.dev/get/hosted/k8s) - Container orchestration
  deployment instructions

### New Features

- **Multi-Project Support**: Create and manage multiple projects and users directly through the UI
- **Data Transformations**: Support for custom
  [data transformations](https://uptrace.dev/features/transformations) to modify and enrich incoming
  telemetry data
- **Let's Encrypt Integration**: Automatic SSL certificate provisioning and renewal using Let's
  Encrypt
- **Large Trace Visualization**: Enhanced visualization capabilities for large traces containing
  10,000 to 1,000,000 spans
- **Bidirectional Span Links**: Automatic backlinks for target spans, providing improved trace
  navigation and context

### Licensing Changes

- **SSO and 2FA**: Single Sign-On and Two-Factor Authentication features now require a commercial
  license

### Performance Improvements

- **Flexible TTL Configuration**: Set separate Time-To-Live (TTL) values for spans, logs, and events
  tables
- **Enhanced Metrics Performance**: Improved metrics processing with reduced data storage
  requirements
- **Faster Attribute Operations**: Filtering and aggregating custom attributes is now 5-10x faster
  thanks to the new JSON data type implementation
- **Pre-aggregated Attributes**: Selected attributes in the tracing section are now pre-aggregated
  to further enhance query performance

### System Requirements

- **ClickHouse**: v25.3 or later (required for JSON data type support)
- **Redis**: Required for caching operations
- **PostgreSQL**: Compatible with rewritten schema

## v1.7.3 - May 8 2024

- `_count` is no longer an alias for `sum(_count)`, but instead points to the column. Use
  `sum(_count)` and `per_min(sum(_count))` instead.

- Fixed span duration unit in percentiles chart.

## v1.7.2 - April 16 2024

- `log_message` and `exception_message` are automatically promoted to `display_name` when used in
  logs. The original attribute is removed to avoid double indexing.

  Instead of `log_message contains "something"` you should use `display_name contains "something"`.

- The HTTP system is split into `httpclient` and `httpserver`.
- Improved service graph edge discovery.

## v1.7.1 - April 09 2024

- [Grafana APIs](https://uptrace.dev/get/grafana.html) now require project id and user token instead
  of Uptrace DSN. The old authentication method will continue working for a while.

## v1.7.0 - March 28 2024

This release is backwards compatible with v1.6. If you're upgrading from an older version, please
upgrade to the v1.6.x first, because some migrations from v1.6.x have been removed.

You can also check this [blog post](https://uptrace.dev/blog/uptrace-v17.html) for details.

#### Features

- Added dark mode.
- App navigation is reworked using the navigation drawer.
- Added a list of dashboards.
- Uptrace now provides quick access to dashboards that have the same prefix. You may need to reset
  your existing dashboards via UI to benefit from this.
- Added ability to query spans using system metrics provided by Uptrace: `uptrace_tracing_spans`,
  `uptrace_tracing_logs`, and `uptrace_tracing_events`.
- You can now search over monitors.
- Added pre-built dashboards for .NET system metrics reported by OpenTelemetry.
- Added ability to auto-reload data every 1 minute. See the date picker in the right corner.
- Added `user.auth_token` and support for [JSON API](https://uptrace.dev/get/json-api.html).

#### Improvements

- Added new project setting `force_span_name` by [@HeCorr](https://github.com/HeCorr). See the
  config for details.
- Added ability to disable TLS when configuring SMTP mailer by [@rgl](https://github.com/rgl).

#### Changed

- Trace and span ids are now hex-encoded.

#### Fixed

- Updated Tempo API to the latest version. You should be able to use
  [Tempo integration](https://uptrace.dev/get/grafana.html#tempo-traces) with the recent Grafana
  version.

## v1.6.2 - January 22 2024

Starting from this release, Uptrace returns all timeseries when a query doesn't contain an
aggregation:

```shell
# Returns all timeseries grouped by the _hash.
$metric_name

# Returns a single timeseries.
sum($metric_name)
```

As a result, Uptrace is now mostly compatible with the Prometheus query language, for example, the
following PromQL expressions are also valid in Uptrace:

- `increase($metric_name)` and `delta($metric_name)`
- `rate($metric_name[5m])` and `irate($metric_name[5m])`
- `avg_over_time($go_goroutines[5m])`
- `avg by (foo)(sum by(foo, bar)($metric_name))`
- `$metric_name offset 1w`
- Math between series with automatic many-to-one/one-to-many vector matching, for example,
  `sum($mem by (type)) / sum($mem) as mem`.

See [documentation](https://uptrace.dev/get/promql-compat.html) for more details.

## v1.6.0 - January 3 2024

This release is partially backwards incompatible with v1.5. PostgreSQL database will be migrated
automatically preserving your settings and dashboards, but you will need to reset the ClickHouse
database schema to upgrade:

```shell
uptrace ch reset
```

The recommended upgrade path is to start a separate Uptrace v1.6 instance in parallel with v1.5,
write data to both instances, and switch when you have enough data in the new instance. This is more
or less what we do when deploying major changes to Uptrace Cloud.

The license is changed from BSL to AGPL v3.0 which is a license approved by Open Software
Foundation.

You can also check this [blog post](https://uptrace.dev/blog/uptrace-v16.html) for details.

#### Breaking

- ClickHouse database schema is changed.

- Metric names and attributes are automatically changed to comply with Prometheus/Loki restrictions,
  for example, `service.name` becomes `service_name`.

- Upgraded to
  [v1.21](https://github.com/open-telemetry/opentelemetry-specification/blob/main/schemas/1.21.0)
  OpenTelemetry semantic conventions which introduced some breaking changes to attribute names. Most
  notably:
  - `http.method` is renamed to `http.request.method`.
  - `http.status_code` is renamed to `http.response.status_code`.

  Uptrace will automatically rename attributes, but you may need to update your favorite queries.

#### Features

- Added [service graphs](https://app.uptrace.dev/overview/1/service-graph). Service Graphs provide a
  visual representation of service interactions, dependencies, and performance metrics. Service
  graphs are built by analyzing span relationships and require certain span attributes.

- Added ability to group dashboard charts into rows, for example, charts can be grouped by category
  such as CPU, RAM, Network metrics.

- Added support for [Prometheus remote write](https://uptrace.dev/get/ingest/prometheus.html), which
  allows Prometheus to send its collected metrics data directly to a long-term storage solution such
  as Uptrace.

- Added ability to use Uptrace as a
  [Prometheus datasource in Grafana](https://uptrace.dev/get/grafana.html). Uptrace uses the
  original Prometheus engine so all Prometheus queries should be supported and you should be able to
  use existing Grafana dashboards with the Uptrace data source.

- Added [chart annotations](https://uptrace.dev/get/annotations.html) support. Annotations are
  labels or notes added to a chart to provide additional information or context. Annotations help
  clarify the data presented in the chart and help the viewer understand key points or trends.

#### Improvements

- Simplified UI for building dashboards and metric monitors.
- Added support for `per_min(sum(attr_key))` expressions in tracing query language.
- Added support for simple expressions like `sum(.duration) / .count` which is the same as
  `avg(.duration)`.
- Uptrace DSN now contains gRPC port so the same DSN can be used by all OpenTelemetry distributions
  provided by Uptrace such as uptrace-go, uptrace-js, etc.
- Improved templates for emails and Slack notifications.
- Added more metrics dashboards for HTTP checks and Kubernetes.
- Added more indexed semantic attributes in ClickHouse.
- Spans to metrics conversion requires a recent ClickHouse version that supports
  `allow_experimental_analyzer = 1`.

#### Changes

- Changed the default spans TTL to 7 days, metrics TTL to 30 days.

- There is now a single TLS config shared by HTTP and GRPC servers and defined in `listen.tls`
  section. The old configuration still works, but the first TLS config will be used.

## v1.5.2 - July 6 2023

- When authenticating via email and password, Uptrace ignores users in the PostgreSQL database and
  only allows users in the YAML config. This is needed when you want to disable email:password
  authentication completely.

## v1.5.0 - June 16 2023

This release is backwards compatible with v1.4.x, but contains a ClickHouse mutation to add new
columns.

#### Features

- Add support for `display.name` attribute. You can now use `display.name contains "get"` to search
  for spans and events/logs at the the same time. You can also use `display.name` attribute to
  [override](https://uptrace.dev/get/grouping.html) default span/event/log names.
- Ported Timeseries tab from [Uptrace Cloud](https://uptrace.dev/cloud) version.
- Allow to explore attribute values, for example, you can click on the `enduser.name` attribute to
  see the list of affected users.
- Metric dashboard templates can now automatically create monitors.
- Added ability to receive metrics and logs from
  [AWS CloudWatch](https://uptrace.dev/get/ingest/aws-cloudwatch.html).
- Uptrace can be used as a [Tempo data source](https://uptrace.dev/get/grafana.html) in Grafana.
  Grafana TraceQL is NOT supported, but Search is fully supported.

#### Improvements

- System picker now shows number of matching groups. Also added a preset for Spans / Logs / Events
  groups.
- `span.` prefix is replaced with `.`, for example, instead of `p50(span.duration)` you can write
  `p50(.duration)`. Old names are deprecated, but still supported.
- Added more metric dashboards and improved existing ones.
- Uptrace config now has sensible defaults so you can start with an empty YAML config file and add
  changes as you go.
- Added `uptrace config dump` to view the current Uptrace config in YAML format.
- Added a Docker image for linux/arm64 in addition to linux/amd64.

#### Fixes

- Preserve filters when navigating from "Overview" to "Tracing" tabs.
- Fixed numeric aggregations over custom attributes, for example, `sum(custom.attribute)`.
- Fixed attributes normalizations, for example, `X-Request-Id` is normalized to `x_request_id`.

#### Other changes

- Renamed `logs.LEVEL` to `logging.LEVEL` so it is not confused with logs processing.

## v1.4.0 - Apr 21 2023

#### Breaking changes

- PostgreSQL database is mandatory now. PostgreSQL is only used to store metadata such as users,
  dashboards, metric names, alerts etc. PostgreSQL DB usually only takes few megabytes of disk
  space.

  SQLite support is removed.

- Alerting rules in the config are ignored. Use Uptrace UI to create
  [monitors](https://uptrace.dev/get/alerting.html) instead.

- AlertManager client config is ignored. Use Uptrace UI to create AlertManager notification channel.

- `auth.users.username` is replaced with `auth.users.email` so users can receive email
  notifications. As the result, emails are used to authenticate users. To set user's name, use
  `auth.users.name` field.

#### Features

- Port Alerts, Monitors, and Notifications Channels from
  [Uptrace Enterprise](https://uptrace.dev/compare) edition.
- Accept errors and spans from [Sentry SDK](https://uptrace.dev/get/ingest/sentry.html).
- Documented [FluentBit](https://uptrace.dev/get/ingest/fluent-bit.html) integration.
- Add filter facets for metrics.

#### Improvements

- Improve UI for switching between table/grid metric views.
- Allow to quickly change group by in the grid.
- Allow to add/edit text gauges.
- Support table visualization in the grid view.
- Support heatmaps in the grid view.
- Allow to customize grid size
- Allow to customize colors for timeseries (just like units).
- Allow to edit dashboards using YAML.
- Add Kafka metrics dashboard.
- Add `site.path` setting to host Uptrace behind a proxy.

#### Fixes

- Fix cumulative to delta metrics conversion.
- Fix exponential histograms handling.
- Fix project token syncing.
- Respect `site.addr` when building OTLP/gRPC and OTLP/HTTP endpoints and DSNs.

### Migrating from previous versions

Uptrace v1.4.0 contains backwards incompatible changes that require resetting database schema.

To migrate to v1.4.0:

1. Stop Uptrace: `sudo systemctl stop uptrace`.
1. [Install](https://uptrace.dev/get/install.html) new version.
1. Configure PostgreSQL database in `uptrace.yml` config.

   ```yaml
   pg:
     addr: localhost:5432
     user: uptrace
     password: uptrace
     database: uptrace
   ```

1. Reset PostgreSQL and ClickHouse databases:

   ```shell
   uptrace pg reset
   uptrace ch reset
   ```

1. Start Uptrace: `sudo systemctl start uptrace`.

## v1.3.0 - Jan 20 2023

- Added ability to parse logs as spans using Vector remap language. See
  [documentation](https://uptrace.dev/get/ingest/vector.html#converting-logs-to-spans) and
  [PostgreSQL](https://dev.to/uptrace/monitoring-postgresql-15-logs-with-vector-and-uptrace-4165)
  example for details.

- Added support for Summary metrics type.

- Zipkin ingestion API now requires an Uptrace DSN in one of the following locations:
  - `uptrace-dsn` HTTP header.
  - `Authorization` HTTP header.
  - `dsn` URL query, for example, `/api/v2/spans?dsn=[dsn]`.

- Allow to configure [uptrace-go](https://uptrace.dev/get/opentelemetry-go.html) to collect Uptrace
  telemetry data:

```yaml
##
## uptrace-go client configuration.
## Uptrace sends internal telemetry here. Defaults to listen.grpc.addr.
##
uptrace_go:
  # dsn: http://project1_secret_token@localhost:14317/1
  # tls:
  #   cert_file: config/tls/uptrace.crt
  #   key_file: config/tls/uptrace.key
  #   insecure_skip_verify: true
```

- Added [sprig](http://masterminds.github.io/sprig/) functions to alerting
  [templates](https://uptrace.dev/get/alerting.html#templating).

- Allow to configure logging level via `logs.level` configuration option, for example:

```yaml
logs:
  level: ERROR
```

## v1.2.0 - Nov 8 2022

- Added ability to filter spans using facets.
- Added ability to select multiple systems.
- Added 2 quick filters by `deployment.environment` and `service.name` attributes on the Overview
  page.
- Added support for [creating metrics from spans](https://uptrace.dev/get/span-metrics.html) so they
  can be monitored like usual metrics, for example:

```yaml
# First, create a metric from incoming spans.
metrics_from_spans:
  - name: uptrace.tracing.spans
    description: Spans count (excluding events)
    instrument: counter
    unit: 1
    value: span.count
    attrs:
      - span.system as system
      - service.name as service
      - host.name as host
      - span.status_code as status
    where: not span.is_event

# Then, monitor that metric.
alerting:
  rules:
    - name: Service has high error rate
      metrics:
        - uptrace.tracing.spans as $spans
      query:
        - $spans{status="error"} / $spans > 0.1 group by service.name
      for: 5m
```

- Alerting rules annotations now support templating just like Prometheus, for example:

```yaml
alerting:
  rules:
    - name: Filesystem usage >= 90%
      metrics:
        - system.filesystem.usage as $fs_usage
      query:
        - group by host.name
        - group by device
        - where device !~ "loop"
        - $fs_usage{state="used"} / $fs_usage >= 0.9
      for: 5m
      annotations:
        summary: 'FS usage is {{ $values.fs_usage }} on {{ $labels.host_name }} and {{ $labels.device }}'
```

- Tweaked spans grouping and added 2 related options:
  - `project.group_by_env` - group spans by `deployment.environment` attribute.
  - `project.group_funcs_by_service` - group `funcs` spans by `service.name` attribute.

- Added project settings page where you can check available settings and project DSN.

- Changed ClickHouse schema to not use column names with dots in them which causes various issues
  with migrations, for example, such columns can't be renamed.

  If you have an existing ClickHouse database, you will have to reset it with:

```shell
uptrace ch reset
```

## v1.1.0 - Oct 1 2022

- Added additional ways to authenticate users via
  [Keycloak](https://uptrace.dev/get/auth-keycloak.html),
  [Google Cloud](https://uptrace.dev/get/auth-keycloak.html), and
  [Cloudflare](https://uptrace.dev/get/auth-cloudflare.html). Contributed by
  [@aramperes](https://github.com/aramperes).

- Added gauges support to Metrics UI. Only used in Redis Enterprise
  [example](example/redis-enterprise) for now.

- Renamed Logs tab to Events and moved all events there.

- Added support for PostgreSQL database instead of SQLite. This requires resetting SQLite database
  with:

```shell
uptrace db reset
```

- [Docker example](example/docker) is updated to work on Windows.
- Added [Redis Enterprise](example/redis-enterprise) example and metrics dashboards.

## v1.0.2 - Sep 8 2022

- Rename `alertmanager` YAML section to `alertmanager_client` so users don't get confused.

## v1.0.0 - Sep 6 2022

### Added

- Accept and store [OpenTelemetry Metrics](https://uptrace.dev/opentelemetry/metrics.html) in
  ClickHouse.
- Added metrics monitoring using [alerting rules](https://uptrace.dev/get/alerting.html).
- Added ability to send notifications via email/Slack/Telegram using
  [AlertManager](https://uptrace.dev/get/alerting.html#AlertManager).
- Added ability to configure [TLS](https://uptrace.dev/get/config.html#tls).
- Expand env vars in the YAML config, for example:

```yaml
ch:
  dsn: 'clickhouse://${CLICKHOUSE_USER}:@${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/${CLICKHOUSE_DB}?sslmode=disable'
```

### Upgrading

To upgrade, grab the latest
[uptrace.yml](https://github.com/uptrace/uptrace/blob/master/config/uptrace.yml) config, reset
ClickHouse database, and restart Uptrace:

```shell
uptrace ch reset
sudo systemctl restart uptrace
```

## v0.2.15 - Jun 8 2022

### 💡 Enhancements 💡

- Added support for accepting Zipkin spans at `http://uptrace:14318/api/v2/spans`.
- Added support for accepting Vector logs. See the [example](example/vector-logs).
- Added "Slowest groups" to the Overview tab
- Added new config option `ch_schema.compression`. You can now set ClickHouse compression via
  Uptrace config.
- Added new config option `ch_schema.replicated` if you want to use ClickHouse replication.
- Renamed the config option `retention.ttl` to `ch_schema.ttl`.
- Added new config option `spans.buffer_size`.
- Added new config option `spans.batch_size`.

## v0.2.14 - Apr 19 2022

- Fix incorrect `ORDER BY` when focusing on spans.
- Parse HTTP user agent into smaller parts.
- Always show `service.name` attribute when viewing traces.

## v0.2.13 - Apr 7 2022

- Fix incorrect `ORDER BY` when viewing spans.
- Improve config file auto-discovery.
- Update msgpack library.

## v0.2.12 - Mar 30 2022

- Automatically run ClickHouse migrations when Uptrace is started.
- Added 15 and 30 minutes periods.

## v0.2.11 - Mar 29 2022

- Added ability to filter spans by clicking on a chip.
- Added explore menu for each span attribute.
- Better handle situations when `service.name` or `host.name` attributes are not available.
- Support ZSTD and snappy decompression in OTLP.

## v0.2.8 - Mar 15 2022

- Fixed duration filter.
- Added chart resizing when window size is changed.
- Added GRPC metrics stub to remove errors from logs.

## v0.2.5 - Mar 09 2022

- Added list of spans.
- Fixed links to services and hostnames.
- Fixed SQL grouping.

## v0.2.4 - Feb 24 2022

### 💡 Enhancements 💡

- Make sure projects have unique tokens.
- Make user authentication optional by commenting out users section in the YAML config.
- Fixed jumping to a trace to account for the project id.
- Added missing trace find route.

## v0.2.2 - Feb 22 2022

### 💡 Enhancements 💡

- Added log out button.
- Added more concise syntax support to Uptrace query language, for example,
  `{p50,p90,p99}(span.duration)` instead of
  `p50(span.duration), p90(span.duration), p99(span.duration)`.
- Improved Uptrace query parsing.

## v0.2.0 - Jan 25 2022

### 💡 Enhancements 💡

- Added support for exceptions and in-app logs. See
  [Zap](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelzap) and
  [logrus](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otellogrus) integrations.
- Added services and hostnames overview.
- Added SQL query formatting when viewing spans.
- Require user authentication. Users are defined in the YAML config.
- Added support for having multiple isolated projects in the same database. Projects are defined in
  the YAML config.
- Added ability to filter query results, for example,
  `group by span.group_id | p50(span.duration) | where p50(span.duration) > 10ms`.
- Improved error handling on invalid Uptrace queries.
- Use faster and more compact MessagePack encoding to store spans in `spans_data` table.
- Add more attributes to ClickHouse index.

To upgrade, reset ClickHouse schema with the following command (existing data will be lost):

```go
# Using binary
./uptrace --config=/etc/uptrace/uptrace.yml ch reset

# Using sources
go run cmd/uptrace/main.go --config=config/uptrace.yml ch reset
```

## v0.1.0 - Dec 27 2021

Initial release.
