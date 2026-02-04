package bSdkMiddle

import "github.com/gin-gonic/gin"

// CheckAuth 检查用户身份认证信息
//
// CheckAuth 是一个 Gin 中间件，用于验证请求头中携带的身份认证凭据（如 JWT Token）//。
// 它会解析 Token 并将其中的用户信息（如 User ID, 角色）写入上下文，供后续的 Handler 使用。
//
// 如果 Token 缺失、格式错误或验证失败，该函数将中断请求链，并立即返回相应的 401 或 403 错误响应。
//
// 认证失败响应:
//
//	401: 未提供认证凭据或凭据无效。
//	403: 凭据有效，但无权访问该资源（如果包含权限检查逻辑）。
//
// 上下文设置:
//
//	验证通过后，通常会在 `ctx.Set("userID", ...)` 中设置用户标识。 CheckAuth 验证用户身份的中间件
//
// CheckAuth 是一个 Gin 中间件函数，用于在请求处理前验证用户的身份认证信息。
// 它会解析请求头中的 Token 或 Session，验证其有效性，并根据验证结果决定是否继续处理请求。
//
// 如果验证成功，会将用户信息注入到上下文中，供后续 Handler 使用。
//
// 参数说明:
//   - ctx: *gin.Context, Gin 的上下文对象，封装了请求和响应信息。
//
// 注意: 如果验证失败，通常会中断请求处理并直接返回 401 Unauthorized 错误。
func CheckAuth(ctx *gin.Context) {

}
