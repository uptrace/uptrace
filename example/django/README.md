# Instrumenting Django with OpenTelemetry

Install dependencies:

```shell
pip install -r requirements.txt
```

To run this example, [start](https://github.com/uptrace/uptrace/tree/master/example/docker) Uptrace
and run:

```shell
export UPTRACE_DSN=http://project2_secret_token@localhost:14318/2
./manage.py migrate
./manage.py runserver
```

And open http://localhost:8000

See
[Getting started with Django, PostgreSQL/MySQL, and OpenTelemetry](https://get.uptrace.dev/opentelemetry/django.html)
for details.
