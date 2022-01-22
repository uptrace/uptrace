# Uptrace Open Source demo

This example demonstrates how to quickly start Uptrace and ClickHouse using Docker images. It uses
[uptrace.yml](uptrace.yml) to configure Uptrace.

To run Uptrace permanently, we recommend using pre-compiled
[binaries](https://docs.uptrace.dev/guide/os.html#installation).

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
distro. To get some test data, just reload the UI few times. It usually takes about 30 seconds for
the data to appear.

See the [documentation](https://docs.uptrace.dev/guide/os.html#otlp) for configuring Uptrace client
for your programming language.
