package bSdkStartup

import (
	"context"
	"fmt"
	"net/http"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	"golang.org/x/oauth2"
)

// OAuthContextHandlerFunc 返回一个 Gin 中间件处理函数，用于将 OAuth2 配置和用户信息 URI 注入到请求上下文中。
//
// 该函数是一个中间件工厂，它接收 OAuth2 配置和用户信息端点 URI，
// 并返回一个闭包函数。这个返回的函数会将这些信息存储到 `gin.Context` 中，
// 以供后续的处理函数（如登录回调或用户信息获取）使用。
//
// 参数:
//   - config: OAuth2 的配置对象，包含 ClientID、ClientSecret、Endpoint 等必要信息。
//   - userinfoURI: 用于获取用户资源详细信息的 URI 地址。
//
// 内部实现:
//   - 使用键 `bSdkConst.CtxOAuthConfig` 将 config 存入上下文。
//   - 使用键 `bSdkConst.CtxOAuthUserinfoURI` 将 userinfoURI 存入上下文。
//   - 调用 `ctx.Next()` 继续处理请求链。
func OAuthContextHandlerFunc(config *oauth2.Config, userinfoURI string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.Set(bSdkConst.CtxOAuthConfig, config)
		ctx.Set(bSdkConst.CtxOAuthUserinfoURI, userinfoURI)

		ctx.Next()
	}
}

// OAuthConfigStartup 初始化并构建 OAuth2 配置对象及用户信息端点。
//
// 该函数优先检查是否存在 SSO_WELL_KNOWN_URI 环境变量。如果存在，会自动请求该地址
// 以获取授权和令牌端点的元数据。随后，从环境变量中读取必要的客户端配置（如 ClientID、
// ClientSecret、RedirectURI 等），构建标准的 `oauth2.Config` 结构体。
//
// 此函数通常在应用启动或中间件初始化时调用，用于统一管理 SSO 登录所需的参数。
//
// 参数说明:
//   - ctx: 上下文对象，用于传递日志上下文和请求取消信号。
//
// 返回值:
//   - *oauth2.Config: 初始化完成的 OAuth2 配置对象，包含客户端ID、密钥及端点地址。
//   - string: 用户信息端点的 URI 字符串，用于后续获取用户详情。
//
// 注意: 此函数会在关键配置（如 ClientID 或 Secret）缺失，或获取 Well-Known 元数据
// 失败时触发 panic。请确保相关环境变量已正确设置。
func OAuthConfigStartup(ctx context.Context) (*oauth2.Config, string) {
	xLog.WithName(xLog.NamedINIT).Info(ctx, "初始化 OAuth 配置")

	var (
		wkAuthURI     string // well-known 登录端点
		wkTokenURI    string // well-known 令牌端点
		wkUserinfoURI string // well-known 获取用户信息端点
	)
	if getWellKnown := xEnv.GetEnvString(bSdkConst.EnvSsoWellKnownURI, ""); getWellKnown != "" {
		xLog.WithName(xLog.NamedINIT).Info(ctx, "使用 SSO_WELL_KNOWN_URI 环境变量配置 OAuth2 Endpoint")

		client := resty.New()
		wellKnown := make(map[string]any)
		resp, err := client.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json").
			SetResult(&wellKnown).
			Get(getWellKnown)
		if err != nil {
			panic(fmt.Sprintf("无法获取 SSO_WELL_KNOWN_URI 环境变量配置的元数据: %v", err))
		}

		if resp.StatusCode() != http.StatusOK {
			panic(fmt.Sprintf("SSO_WELL_KNOWN_URI 返回非成功状态码: %d", resp.StatusCode()))
		}

		wkAuthURI = wellKnown["authorization_endpoint"].(string)
		wkTokenURI = wellKnown["token_endpoint"].(string)
		wkUserinfoURI = wellKnown["userinfo_endpoint"].(string)
	}

	// 获取环境变量
	clientID := xEnv.GetEnvString(bSdkConst.EnvSsoClientID, "")
	clientSecret := xEnv.GetEnvString(bSdkConst.EnvSsoClientSecret, "")
	clientRedirectURI := xEnv.GetEnvString(bSdkConst.EnvSsoRedirectURI, "")
	authURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointAuthURI, wkAuthURI)
	tokenURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointTokenURI, wkTokenURI)
	userinfoURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointUserinfoURI, wkUserinfoURI)

	if clientID == "" || clientSecret == "" || clientRedirectURI == "" || authURI == "" || tokenURI == "" || userinfoURI == "" {
		panic("SSO 客户端配置缺失")
	}

	// 调用跳转
	oAuthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  clientRedirectURI,
		Scopes:       []string{"openid", "profile", "email", "phone"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURI,
			TokenURL: tokenURI,
		},
	}

	return oAuthConfig, userinfoURI
}
