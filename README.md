# Beacon SSO SDK

面向 Gin 项目的轻量级 OAuth2/SSO SDK，提供配置引导、回调处理与上下文注入能力，便于快速接入统一登录。

## 功能概览
- 自动读取环境变量并构建 OAuth2 配置
- Gin 中间件注入 OAuth 配置与 Userinfo URI
- 提供登录回调处理器（授权码换取令牌）

## 快速开始

### 1) 安装
```bash
go get github.com/phalanx/beacon-sso-sdk
```

### 2) 初始化并挂载路由
```go
package main

import (
    "context"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    bSdkRoute "github.com/phalanx/beacon-sso-sdk/route"
    bSdkStartup "github.com/phalanx/beacon-sso-sdk/startup"
    "gorm.io/gorm"
)

func main() {
    ctx := context.Background()
    config, userinfoURI := bSdkStartup.OAuthConfigStartup(ctx)

    r := gin.Default()
    r.Use(bSdkStartup.OAuthContextHandlerFunc(config, userinfoURI))

    // TODO: 替换为你的 db / rdb 实例
    var db *gorm.DB
    var rdb *redis.Client

    oauthRoute := bSdkRoute.NewOAuthRoute(ctx, db, rdb)
    oauthRoute.OAuthRouter(r.Group("/api"))

    _ = r.Run(":8080")
}
```

### 3) 回调地址
- 默认注册路径：`GET /api/oauth/callback?code=...&state=...`

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

## 项目结构
- `handler/`: OAuth 回调处理器
- `route/`: Gin 路由注册
- `middleware/`: 中间件（预留认证校验）
- `startup/`: OAuth 配置初始化
- `constant/`: 环境变量与上下文键
- `utility/`: 上下文辅助方法

## 开发与测试
```bash
go mod download
go fmt ./...
go vet ./...
go test ./...
```
