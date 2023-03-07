[**English**](README.md) | 简体中文

# 开源 APM：OpenTelemetry 跟踪、指标和日志

[![构建工作流](https://github.com/uptrace/uptrace/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/uptrace/uptrace/actions)
[![聊天](https://img.shields.io/badge/-telegram-red?color=white&logo=telegram&logoColor=black)](https://t.me/uptrace)

Uptrace 是一种开源 APM 工具，支持分布式跟踪、指标和日志。您可以使用它可以监控应用程序并设置自动警报
以通过电子邮件、Slack 接收通知，电报等等。

Uptrace 使用 OpenTelemetry 收集数据并使用 ClickHouse 数据库存储数据。ClickHouse 是只有依赖。

**特征**：

- 跨度/日志分组。
- 类似 SQL 的查询语言来聚合跨度。
- 类似 Promql 的语言来聚合和监控指标。
- 使用 AlertManager 的电子邮件/Slack/Telegram 通知。
- 预建指标仪表板。
- 通过 YAML 配置的多个用户/项目。
- 使用 OpenID Connect 的单点登录 (SSO)：Keycloak、Cloudflare、Google Cloud 等。

**为什么是 Uptrace？**

- 用于跟踪、指标和日志的单一 UI。
- 高效摄取：单核上超过 10K 跨度/秒。
- 使用 ZSTD 进行出色的磁盘压缩，例如，1KB 跨度可以压缩到 <40 字节。
- S3 存储支持，能够自动将冷数据上传到类似 S3 的存储或 HDD。
- 通过电子邮件、Slack、Telegram 等方式自动发出通知警报。

![Uptrace 主页](./example/docker/images/home.png)

![多面过滤器](./example/docker/images/facets.png)

![类似普罗米修斯的指标](./example/docker/images/metrics.png)

＃＃ 快速开始

只需几分钟的时间，您就可以通过运行一个命令来决定是否需要 Uptrace
[docker-compose 示例](示例/docker)。也可以和大众一起玩
[Uptrace 演示](https://app.uptrace.dev/play)（无需登录）。

然后按照 [入门指南](https://uptrace.dev/get/get-started.html) 正确设置通过下载 Go 二进制文件或安装
DEB/RPM 包进行 Uptrace。你只需要一个 ClickHouse 数据库开始使用 Uptrace。

＃＃ 帮助

有问题吗？通过 [Telegram](https://t.me/uptrace) 或
[开始讨论](https://github.com/uptrace/uptrace/discussions) 在 GitHub 上。

＃＃ 常问问题

**许可证是什么？**

Business Source [License](LICENSE) 与 Apache 2.0 相同，唯一的例外是您不能使用该代码创建云服务，或者
换句话说，不能将产品转售给他人。

BSL 被 MariaDB、Sentry、CockroachDB、Couchbase 和许多其他公司采用。在大多数情况下，它是一个比例如
AGPL 更宽松的许可证，因为它允许您对代码。

三年后，该代码也可在 Apache 2.0 许可下使用。您可以了解更多关于 BSL
[此处](https://mariadb.com/bsl-faq-adopting/)。

**我可以使用 Uptrace 监控商业或生产级应用程序吗？**

是的，您可以使用 Uptrace 来监控**您的**应用程序并为员工提供对 Uptrace 应用程序没有任何限制。

**为什么要获得 BSL 许可**？

我们选择许可证的目的是允许用户使用 Uptrace 监控他们的应用程序，但禁止其他公司使用该代码创建云服务。

我们自己提供[监控服务](https://uptrace.dev/)，以便通过我们的工作获利，持续发展努力。

**你是开源的吗？**

从技术上讲，BSL 许可证被归类为源代码可用，但我们继续使用该术语在源代码公开的基础上开源。

现有的 SEO 实践并没有给我们留下太多选择，而我们的竞争对手或多或少也这样做。

**数据库架构稳定吗？**

是的，但我们仍在对数据库模式进行更改并计划切换到
[ClickHouse 动态子列](https://github.com/ClickHouse/ClickHouse/pull/23932) 当那特点是
[稳定](https://github.com/ClickHouse/ClickHouse/issues?q=is%3Aissue+is%3Aopen+label%3Acomp-type-object)
足够的。

## 贡献

请参阅 [为 Uptrace 做贡献](https://uptrace.dev/get/contributing.html)。

<a href="https://github.com/uptrace/uptrace/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=uptrace/uptrace" />
</a>
