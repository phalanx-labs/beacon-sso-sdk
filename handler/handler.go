package bSdkHandler

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// handler 是应用程序的 HTTP 处理器结构体。
//
// 它封装了处理请求所需的核心依赖组件，包括数据库连接 (db)、
// 缓存客户端 (rdb) 和日志记录器 (log)。
//
// 成员变量:
//   - db:  *gorm.DB 类型的数据库实例，用于执行持久化操作。
//   - rdb: *redis.Client 类型的缓存客户端，用于高速数据缓存。
//   - log: *xLog.LogNamedLogger 类型的日志记录器，用于结构化日志输出。
type handler struct {
	db      *gorm.DB             // GORM 数据库实例
	rdb     *redis.Client        // Redis 客户端实例
	log     *xLog.LogNamedLogger // 日志实例
	service *service
}

// service 业务逻辑处理层的核心结构体
//
// 它负责封装应用程序的核心业务规则和逻辑，作为 HTTP 处理器（handler）与底层数据访问层之间的桥梁。
// 通常在 `registerService` 方法中初始化并注入到处理器中。
type service struct {
}

// registerService 注册 Service 的内容
func (h *handler) registerService(ctx context.Context) {
	h.service = &service{}
}

// =============
//  Handler注册
// =============

// AuthHandler 是登录回调处理的请求器。
type AuthHandler handler

// NewAuthHandler 创建并初始化一个 AuthHandler 实例
func NewAuthHandler(db *gorm.DB, rdb *redis.Client, ctx context.Context) *AuthHandler {
	newHandler := &AuthHandler{
		db:  db,
		rdb: rdb,
		log: xLog.WithName(xLog.NamedCONT),
	}
	(*handler)(newHandler).registerService(ctx)
	return newHandler
}
