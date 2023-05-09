# Uptrace and ClickHouse cluster example

**Step 1**. Download the example using Git:

```shell
git clone https://github.com/uptrace/uptrace.git
cd uptrace/example/docker
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
