package bSdkConst

import xEnv "github.com/bamboo-services/bamboo-base-go/env"

const (
	EnvSsoClientID            xEnv.EnvKey = "SSO_CLIENT_ID"             // 单点登录客户端 ID
	EnvSsoClientSecret        xEnv.EnvKey = "SSO_CLIENT_SECRET"         // 单点登录客户端 Secret
	EnvSsoWellKnownURI        xEnv.EnvKey = "SSO_WELL_KNOWN_URI"        // 单点登录元数据端点
	EnvSsoRedirectURI         xEnv.EnvKey = "SSO_REDIRECT_URI"          // 单点登录回调地址
	EnvSsoEndpointAuthURI     xEnv.EnvKey = "SSO_ENDPOINT_AUTH_URI"     // 单点登录授权端点
	EnvSsoEndpointTokenURI    xEnv.EnvKey = "SSO_ENDPOINT_TOKEN_URI"    // 单点登录令牌端点
	EnvSsoEndpointUserinfoURI xEnv.EnvKey = "SSO_ENDPOINT_USERINFO_URI" // 单点登录用户信息端点
)
