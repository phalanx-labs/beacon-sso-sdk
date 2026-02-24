package bSdkUtil

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
	"golang.org/x/oauth2"
)

// GetOAuthConfig 从上下文中检索 OAuth 配置
//
// 该函数尝试从传入的上下文（context）中获取已注入的 `oauth2.Config` 对象。
// 它常用于处理 OAuth 回调或生成授权 URL 的业务逻辑中。
//
// 参数说明:
//   - ctx: 请求上下文对象，必须包含 `bSdkConst.CtxOAuthConfig` 键值。
//
// 返回值:
//   - *oauth2.Config: 从上下文中提取的 OAuth 配置实例。如果配置不存在则引发 panic。
//
// 注意: 如果在上下文中找不到对应的配置（即注入失败），该函数会记录错误日志并 panic。
func GetOAuthConfig(ctx context.Context) *oauth2.Config {
	get, err := xCtxUtil.Get[*oauth2.Config](ctx, bSdkConst.CtxOAuthConfig)
	if err != nil {
		xLog.WithName(xLog.NamedUTIL).Error(ctx, err.ErrorMessage.String())
		panic(err.ErrorMessage.String())
	}
	return get
}

// GetOAuthUserinfoURI 从上下文中获取 OAuth 用户信息 URI
//
// 该函数尝试从 Gin 上下文中检索预先注入的 OAuth 用户信息 URI。
// 如果上下文中不存在该键值，函数将引发 panic 并记录错误日志。
//
// 参数说明:
//   - ctx: Gin 上下文对象，用于传递请求范围的数据。
//
// 返回值:
//   - string: 从上下文中检索到的 OAuth 用户信息 URI 字符串。
//
// 注意: 此函数依赖于中间件或前置逻辑将 `bSdkConst.CtxOAuthUserinfoURI` 键注入到上下文中。
// 如果获取失败，程序将 panic，通常意味着中间件配置缺失或执行顺序错误。
func GetOAuthUserinfoURI(ctx context.Context) string {
	get, err := xCtxUtil.Get[string](ctx, bSdkConst.CtxOAuthUserinfoURI)
	if err != nil {
		xLog.WithName(xLog.NamedUTIL).Error(ctx, err.ErrorMessage.String())
		panic(err.ErrorMessage.String())
	}
	return get
}
