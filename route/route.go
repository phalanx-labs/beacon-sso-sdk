package bSdkRoute

import (
	"context"

	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx/beacon-sso-sdk/handler"
)

type Route struct {
	ctx context.Context // 上下文，用于控制取消和超时
}

func NewOAuthRoute(ctx context.Context) *Route {
	return &Route{
		ctx: ctx,
	}
}

func (r *Route) OAuthRouter(route *gin.RouterGroup) {
	group := route.Group("/oauth")

	authHandler := bSdkHandler.NewAuthHandler(r.ctx)

	group.GET("/login", authHandler.Login)
	group.GET("/callback", authHandler.Callback)
	group.POST("/logout", authHandler.Logout)
}
