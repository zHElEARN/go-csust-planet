# go-csust-planet

这是 [云岭星球](https://github.com/zHElEARN/CSUSTPlanet) 的配套后端项目，基于 Go 和 Svelte 开发，为移动端提供静态配置数据支持。

## 功能特性

本后端项目主要提供以下功能支持：

- 校历与配置管理：提供学期校历、校园地图标注点、公告发布以及应用版本检查等配置信息。
- 后台管理：提供 Web 端管理页面与配套 API。

后端模块划分与新增功能规范见 [架构文档](docs/architecture.md)。

## 构建

> [!IMPORTANT]
> **构建要求**：本项目需要连接 PostgreSQL 数据库。

### 步骤

1. 克隆项目

   ```bash
   git clone https://github.com/zHElEARN/go-csust-planet.git
   cd go-csust-planet
   ```

2. 安装依赖

   本项目使用 Go Modules 管理后端依赖：

   ```bash
   go mod download
   ```

   使用 pnpm 管理后台管理系统依赖：

   ```bash
   cd admin
   pnpm install
   ```

3. 项目配置

   复制环境变量模板并根据实际情况修改配置信息：

   ```bash
   cp .env.template .env
   ```

   你需要确保 `.env` 文件中包含以下关键配置：
   - 数据库连接信息 (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` 等)
   - 后台管理 Token (`ADMIN_BEARER_TOKEN`)
   - Swagger 文档访问密码 (`SWAGGER_PASSWORD`)

   Swagger 文档会始终保持开启，访问地址为 `/swagger/index.html`，并使用 Basic Auth 保护：
   - 用户名固定为 `swagger`
   - 密码来自环境变量 `SWAGGER_PASSWORD`

4. 运行项目

   直接启动：

   ```bash
   go run main.go
   ```

   或者使用 [Air](https://github.com/cosmtrek/air) 进行热重载开发：

   ```bash
   air
   ```

   同时启动后台管理系统：

   ```bash
   cd admin
   pnpm dev
   ```

## 部署

配套云岭星球应用时，需要同时维护 Debug 和 Release 的后端，参见示例 [docker-compose.yml](docker-compose.yml) 和 [scripts/deploy-remote.sh](scripts/deploy-remote.sh) 文件。

## 许可证

本项目采用 **MIT License**。

这意味着：

- 您可以自由地商业化使用、复制、修改和分发本项目的源代码及其副本。
- 您只需在分发时保留原作者的版权声明和许可声明即可。
- 您可以将本项目代码集成到您的闭源或商业项目中，且无需公开您自己的源代码。
- 作者不对使用本项目产生的任何后果承担法律责任。

详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎大家为 go-csust-planet 做出贡献，您可以 Fork 项目，进行修改并提交 Pull Request。

如果您在使用过程中遇到问题，或对 go-csust-planet 有任何建议，也欢迎提交 Issue 来告知我们！

---

_免责声明: 本项目仅供学习与技术研究使用，请勿用于任何非法用途。在使用过程中请遵守学校相关网络安全规定。_
