# Uptrace Open Source demo

This example demonstrates how to start Uptrace and ClickHouse using Docker. It uses
[uptrace.yml](uptrace.yml) to configure Uptrace.

**Step 1**. Start the services:

```shell
docker-compose up -d
```

**Step 2**. Make sure Uptrace is running:

```shell
docker-compose logs uptrace
```

**Step 3**. Open Uptrace UI at http://localhost:14318

Uptrace will monitor itself using [uptrace-go](https://github.com/uptrace/uptrace-go) OpenTelemetry
distro. To get some test data, just reload the UI few times.

See the [documentation](https://docs.uptrace.dev/guide/os.html#otlp) for configuring Uptrace client
for your programming language.
