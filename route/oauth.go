package bSdkRoute

import (
	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx-labs/beacon-sso-sdk/handler"
)

// OAuthRouter 注册 OAuth 相关路由
//
// 该路由组包含以下端点：
//   - GET /oauth/login - OAuth 登录跳转
//   - GET /oauth/callback - OAuth 登录回调
//   - POST /oauth/logout - OAuth 登出
func (r *Route) OAuthRouter(route *gin.RouterGroup) {
	group := route.Group("/oauth")

	authHandler := bSdkHandler.NewAuthHandler(r.ctx)

	group.GET("/login", authHandler.Login)
	group.GET("/callback", authHandler.Callback)
	group.POST("/logout", authHandler.Logout)
}
