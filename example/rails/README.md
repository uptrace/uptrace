# Instrumenting Rails with OpenTelemetry example

Install dependencies:

```shell
bundle install
```

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
UPTRACE_DSN="http://project1_secret@localhost:14318?grpc=14317" rackup main.ru
```

And open http://localhost:9292

## Documentation

See [Getting started with Rails and OpenTelemetry](https://uptrace.dev/get/opentelemetry-rails.html)
for details.
