package bSdkRoute

import (
	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx-labs/beacon-sso-sdk/handler"
	bSdkMiddle "github.com/phalanx-labs/beacon-sso-sdk/middleware"
)

// AccountRouter 注册账户相关路由
//
// 该路由组包含以下端点：
//   - POST /account/register/email - 邮箱注册（公开）
//   - POST /account/login/password - 密码登录（公开）
//   - POST /account/password/change - 修改密码（需要认证）
//   - POST /account/token/revoke - 注销令牌（需要认证）
func (r *Route) AccountRouter(route *gin.RouterGroup) {
	group := route.Group("/account")

	accountHandler := bSdkHandler.NewAccountHandler(r.ctx)

	// 公开接口
	group.POST("/register/email", accountHandler.RegisterByEmail)
	group.POST("/login/password", accountHandler.PasswordLogin)

	// 需要认证的接口
	group.POST("/password/change", bSdkMiddle.CheckAuth(r.ctx), accountHandler.ChangePassword)
	group.POST("/token/revoke", bSdkMiddle.CheckAuth(r.ctx), accountHandler.RevokeToken)
}
