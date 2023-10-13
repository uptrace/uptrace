选择语言: [English](README.md) | **简体中文**

# 开源 APM：OpenTelemetry 追踪、指标和日志

[![build workflow](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![Chat](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)
[![Slack](https://img.shields.io/badge/slack-uptrace.svg?logo=slack)](https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw)

Uptrace 是一种开源的 APM 工具，支持分布式跟踪、指标和日志。您使用它可以监控应用程序并设置自动报警，
可以通过电子邮件、Slack、Telegram 等方式接收通知。

Uptrace 使用 OpenTelemetry 收集数据并使用 ClickHouse 数据库存储数据。ClickHouse 是 Uptrace 的唯一依
赖。

**功能**

- 跨度(span) / 日志分组。
- 类似 SQL 风格的 span 统计。
- 类似 Promql 风格的统计和监控指标。
- 可以使用电子邮件、Slack、Telegram 等接收报警信息。
- 预绘制指标图表。
- 支持 YAML 配置多个用户和项目。
- 支持 Keycloak、Cloudflare、Google Cloud 等方式单点登录(SSO)。

**亮点**

- 统一风格的用于跟踪、指标和日志的 UI 界面。
- 高性能：单核每秒处理超过 10K span。
- 使用 ZSTD 压缩算法进行磁盘压缩，1KB span 可以压缩到 40 字节以内。
- 支持 S3 存储，能够自动将冷数据上传到 S3 或 HDD 等进行存储。
- 通过电子邮件、Slack、Telegram 等方式自动发出报警信息。

**System overview**

![System overview](./example/docker/images/home.png)

**Faceted filters**

![Faceted filters](./example/docker/images/facets.png)

**Metrics**

![Metrics](./example/docker/images/metrics.png)

**Alerts**

![Alerts](./example/docker/images/alerts.png)

## 快速开始

你可以通过 [docker-compose](example/docker) 示例来快速运行 Uptrace，也可以和大家一起体验
[在线 Demo](https://app.uptrace.dev/play)（无需登录）。

在 [入门指南](https://uptrace.dev/get/get-started.html) 有详细安装步奏，可以下载 Go 二进制文件或
DEB/RPM 包安装 Uptrace，额外仅需要安装 ClickHouse 数据库。

## 帮助

可以通过 [Telegram](https://t.me/uptrace),
[Slack](https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw) 或
GitHub [start a discussion](https://github.com/uptrace/uptrace/discussions) 寻求帮助。

## 贡献

请参阅 [为 Uptrace 做贡献](https://uptrace.dev/get/contributing.html)。

<a href="https://github.com/uptrace/uptrace/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=uptrace/uptrace" />
</a>
