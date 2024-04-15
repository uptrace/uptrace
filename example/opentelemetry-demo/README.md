# OpenTelemetry Demo example for Uptrace

This example demonstrates how to run
[opentelemetry-demo](https://github.com/open-telemetry/opentelemetry-demo) with Uptrace backend.

## Using Docker

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

## Using Helm and Uptrace Cloud

Add OpenTelemetry Demo Helm repo:

```
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
```

Create `override-values.yml` file with the Otel Collector configuration for Uptrace. Don't forget to
specify your Uptrace DSN.

```yaml
opentelemetry-collector:
  config:
    exporters:
      otlp/uptrace:
        endpoint: https://otlp.uptrace.dev:4317
        tls: { insecure: false }
        headers:
          uptrace-dsn: '<YOUR_DSN_GOES_HERE>'

    service:
      pipelines:
        traces:
          exporters: [spanmetrics, otlp/uptrace]
        metrics:
          exporters: [otlp/uptrace]
        logs:
          exporters: [otlp/uptrace]
```

Start the demo:

```shell
helm install my-otel-demo open-telemetry/opentelemetry-demo --values override-values.yml
```

To check the status of OpenTelemetry Demo pods:

```shell
kubectl get pods
```
