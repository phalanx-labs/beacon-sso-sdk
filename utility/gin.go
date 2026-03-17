package bSdkUtil

import (
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	"github.com/gin-gonic/gin"
)

// GetAuthorization 从 Gin 上下文中提取 Authorization 请求头的值。
//
// 该函数尝试从上下文中获取 `Authorization` 头信息。如果存在则返回其字符串值，否则返回空字符串。
//
// 参数:
//   - ctx: Gin 的上下文对象，用于访问请求元数据。
//
// 返回值:
//   - string: 返回 Authorization 头的值；如果未设置则返回空字符串。
func GetAuthorization(ctx *gin.Context) string {
	get, exist := ctx.Get(xHttp.HeaderAuthorization.String())
	if exist {
		return get.(string)
	}
	return ""
}

// GetAccessToken 从 Gin Context 获取已验证的 Access Token
//
// 该方法用于从 Gin 上下文中提取经过 CheckAuth 中间件验证的 Access Token。
// 配合 CheckAuth 中间件使用，避免 Handler 层重复获取和验证 Token。
//
// 参数:
//   - ctx: Gin 的上下文对象，必须经过 CheckAuth 中间件处理。
//
// 返回值:
//   - string: 访问令牌字符串。
//   - *xError.Error: 如果 Token 不存在则返回错误。
func GetAccessToken(ctx *gin.Context) (string, *xError.Error) {
	token, exists := ctx.Get(xHttp.HeaderAuthorization.String())
	if !exists {
		return "", xError.NewError(ctx, xError.ParameterEmpty, "需要访问令牌参数", false, nil)
	}
	return token.(string), nil
}
