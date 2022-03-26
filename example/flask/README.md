# Instrumenting Flask with OpenTelemetry example

Install dependencies:

```shell
pip install -r requirements.txt
```

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
UPTRACE_DSN=http://project2_secret_token@localhost:14317/2 python3 main.py
```

And open http://localhost:8000

See
[Getting started with Flask, SQLAlchemy, and OpenTelemetry](https://get.uptrace.dev/opentelemetry/flask.html)
for details.
