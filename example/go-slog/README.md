# OpenTelemetry slog example for Uptrace

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
UPTRACE_DSN="http://project1_secret@localhost:14318?grpc=14317" go run .
```

Then open the URL from the console output to view the trace.

See [OpenTelemetry Slog](https://uptrace.dev/get/instrument/opentelemetry-slog.html) for details.
