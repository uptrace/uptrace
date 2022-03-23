# Instrumenting Flask with OpenTelemetry

Install dependencies:

```shell
pip install -r requirements.txt
```

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
export UPTRACE_DSN=http://project2_secret_token@localhost:14317/2
python3 main.py
```

And open http://localhost:8000

See [Getting started with Flask and SQLAlchemy](https://get.uptrace.dev/opentelemetry/flask.html)
for details.
