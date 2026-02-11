package bSdkStartup

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	xRegNode "github.com/bamboo-services/bamboo-base-go/register/node"
	"github.com/go-resty/resty/v2"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	"golang.org/x/oauth2"
)

// NewOAuthConfig 初始化并返回 OAuth2 及重定向 URI 的依赖注入节点列表。
//
// 该函数聚合了 `oAuthConfig` (包含 ClientID、Endpoint 等核心配置) 和
// `oAuthRedirectURI` (包含重定向地址) 的注册节点，用于在应用启动时
// 批量注册 SSO 相关的上下文键值。
//
// 返回值:
//   - 包含所有已定义 OAuth 注册节点的切片。
func NewOAuthConfig() []xRegNode.RegNodeList {
	regNode := make([]xRegNode.RegNodeList, 0)
	regNode = append(regNode, oAuthConfig())
	regNode = append(regNode, oAuthRedirectURI())
	return regNode
}

// oAuthConfig 初始化 OAuth2 配置并注册依赖项。
//
// 该函数负责从环境变量或 Well-Known 元数据端点加载 OAuth2 客户端所需的配置信息，
// 并将其封装为 `oauth2.Config` 对象注册到依赖注入容器中。
//
// 配置加载逻辑优先级：
//  1. 如果设置了 `SSO_WELL_KNOWN_URI` 环境变量，函数将发起 HTTP GET 请求获取
//     OpenID Connect 的元数据，从而自动解析 Authorization、Token、Userinfo、Introspection 与 Revocation 端点。
//  2. 否则，将尝试从 `SSO_ENDPOINT_*` 相关的环境变量读取端点地址。
//
// 函数会校验必要的配置（如 ClientID, Secret, RedirectURL 等），如果缺失则会触发 Panic。
//
// 返回的注册节点列表将包含以下键：
//   - `CtxOAuthConfig`: 存储初始化后的 `*oauth2.Config` 实例。
//   - `CtxOAuthUserinfoURI`: 存储用于获取用户信息的端点地址字符串。
//
// 参数:
//   - ctx: 用于日志记录和 HTTP 请求的上下文。
//
// 返回值:
//   - func(ctx context.Context) (any, error): 返回一个函数，该函数在调用时返回 OAuth2 配置和元数据 URI 的注册节点列表。
func oAuthConfig() xRegNode.RegNodeList {
	return xRegNode.RegNodeList{
		Key: bSdkConst.CtxOAuthConfig,
		Node: func(ctx context.Context) (any, error) {
			log := xLog.WithName(xLog.NamedINIT)
			log.Info(ctx, "初始化 OAuth 配置")

			var (
				wkAuthURI          string // well-known 登录端点
				wkTokenURI         string // well-known 令牌端点
				wkUserinfoURI      string // well-known 获取用户信息端点
				wkIntrospectionURI string // well-known 令牌自省端点
				wkRevocationURI    string // well-known 令牌注销端点
			)
			if getWellKnown := xEnv.GetEnvString(bSdkConst.EnvSsoWellKnownURI, ""); getWellKnown != "" {
				log.Info(ctx, "使用 SSO_WELL_KNOWN_URI 环境变量配置 OAuth2 Endpoint")

				client := resty.New()
				wellKnown := make(map[string]any)
				resp, err := client.R().
					SetContext(ctx).
					SetHeader("Accept", "application/json").
					SetResult(&wellKnown).
					Get(getWellKnown)

				if err != nil {
					return nil, fmt.Errorf("无法获取 SSO_WELL_KNOWN_URI 环境变量配置的元数据: %v", err)
				}
				if resp.StatusCode() != http.StatusOK {
					return nil, fmt.Errorf("SSO_WELL_KNOWN_URI 返回非成功状态码: %d", resp.StatusCode())
				}

				wkAuthURI = readWellKnownURI(wellKnown, "authorization_endpoint")
				wkTokenURI = readWellKnownURI(wellKnown, "token_endpoint")
				wkUserinfoURI = readWellKnownURI(wellKnown, "userinfo_endpoint")
				wkIntrospectionURI = readWellKnownURI(wellKnown, "introspection_endpoint")
				wkRevocationURI = readWellKnownURI(wellKnown, "revocation_endpoint")
			}

			// 获取环境变量
			clientID := xEnv.GetEnvString(bSdkConst.EnvSsoClientID, "")
			clientSecret := xEnv.GetEnvString(bSdkConst.EnvSsoClientSecret, "")
			clientRedirectURI := xEnv.GetEnvString(bSdkConst.EnvSsoRedirectURI, "")
			authURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointAuthURI, wkAuthURI)
			tokenURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointTokenURI, wkTokenURI)
			userinfoURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointUserinfoURI, wkUserinfoURI)
			introspectionURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointIntrospectionURI, wkIntrospectionURI)
			revocationURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointRevocationURI, wkRevocationURI)

			if clientID == "" || clientSecret == "" || clientRedirectURI == "" || authURI == "" || tokenURI == "" || userinfoURI == "" || introspectionURI == "" || revocationURI == "" {
				xLog.Panic(ctx, "SSO 客户端配置缺失",
					slog.String("client_id", clientID),
					slog.String("client_secret", clientSecret),
					slog.String("redirect_uri", clientRedirectURI),
					slog.String("auth_uri", authURI),
					slog.String("token_uri", tokenURI),
					slog.String("userinfo_uri", userinfoURI),
					slog.String("introspection_uri", introspectionURI),
					slog.String("revocation_uri", revocationURI),
				)
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

			// 同时设置环境变量，供其他逻辑层使用
			if envErr := xEnv.SetEnv(bSdkConst.EnvSsoEndpointUserinfoURI, userinfoURI); envErr != nil {
				return nil, fmt.Errorf("设置环境变量失败: %v", envErr)
			}
			if envErr := xEnv.SetEnv(bSdkConst.EnvSsoEndpointIntrospectionURI, introspectionURI); envErr != nil {
				return nil, fmt.Errorf("设置环境变量失败: %v", envErr)
			}
			if envErr := xEnv.SetEnv(bSdkConst.EnvSsoEndpointRevocationURI, revocationURI); envErr != nil {
				return nil, fmt.Errorf("设置环境变量失败: %v", envErr)
			}
			if envErr := xEnv.SetEnv(bSdkConst.EnvSsoRedirectURI, clientRedirectURI); envErr != nil {
				return nil, fmt.Errorf("设置环境变量失败: %v", envErr)
			}

			return oAuthConfig, nil
		},
	}
}

// oAuthRedirectURI 初始化并注册 OAuth2 重定向 URI 的依赖注入节点。
//
// 该函数从环境变量 `SSO_REDIRECT_URI` 中读取配置的回调地址，
// 并将其注册到依赖注入容器中，以便在 OAuth2 认证流程中使用。
//
// 注册的上下文键为 `CtxOAuthUserinfoURI`。
func oAuthRedirectURI() xRegNode.RegNodeList {
	return xRegNode.RegNodeList{
		Key: bSdkConst.CtxOAuthUserinfoURI,
		Node: func(ctx context.Context) (any, error) {
			return xEnv.GetEnvString(bSdkConst.EnvSsoRedirectURI, ""), nil
		},
	}
}

func readWellKnownURI(wellKnown map[string]any, field string) string {
	value, exist := wellKnown[field]
	if !exist {
		return ""
	}

	uri, ok := value.(string)
	if !ok {
		return ""
	}

	return uri
}
