# Uptrace demo for monitoring Redis Enterprise clusters

## Starting containers

**Step 1**. Download the example using Git:

```shell
git clone https://github.com/uptrace/uptrace.git
cd uptrace/example/redis-enterprise
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

## Creating Redis Enterprise cluster

To create a Redis Enterprise cluster, open [https://localhost:8443/](https://localhost:8443/) and
then follow official
[instructions](https://docs.redis.com/latest/rs/installing-upgrading/get-started-docker/).

Once everything is done, you should be able to connect to the created Redis database:

```shell
redis-cli -p 12000
```

## Alerting

Uptrace can monitor metrics using [alerting rules](https://uptrace.dev/get/alerting.html#alerting)
and send notifications via email/Slack/Telegram using AlertManager integration.

This example uses MailHog to test email notifications. Open
[http://localhost:8025](http://localhost:8025) to view available notifications and
[http://localhost:9093](http://localhost:9093) to view alerts.

See [documentation](https://uptrace.dev/get/alerting.html) for more details.
