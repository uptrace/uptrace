# Go-zero api example
## 1. Start Uptrace
To run this example, 
[start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace

## 2. Add Uptrace to go-zero api config

Add these below config to `api/etc/api-api.yaml`, to enable uptrace report for go-zero api

```yaml

Telemetry:
  Name: api-api
  Endpoint: localhost:14317
  Sampler: 1.0
  Batcher: otlpgrpc
  OtlpHeaders:
    uptrace-dsn: http://project2_secret_token@localhost:14317/2

```

## 3. Start the go-zero server
```shell
cd api
go run api.go -f etc/api-api.yaml
```
## 4. Test the reporting

Then open http://localhost:8888/from/you   
Check the result in Uptrace UI http://localhost:14318/spans/2/  

See
[Getting started with GoZero and OpenTelemetry](https://uptrace.dev/get/ingest/opentelemetry.html#already-using-opentelemetry-sdk)
for details.
