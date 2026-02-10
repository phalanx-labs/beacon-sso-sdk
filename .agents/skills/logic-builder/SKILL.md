---
name: logic-builder
description: 面向 Beacon SSO SDK 的业务逻辑开发规范技能。用于在 `logic/` 新建逻辑组件、扩展现有逻辑方法，并联动 `handler/`、`repository/`、`repository/cache/`、`models/`、`middleware/`、`route/` 与 `startup/` 完成端到端改造。凡是涉及 OAuth/SSO 流程新增、state/token 校验调整、缓存读写策略变更、错误码和日志治理、或逻辑层重构时都应使用本技能。
---

# Logic Builder

## 目标

将 Beacon SSO 的逻辑开发流程标准化，优先保证：
- 分层清晰（`handler -> logic -> repository -> cache/model`）。
- 错误一致（统一返回 `*xError.Error`）。
- 日志可追踪（`实体|方法 - 动作`）。
- 变更可验证（最少执行 `go fmt ./... && go vet ./... && go test ./...`）。

## 仓库逻辑约束

在实现任何逻辑前，先遵守以下固定约束：

1. 保持包命名风格
- `logic` 包名使用 `bSdkLogic`。
- `repository` 包名使用 `bSdkRepo`。
- `repository/cache` 包名使用 `bSdkCache`。

2. 保持构造器与依赖注入风格
- 在 `logic/*.go` 定义 `type XxxLogic struct`。
- 构造器固定为 `NewXxx(db *gorm.DB, rdb *redis.Client) *XxxLogic`。
- 在构造器中注入日志与仓储依赖。

3. 保持方法签名与错误风格
- 业务方法优先使用 `ctx *gin.Context`。
- 可预期业务错误统一返回 `*xError.Error`，不要在 logic 层直接 panic。
- 入参校验失败优先使用 `xError.ParameterEmpty` 或语义对应错误码。

4. 保持日志风格
- 入口日志统一 `Info`。
- 文案格式统一 `XxxLogic|Method - 中文动作描述`。
- 非阻断失败（例如缓存清理失败）记录 `Warn`，不要中断主流程。

5. 保持缓存与时间语义
- 缓存结构使用 `models/*`，字段带 `redis`/`json` tag。
- 过期时间统一使用 `time.RFC3339` 序列化/反序列化。
- Redis 键通过 `constant/cache.go` 的 `RedisKey` 构造，不要硬编码。

## 工作流决策

先判断任务类型，再按对应流程执行：

- 新增领域能力（例如新增 userinfo 拉取、登出、租户切换）→ 执行“新建 Logic 流程”。
- 在现有 `OAuthLogic` 上加方法/改行为（例如 token 校验策略）→ 执行“扩展 Logic 流程”。
- 仅调整路由暴露方式而不改业务规则 → 只改 `handler/route`，避免污染 logic。

## 新建 Logic 流程

1. 先定义领域边界
- 明确一个核心职责，不在一个 Logic 中混入多个正交能力。
- 先写出输入、输出、依赖（DB/Redis/第三方 API）。

2. 新建逻辑骨架
- 在 `logic/` 创建 `<domain>.go`。
- 定义 `type XxxLogic struct { db; rdb; log; repo... }`。
- 添加 `NewXxx(...)` 构造器，并注入所需 repo。

3. 下沉数据访问
- 在 `repository/` 增加对应 repo（必要时新增 `repository/cache/`）。
- 在 repo 层把底层错误映射为 `*xError.Error`，logic 层只编排流程。

4. 暴露到 handler
- 在 `handler/handler.go` 的 `service` 结构体注册新 logic。
- 在 handler 中调用 logic 方法，并用 `_ = ctx.Error(xErr)` 交给统一错误链路。

5. 接入路由与中间件
- 在 `route/` 注册新 handler 方法。
- 需要鉴权时复用或扩展 `middleware/`，避免在 handler 重复鉴权逻辑。

6. 校验依赖上下文
- 若新增上下文依赖，更新 `startup/` 注册节点与 `utility/context.go` 读取方法。
- 新增环境变量时同步更新 `constant/environment.go`。

## 扩展 Logic 流程

1. 先锁定变更点
- 标注现有方法输入输出是否变更。
- 优先保持旧方法签名稳定，新增能力尽量通过新方法承载。

2. 保持行为兼容
- 兼容已有成功路径。
- 对新增失败路径补充精确错误码与错误消息。

3. 强化非功能约束
- 补齐日志字段（至少带核心标识，如 `state`/`token` 的可审计片段）。
- 区分阻断错误与可降级错误；仅阻断关键链路。

4. 回收副作用
- 对一次性凭证（如 state）在成功后执行删除。
- 删除失败记录告警，不影响主结果。

## 逻辑实现模板

```go
func (l *XxxLogic) Execute(ctx *gin.Context, req string) (*Result, *xError.Error) {
	l.log.Info(ctx, "XxxLogic|Execute - 执行业务动作")

	if req == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "参数为空", false, nil)
	}

	data, xErr := l.repo.Get(ctx, req)
	if xErr != nil {
		return nil, xErr
	}

	if warnErr := l.repo.Delete(ctx, req); warnErr != nil {
		l.log.Warn(ctx, "XxxLogic|Execute - 清理副作用失败")
	}

	return data, nil
}
```

## 交付前检查

每次完成逻辑改造后，按顺序执行：

1. 代码风格检查
- `go fmt ./...`

2. 静态检查
- `go vet ./...`

3. 测试检查
- `go test ./...`

4. 变更完整性核对
- 确认是否同步更新了 `handler/`、`route/`、`repository/`、`models/`、`constant/`、`startup/`。
- 确认错误码、日志文案、缓存键和 TTL 与现有风格一致。
