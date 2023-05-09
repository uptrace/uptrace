# Go-zero api example

**Step 1**. [Start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace using
Docker.

**Step 2**. Update go-zero config at `api/etc/api-api.yaml` to start sending data to Uptrace:

```yaml
Telemetry:
  Name: api-api
  Endpoint: localhost:14317
  Sampler: 1.0
  Batcher: otlpgrpc
  OtlpHeaders:
    uptrace-dsn: http://project2_secret_token@localhost:14317/2
```

**Step 3**. Start the go-zero server:

```shell
go run api.go -f etc/api-api.yaml
```

**Step 4**. Then open http://localhost:8888/from/you to trigger a request for go-zero API.

**Step 5**. Open Uptrace UI at [http://localhost:14318](http://localhost:14318).

See
[Getting started with GoZero and OpenTelemetry](https://uptrace.dev/get/ingest/opentelemetry.html#already-using-opentelemetry-sdk)
for details.
