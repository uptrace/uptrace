# Uptrace Docker demo

## Getting started

This example demonstrates how to quickly start Uptrace using Docker. For full instructions, see the
[Docker guide](https://uptrace.dev/get/hosted/docker). You can also deploy Uptrace using
[Ansible](https://uptrace.dev/get/hosted/ansible) or [Kubernetes](https://uptrace.dev/get/hosted/k8s).

**Step 1**. Download the example using Git:

```shell
git clone https://github.com/uptrace/uptrace.git
cd uptrace/example/docker
```

**Step 2**. Start the services using Docker:

```shell
docker compose pull
docker compose up -d
```

**Step 3**. Make sure all services are running and Uptrace is ready:

```shell
docker compose ps
docker compose logs uptrace
curl http://localhost:14318/health/ready
```

**Step 4**. Open Uptrace UI at [http://localhost:14318](http://localhost:14318). To log in, use
email `admin@uptrace.local` and password `admin`.

Uptrace will monitor itself using [uptrace-go](https://github.com/uptrace/uptrace-go) OpenTelemetry
distro. To get some test data, just reload the UI few times. It usually takes about 30 seconds for
the data to appear.

To configure OpenTelemetry for your programming language, see
[documentation](https://uptrace.dev/get).

## Services

| Service       | Host ports                     | Description                                                 |
| ------------- | ------------------------------ | ----------------------------------------------------------- |
| uptrace       | 14318 (HTTP/UI), 14317 (gRPC)  | Uptrace UI, API, and OTLP ingestion                         |
| otelcol       | 4317 (gRPC), 4318 (HTTP)       | OpenTelemetry Collector that forwards data to Uptrace       |
| clickhouse    | 8123 (HTTP), 9000 (native)     | Stores spans, logs, and metrics                             |
| postgres      | 5432                           | Stores users, projects, dashboards, and monitors            |
| redis         | —                              | Cache                                                       |
| mailpit       | 8025 (UI), 1025 (SMTP)         | Captures emails sent by Uptrace                             |
| grafana       | 3000                           | Grafana with Uptrace Tempo and Prometheus data sources      |
| prometheus    | —                              | Scrapes node_exporter and remote-writes metrics to Uptrace  |
| node_exporter | 9100 (host network)            | Host metrics for Prometheus                                 |
| vector        | —                              | Generates demo logs and sends them to Uptrace               |

If `docker compose up` fails with `port is already allocated`, stop the program that uses the port
or change the host port in `docker-compose.yml`.

## Sending data

Point your OpenTelemetry SDK at the collector, which adds the project DSN for you:

```shell
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

You can also send data to Uptrace directly. In that case, pass the project DSN in the `uptrace-dsn`
header:

```shell
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:14317
export OTEL_EXPORTER_OTLP_HEADERS=uptrace-dsn=http://project1_secret@localhost:14318?grpc=14317
```

## Alerting

This example uses Mailpit to test email notifications. Open
[http://localhost:8025](http://localhost:8025) to view available email notifications.

See [Alerting and Notifications](https://uptrace.dev/features/alerting) for more details.

## OpenTelemetry Collector

This example also comes with a pre-configured OpenTelemetry Collector to monitor
[host metrics](https://uptrace.dev/opentelemetry/collector/host-metrics) and
[PostgreSQL](https://uptrace.dev/guides/opentelemetry-postgresql).

## Grafana

Open Grafana at [http://localhost:3000](http://localhost:3000) and log in with `admin` / `admin`.
Grafana comes with the Uptrace Tempo and Prometheus data sources, so you can query Uptrace traces and
metrics from Grafana.

## Cleanup

To stop the services and delete all data:

```shell
docker compose down -v
```
