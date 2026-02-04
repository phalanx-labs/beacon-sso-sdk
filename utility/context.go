package bSdkUtil

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	"golang.org/x/oauth2"
)

// GetOAuthConfig 从 Gin 上下文中提取 OAuth 配置实例。
//
// 该函数是一个辅助方法，用于从当前请求的上下文中检索预先注入的 `*oauth2.Config` 对象。
// 它严格检查上下文键 `bSdkConst.CtxOAuthConfig`。
// 如果配置不存在，记录错误日志并引发 panic。
//
// 参数说明:
//   - c: Gin 的上下文对象，携带请求作用域的数据。
//
// 返回值:
//   - *oauth2.Config: 从上下文中提取的 OAuth2 配置对象。</think>// GetOAuthConfig 从 Gin 上下文中提取 OAuth 配置实例
//
// 该函数尝试从当前请求的上下文中检索已注入的 *oauth2.Config 对象。
// 它使用键 "oauth_config" 进行查找。如果上下文中不存在该配置，
// 函数会记录错误日志并直接引发 panic，终止请求处理。
//
// 参数:
//   - c: Gin 的上下文对象
//
// 返回值:
//   - *oauth2.Config: OAuth2 配置对象
func GetOAuthConfig(ctx *gin.Context) *oauth2.Config {
	value, exists := ctx.Get(bSdkConst.CtxOAuthConfig)
	if exists {
		return value.(*oauth2.Config)
	}
	slog.ErrorContext(ctx, "在上下文中找不到 OAuth 配置，真的注入成功了吗？")
	panic("在上下文中找不到 OAuth 配置，真的注入成功了吗？")
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
func GetOAuthUserinfoURI(ctx *gin.Context) string {
	getString := ctx.GetString(bSdkConst.CtxOAuthUserinfoURI)
	if getString != "" {
		return getString
	}
	slog.ErrorContext(ctx, "在上下文中找不到 OAuth 用户信息 URI，真的注入成功了吗？")
	panic("在上下文中找不到 OAuth 用户信息 URI，真的注入成功了吗？")
}
