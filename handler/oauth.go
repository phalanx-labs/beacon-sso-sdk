package bSdkHandler

import (
	"errors"
	"net/http"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// Login 处理 OAuth2 登录跳转请求
//
// 使用 OAuth2 SDK 生成授权跳转链接，并触发 302 重定向到 SSO 提供商的授权页面。
// 可选接受 query 参数 state，用于防止 CSRF。
//
// @Summary     [公开] OAuth2 登录跳转
// @Description 生成 OAuth2 授权链接并重定向到 SSO 提供商的授权页面
// @Tags        OAuth接口
// @Accept      json
// @Produce     json
// @Success     302  {string}  string  "重定向到 SSO 授权页面"
// @Router      /oauth/login [GET]
func (h *AuthHandler) Login(ctx *gin.Context) {
	h.log.Info(ctx, "Login - 处理登录跳转请求")

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
//
// @Summary     [公开] OAuth2 登录回调
// @Description 处理 SSO 提供商的回调，通过授权码换取访问令牌
// @Tags        OAuth接口
// @Accept      json
// @Produce     json
// @Param       code   query  string  true  "授权码"
// @Param       state  query  string  true  "状态参数（CSRF 防护）"
// @Success     200  {object}  xBase.BaseResponse{data=oauth2.Token}  "登录成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse  "用户拒绝授权或授权失败"
// @Router      /oauth/callback [GET]
func (h *AuthHandler) Callback(ctx *gin.Context) {
	h.log.Info(ctx, "Callback - 处理登录回调请求")

	// 检查是否产生错误返回
	if getErrString, exist := ctx.GetQuery("error"); exist {
		switch getErrString {
		case "access_denied":
			accessErr := errors.New("用户拒绝授权")
			_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(accessErr.Error()), false, accessErr))
			return
		default:
			unknownErr := errors.New("未知错误")
			_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(unknownErr.Error()), false, unknownErr))
			return
		}
	}

	var getToken *oauth2.Token
	getCode, codeExist := ctx.GetQuery("code")
	getState, stateExist := ctx.GetQuery("state")

	if codeExist && stateExist && getCode != "" && getState != "" {
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
	} else {
		getAT := xHttp.GetToken(ctx, xHttp.HeaderAuthorization)
		getRT := xHttp.GetToken(ctx, xHttp.HeaderRefreshToken)
		if getAT == "" || getRT == "" {
			_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要 code/state 参数或令牌参数", false, nil))
			return
		}

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
	}

	xResult.SuccessHasData(ctx, "登录成功", getToken)
}

// Logout 处理 OAuth2 登出请求
//
// 该处理器会根据请求头中的令牌调用 revocation endpoint 进行注销。
// 默认注销 access token；当 query 参数 token_type=refresh_token 时注销刷新令牌。
//
// @Summary     [用户] OAuth2 登出
// @Description 注销访问令牌或刷新令牌，调用 revocation endpoint 进行注销
// @Tags        OAuth接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string  true   "Bearer Access Token 或 Refresh Token"
// @Param       token_type     query   string  false  "令牌类型"  Enums(access_token, refresh_token)  default(access_token)
// @Success     200  {object}  xBase.BaseResponse  "登出成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Router      /oauth/logout [POST]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	h.log.Info(ctx, "Logout - 处理登出请求")

	tokenType := ctx.DefaultQuery("token_type", "access_token")
	var token string
	switch tokenType {
	case "refresh_token":
		token = xHttp.GetToken(ctx, xHttp.HeaderRefreshToken)
	default:
		tokenType = "access_token"
		token = xHttp.GetToken(ctx, xHttp.HeaderAuthorization)
	}

	if token == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要令牌参数", false, nil))
		return
	}

	xErr := h.service.oauthLogic.Logout(ctx, tokenType, token)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "登出成功")
}
