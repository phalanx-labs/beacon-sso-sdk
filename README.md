# Beacon SSO SDK

面向 Gin 项目的轻量级 OAuth2/SSO SDK，提供 OAuth 配置初始化、登录回调、登出注销与业务逻辑能力，便于快速接入统一登录。

## 功能概览
- 自动读取环境变量并构建 OAuth2 配置
- 支持通过 `.well-known` 自动发现 endpoint（包含 introspection/revocation）
- 提供登录回调处理器与登出处理器
- 提供 Userinfo 与 Introspection 业务逻辑能力

## 快速开始

### 1) 安装
```bash
go get github.com/phalanx/beacon-sso-sdk
```

### 2) 初始化并挂载路由（推荐）
> SDK 当前基于 `xReg.Register` 的上下文注入流程工作，
> 不需要使用 `OAuthConfigStartup` 或 `OAuthContextHandlerFunc`。

```go
package main

import (
	"context"

	xConsts "github.com/bamboo-services/bamboo-base-go/context"
	xReg "github.com/bamboo-services/bamboo-base-go/register"
	xRegNode "github.com/bamboo-services/bamboo-base-go/register/node"
	xResult "github.com/bamboo-services/bamboo-base-go/result"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	bSdkRoute "github.com/phalanx/beacon-sso-sdk/route"
	bSdkStartup "github.com/phalanx/beacon-sso-sdk/startup"
	"gorm.io/gorm"
)

func main() {
	// 1) 初始化注册节点（数据库、Redis + SDK OAuth 配置）
	nodes := []xRegNode.RegNodeList{
		{Key: xConsts.DatabaseKey, Node: initDatabase},
		{Key: xConsts.RedisClientKey, Node: initRedis},
	}
	nodes = append(nodes, bSdkStartup.NewOAuthConfig()...)

	reg := xReg.Register(context.Background(), nodes)

	// 2) 注册业务路由
	reg.Serve.GET("/api/status", func(c *gin.Context) {
		xResult.SuccessHasData(c, "服务正常", struct{status string}{
			status: "running",
		})
	})

	// 3) 挂载 SDK OAuth 路由
	oauthRoute := bSdkRoute.NewOAuthRoute(reg.Init.Ctx)
	oauthRoute.OAuthRouter(reg.Serve.Group("/api"))

	// 4) 启动服务
	_ = reg.Serve.Run(":8080")
}

// 数据库初始化节点
func initDatabase(ctx context.Context) (any, error) {
	// TODO: 替换为你的数据库初始化逻辑
	var db *gorm.DB
	return db, nil
}

// Redis 初始化节点
func initRedis(ctx context.Context) (any, error) {
	// TODO: 替换为你的 Redis 初始化逻辑
	var rdb *redis.Client
	return rdb, nil
}
```

### 3) 默认路由
- 登录跳转：`GET /api/oauth/login`
- 登录回调：`GET /api/oauth/callback?code=...&state=...`
- 登出注销：`POST /api/oauth/logout`

## 环境变量
必填：
- `SSO_CLIENT_ID`
- `SSO_CLIENT_SECRET`
- `SSO_REDIRECT_URI`
- `SSO_ENDPOINT_AUTH_URI`
- `SSO_ENDPOINT_TOKEN_URI`
- `SSO_ENDPOINT_USERINFO_URI`
- `SSO_ENDPOINT_INTROSPECTION_URI`
- `SSO_ENDPOINT_REVOCATION_URI`

可选：
- `SSO_WELL_KNOWN_URI`（自动发现端点，支持 authorization/token/userinfo/introspection/revocation）
- `SSO_BUSINESS_CACHE`（业务逻辑缓存开关，支持 `true` / `false`，默认 `false`）

## 项目结构
- `handler/`: OAuth 回调与登出处理器
- `logic/`: OAuth 与业务逻辑（Userinfo/Introspection）
- `route/`: Gin 路由注册
- `middleware/`: 中间件
- `startup/`: OAuth 配置初始化
- `models/`: SDK 模型定义
- `constant/`: 环境变量与上下文键
- `utility/`: 上下文辅助方法

## 开发与测试
```bash
go mod download
go fmt ./...
go vet ./...
go test ./...
```
