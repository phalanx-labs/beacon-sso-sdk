package bSdkHandler

import (
	"errors"
	"net/http"

	xError "github.com/bamboo-services/bamboo-base-go/error"
	xHttp "github.com/bamboo-services/bamboo-base-go/http"
	xResult "github.com/bamboo-services/bamboo-base-go/result"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// Login 处理 OAuth2 登录跳转请求
//
// 使用 OAuth2 SDK 生成授权跳转链接，并触发 302 重定向到 SSO 提供商的授权页面。
// 可选接受 query 参数 state，用于防止 CSRF。
func (h *AuthHandler) Login(ctx *gin.Context) {
	h.log.Info(ctx, "AuthHandler|Login - 处理登录跳转请求")

	oAuth, xErr := h.service.oauthLogic.Create(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	authURL, xErr := h.service.oauthLogic.BuildURL(ctx, oAuth)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Redirect(http.StatusFound, authURL)
}

// Callback 处理 OAuth2 登录回调请求
//
// 接收来自外部 SSO 提供商的回调，通过授权码换取访问令牌，并返回登录结果。
// 该处理器会自动从环境变量中读取 SSO 客户端凭证，并验证请求中携带的 code 和 state 参数。
func (h *AuthHandler) Callback(ctx *gin.Context) {
	h.log.Info(ctx, "AuthHandler|Callback - 处理登录回调请求")

	// 检查是否产生错误返回
	if getErrString, exist := ctx.GetQuery("error"); exist {
		switch getErrString {
		case "access_denied":
			_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, "登录失败", false, errors.New("用户拒绝授权")))
			return
		default:
			_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, "登录失败", false, errors.New("未知错误")))
			return
		}
	}

	// 获取回调代码参数
	getCode, exist := ctx.GetQuery("code")
	if !exist || getCode == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要代码参数", false, nil))
		return
	}
	getState, exist := ctx.GetQuery("state")
	if !exist || getState == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要状态参数", false, nil))
		return
	}

	// 检查登录态
	getAT := xHttp.GetToken(ctx, xHttp.HeaderAccessToken)
	getRT := xHttp.GetToken(ctx, xHttp.HeaderRefreshToken)

	var getToken *oauth2.Token
	if getAT != "" && getRT != "" {
		// 刷新类型
		cacheToken, xErr := h.service.oauthLogic.GetToken(ctx, getAT)
		if xErr != nil {
			_ = ctx.Error(xErr)
			return
		}
		tokenSource, xErr := h.service.oauthLogic.TokenSource(ctx, cacheToken, getRT)
		if xErr != nil {
			_ = ctx.Error(xErr)
			return
		}
		getToken = tokenSource
	} else {
		// 授权类型
		oAuth, xErr := h.service.oauthLogic.Verify(ctx, getState)
		if xErr != nil {
			_ = ctx.Error(xErr)
			return
		}
		token, xErr := h.service.oauthLogic.Exchange(ctx, getCode, oAuth.Verifier)
		if xErr != nil {
			_ = ctx.Error(xErr)
			return
		}
		getToken = token
	}

	xResult.SuccessHasData(ctx, "登录成功", getToken)
}
