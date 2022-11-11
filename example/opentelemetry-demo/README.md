# OpenTelemetry Demo example for Uptrace

This example demonstrates how to run
[opentelemetry-demo](https://github.com/open-telemetry/opentelemetry-demo) with Uptrace backend.

**Step 1**. Download the opentelemetry-demo using Git:

```shell
git clone https://github.com/uptrace/opentelemetry-demo.git
cd opentelemetry-demo
```

**Step 2**. Start the demo:

```shell
docker compose up --no-build
```

**Step 3**. Make sure Uptrace is running:

```shell
docker-compose logs uptrace
```

**Step 4**. Open Uptrace UI at
[http://localhost:14318/overview/2](http://localhost:14318/overview/2)

If something is not working, check OpenTelemetry Collector logs:

```shell
docker-compose logs otelcol
```
