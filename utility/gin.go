package bSdkUtil

import (
	xHttp "github.com/bamboo-services/bamboo-base-go/major/http"
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
