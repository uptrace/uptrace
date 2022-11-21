# Kvrocks example for OpenTelemetry and Uptrace

See
[Getting started with Kvrocks and go-redis](https://kvrocks.apache.org/blog/go-redis-kvrocks-opentelemetry)
for details.

**Step 1**. Download the example using Git:

```shell
git clone https://github.com/uptrace/uptrace.git
cd uptrace/example/kvrocks
```

**Step 2**. Start the services using Docker:

```shell
docker-compose pull
docker-compose up -d
```

**Step 3**. Make sure Uptrace is running:

```shell
docker-compose logs uptrace
```

**Step 4**. Open Uptrace UI at [http://localhost:14318](http://localhost:14318)
