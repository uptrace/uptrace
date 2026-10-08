# Instrumenting Flask with OpenTelemetry example

Install dependencies:

```shell
pip install -r requirements.txt
```

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
UPTRACE_DSN="http://project1_secret@localhost:14318?grpc=14317" python3 main.py
```

And open http://localhost:8000

See
[Getting started with Flask, SQLAlchemy, and OpenTelemetry](https://uptrace.dev/get/opentelemetry-flask-sqlalchemy.html)
for details.
