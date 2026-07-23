# 后端架构

后端采用按业务功能组织的模块化单体。每个业务模块位于 `internal/<feature>`，并遵守单向依赖：HTTP handler 调用 service，service 调用 repository，只有 PostgreSQL repository 可以依赖 GORM。

## 模块模板

每个新业务模块都应包含以下文件：

- `entity.go`：领域实体及 GORM 表映射。
- `repository.go`：service 使用的数据访问接口。
- `postgres_repository.go`：PostgreSQL/GORM 实现、查询、事务和数据库约束映射。
- `service.go` 与 `errors.go`：用例、command 和领域错误；不得依赖 Gin、GORM 或 HTTP DTO。
- `http.go`：Gin handler、请求/响应结构、Swagger 注释，以及领域错误到 HTTP 响应的映射。
- `service_test.go`：使用手写 fake repository 的单元测试；数据库查询和 API 行为由 PostgreSQL 路由集成测试覆盖。

## 新增功能步骤

1. 建立模块实体、repository 接口与 PostgreSQL 实现。
2. 在 service 中定义 command、用例与领域错误，并补充 fake repository 单元测试。
3. 在 HTTP handler 定义 JSON 契约和完整的 Swagger 描述、成功及失败响应，保持 handler 只做绑定、映射和响应。
4. 在 `main.go` 装配 repository、service、handler；在 `router` 注册路由和认证边界。
5. 在 `internal/postgres/migrations` 中新增只向前执行的 Goose SQL migration。migration 是数据库结构变更的唯一来源，已发布的文件不得修改；GORM tag 只维护运行时字段映射。
6. 更新路由集成测试与 Swagger 契约测试，执行 `swag init` 并运行完整测试套件。

## 基础设施

`config` 只读取环境配置。`internal/postgres` 负责数据库连接，并在应用启动时按版本顺序执行尚未应用的 Goose migration。迁移失败会阻止应用启动，由人工处理；应用不提供向下迁移入口。`main.go` 是唯一组合根，不使用全局配置或全局数据库实例。

## 业务请求超时

`/v1` API 从进入路由组开始共享一份业务处理时间预算，覆盖参数绑定、业务逻辑、数据库调用和响应序列化。预算由 `BUSINESS_REQUEST_TIMEOUT` 配置，默认值为 `2s`，经 `config` 和 `main.go` 注入 `router`；service 和 repository 只接收并继续传递 request context。

超时采用协作式取消：middleware 在预算耗尽时返回 503、取消 request context，并丢弃后续响应写入；数据库及其他阻塞调用必须遵守 context。写请求在超时边界上可能已经提交，调用方不得将非幂等请求的 503 视为可无条件重试。该预算不替代 HTTP 服务端的请求读取、响应写入或 keep-alive 超时，本项目当前不对慢客户端传输设置额外限制。
