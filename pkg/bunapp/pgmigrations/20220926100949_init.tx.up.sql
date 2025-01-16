CREATE TABLE taskq_jobs
(
  id bytea NOT NULL PRIMARY KEY,
  queue varchar(500) NOT NULL,
  run_at timestamptz,
  reserved_count int2 NOT NULL,
  reserved_at timestamptz,
  data bytea NOT NULL
);

--bun:split

CREATE INDEX IF NOT EXISTS taskq_jobs_queue_run_at_reserved_at_idx
ON taskq_jobs (queue, run_at, reserved_at);
