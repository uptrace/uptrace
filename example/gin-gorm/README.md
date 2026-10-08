# Gin and GORM example for OpenTelemetry and Uptrace

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
UPTRACE_DSN="http://project1_secret@localhost:14318?grpc=14317" go run .
```

Then open http://localhost:9999

See
[Getting started with Gin, GORM, and OpenTelemetry](https://uptrace.dev/get/instrument/opentelemetry-gin.html)
for details.
