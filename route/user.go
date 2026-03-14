package bSdkRoute

import (
	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx-labs/beacon-sso-sdk/handler"
)

// UserRouter 注册用户相关路由
//
// 该路由组包含以下端点：
//   - GET /user/userinfo - 获取当前用户信息（需要认证）
func (r *Route) UserRouter(route *gin.RouterGroup) {
	group := route.Group("/user")

	userHandler := bSdkHandler.NewUserHandler(r.ctx)

	group.GET("/userinfo", userHandler.GetCurrentUser)
}
