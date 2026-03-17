package bSdkRoute

import (
	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx-labs/beacon-sso-sdk/handler"
	bSdkMiddle "github.com/phalanx-labs/beacon-sso-sdk/middleware"
)

// UserRouter 注册用户相关路由
//
// 该路由组包含以下端点：
//   - GET /user/userinfo - 获取当前用户信息（需要认证）
//   - GET /user/by-id - 根据ID获取用户信息（需要认证）
func (r *Route) UserRouter(route *gin.RouterGroup) {
	group := route.Group("/user")

	userHandler := bSdkHandler.NewUserHandler(r.ctx)

	// 需要认证的接口
	group.GET("/userinfo", bSdkMiddle.CheckAuth(r.ctx), userHandler.GetCurrentUser)
	group.GET("/by-id", bSdkMiddle.CheckAuth(r.ctx), userHandler.GetUserByID)
}
