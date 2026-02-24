package bSdkRepo

import (
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/gin-gonic/gin"
	bSdkModels "github.com/phalanx-labs/beacon-sso-sdk/models"
	bSdkCache "github.com/phalanx-labs/beacon-sso-sdk/repository/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// OAuthRepo OAuth 数据仓储层，负责管理与身份认证相关的持久化数据及缓存交互。
//
// 该结构体封装了对数据库的读写操作，并集成了缓存机制以提升查询性能。
// 它通过组合 `gorm.DB` 进行 ORM 映射，利用 `OAuthCache` 处理高频访问数据的缓存。
type OAuthRepo struct {
	db    *gorm.DB
	cache *bSdkCache.OAuthCache
	log   *xLog.LogNamedLogger
}

// NewOAuthRepo 创建并初始化一个 OAuth 数据仓储实例。
//
// 该函数接受 GORM 数据库连接和 Redis 客户端作为参数，用于底层的持久化存储
// 和缓存操作。它内部会初始化关联的缓存适配器（默认 TTL 为 30 分钟）
// 以及带有命名上下文的日志记录器。
//
// 参数:
//   - db: 已初始化的 GORM 数据库实例，用于数据库交互。
//   - rdb: 已初始化的 Redis 客户端，用于缓存数据。
//
// 返回值:
//   - *OAuthRepo: 配置完成的 OAuth 仓储实例指针。
func NewOAuthRepo(db *gorm.DB, rdb *redis.Client) *OAuthRepo {
	return &OAuthRepo{
		db:    db,
		cache: bSdkCache.NewOAuthCache(rdb),
		log:   xLog.WithName(xLog.NamedREPO, "OAuthRepo"),
	}
}

func (r *OAuthRepo) Store(ctx *gin.Context, state string, verifier string) *xError.Error {
	if state == "" || verifier == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "状态或验证器为空", false, nil)
	}

	fields := &bSdkModels.CacheOAuth{State: state, Verifier: verifier}
	if err := r.cache.SetAllStruct(ctx, state, fields); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "写入 OAuth 缓存失败", false, err)
	}

	return nil
}

func (r *OAuthRepo) Get(ctx *gin.Context, state string) (*bSdkModels.CacheOAuth, *xError.Error) {
	if state == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "状态为空", false, nil)
	}

	values, err := r.cache.GetAllStruct(ctx, state)
	if err != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "读取 OAuth 缓存失败", false, err)
	}

	return values, nil
}

func (r *OAuthRepo) Delete(ctx *gin.Context, state string) *xError.Error {
	if state == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "状态为空", false, nil)
	}

	if err := r.cache.Delete(ctx, state); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "删除 OAuth 缓存失败", false, err)
	}

	return nil
}
