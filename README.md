# go-csust-planet

这是 [长理星球](https://github.com/zHElEARN/CSUSTPlanet) 的配套后端项目，基于 Go 语言开发，为移动端提供核心数据支持与推送服务。

## 功能特性

本后端项目主要提供以下功能支持：

- 统一身份认证：支持对接学校单点登录系统，处理用户登录及 JWT 鉴权。
- 电量实时监控：定时同步宿舍电量数据，并在用户设置得时间点通过 APNs 发送实时推送提醒。
- 校历与配置管理：提供学期校历、校园地图标注点、公告发布以及应用版本检查等配置信息。
- 数据同步：维护用户设备 Token，确保推送服务的准确触达。

## 构建

> [!IMPORTANT]
> **构建要求**：本项目需要连接 PostgreSQL 数据库，并且发送推送功能需要有效的 Apple Push Notification service (APNs) 证书或密钥

### 步骤

1. 克隆项目

   ```bash
   git clone https://github.com/zHElEARN/go-csust-planet.git
   cd go-csust-planet
   ```

2. 安装依赖

   本项目使用 Go Modules 管理依赖：

   ```bash
   go mod download
   ```

3. 项目配置

   复制环境变量模板并根据实际情况修改配置信息（如数据库连接、APNs 密钥路径等）：

   ```bash
   cp .env.template .env
   ```

   你需要确保 `.env` 文件中包含以下关键配置：
   - 数据库连接信息 (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` 等)
   - JWT 密钥 (`JWT_SECRET`)
   - APNs 凭据信息 (`APNS_TEAM_IDENTIFIER`, `APNS_KEY_IDENTIFIER`, `APNS_PRIVATE_KEY_PATH` 等)

4. 运行项目

   直接启动：

   ```bash
   go run main.go
   ```

   或者使用 [Air](https://github.com/cosmtrek/air) 进行热重载开发：

   ```bash
   air
   ```

## 许可证

本项目采用 **MIT License**。

这意味着：

- 您可以自由地商业化使用、复制、修改和分发本项目的源代码及其副本。
- 您只需在分发时保留原作者的版权声明和许可声明即可。
- 您可以将本项目代码集成到您的闭源或商业项目中，且无需公开您自己的源代码。
- 作者不对使用本项目产生的任何后果承担法律责任。

详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎并鼓励大家为 go-csust-planet 做出贡献，您可以 Fork 项目，进行修改并提交 Pull

如果您在使用过程中遇到问题，或对 go-csust-planet 有任何建议，也欢迎提交 Issue来告知我们！

同时，也可以通过邮箱联系我们：[personal@zhelearn.com](mailto:personal@zhelearn.com)

---

_免责声明: 本项目仅供学习与技术研究使用，请勿用于任何非法用途。在使用过程中请遵守学校相关网络安全规定。_
