package bSdkRoute

import (
	"context"

	"github.com/gin-gonic/gin"
	bSdkHandler "github.com/phalanx/beacon-sso-sdk/handler"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Route struct {
	db      *gorm.DB        // GORM 数据库实例
	rdb     *redis.Client   // Redis 客户端实例
	context context.Context // 上下文，用于控制取消和超时
}

func NewOAuthRoute(ctx context.Context, db *gorm.DB, rdb *redis.Client) *Route {
	return &Route{
		db:      db,
		rdb:     rdb,
		context: ctx,
	}
}

func (r *Route) OAuthRouter(route *gin.RouterGroup) {
	group := route.Group("/oauth")

	authHandler := bSdkHandler.NewAuthHandler(r.db, r.rdb, r.context)

	group.GET("/callback", authHandler.Callback)
}
