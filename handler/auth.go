package bSdkHandler

import (
	"log/slog"

	xError "github.com/bamboo-services/bamboo-base-go/error"
	xResult "github.com/bamboo-services/bamboo-base-go/result"
	"github.com/gin-gonic/gin"
	bSdkUtil "github.com/phalanx/beacon-sso-sdk/utility"
)

// Callback 处理 OAuth2 登录回调请求
//
// 接收来自外部 SSO 提供商的回调，通过授权码换取访问令牌，并返回登录结果。
// 该处理器会自动从环境变量中读取 SSO 客户端凭证，并验证请求中携带的 code 和 state 参数。
func (h *AuthHandler) Callback(ctx *gin.Context) {
	h.log.Info(ctx, "AuthHandler|Callback - 处理登录回调请求")

	// 获取回调代码参数
	getCode, exist := ctx.GetQuery("code")
	if exist && getCode == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要代码参数", false, nil))
		return
	}
	getState, exist := ctx.GetQuery("state")
	if exist && getState == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要状态参数", false, nil))
		return
	}
	h.log.Info(ctx, "获取参数", slog.String("code", getCode), slog.String("state", getState))

	// 换取实际令牌
	getToken, oAuthErr := bSdkUtil.GetOAuthConfig(ctx).Exchange(ctx, getCode)
	if oAuthErr != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, "未登录", false, oAuthErr))
		return
	}
	xResult.SuccessHasData(ctx, "登录成功", getToken)
}
