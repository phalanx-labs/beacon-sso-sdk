package bSdkMiddle

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/error"
	xHttp "github.com/bamboo-services/bamboo-base-go/http"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	xResult "github.com/bamboo-services/bamboo-base-go/result"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/utility/ctxutil"
	"github.com/gin-gonic/gin"
	bSdkLogic "github.com/phalanx/beacon-sso-sdk/logic"
)

// CheckAuth 检查用户身份认证信息
//
// 本函数是一个中间件工厂，用于生成 Gin 的 HandlerFunc。
// 它利用提供的 `context.Context` 初始化数据库与 Redis 连接，
// 进而构建 OAuth 逻辑层。返回的中间件函数会执行以下逻辑：
//
//  1. 从请求头的 `Authorization` 字段提取访问令牌。
//  2. 调用 `OAuthLogic` 验证令牌的有效性及过期时间。
//  3. 若验证通过，调用 `ctx.Next()` 放行请求；否则中断请求并返回错误。
//
// 参数说明:
//   - ctx: 上下文环境，必须包含通过 `xCtxUtil` 注入的 DB (*gorm.DB) 和 RDB (*redis.Client)。
//
// 返回值:
//   - gin.HandlerFunc: 配置好的 Gin 中间件处理函数。
func CheckAuth(ctx context.Context) gin.HandlerFunc {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	oAuthLogic := bSdkLogic.NewOAuth(db, rdb)

	return func(ctx *gin.Context) {
		log := xLog.WithName(xLog.NamedMIDE)
		log.Info(ctx, "CheckAuth - 检查用户身份认证信息")

		// 获取用户身份令牌
		getAT := xHttp.GetToken(ctx, xHttp.HeaderAccessToken)
		if getAT == "" {
			xResult.AbortError(ctx, xError.ParameterEmpty, "需要访问令牌参数", nil)
			return
		}

		isValid, xErr := oAuthLogic.VerifyExpiry(ctx, getAT)
		if xErr != nil {
			xResult.AbortError(ctx, xErr.ErrorCode, xErr.ErrorMessage, xErr.Data)
			return
		}

		// 验证是否过期
		if !isValid {
			xResult.AbortError(ctx, xError.TokenExpired, "访问令牌已过期", nil)
			return
		}

		// 校验通过
		ctx.Next()
	}
}
