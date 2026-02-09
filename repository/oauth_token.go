package bSdkRepo

import (
	xError "github.com/bamboo-services/bamboo-base-go/error"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	"github.com/gin-gonic/gin"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	bSdkCache "github.com/phalanx/beacon-sso-sdk/repository/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// OAuthTokenRepo OAuth 令牌数据仓储层，负责管理已换取的访问令牌缓存。
//
// 该结构体专注于 Token 的缓存管理，与 OAuthRepo 分离以保持职责单一。
type OAuthTokenRepo struct {
	db    *gorm.DB
	cache *bSdkCache.OAuthTokenCache
	log   *xLog.LogNamedLogger
}

// NewOAuthTokenRepo 创建并初始化一个 OAuth 令牌仓储实例。
//
// 参数:
//   - db: 已初始化的 GORM 数据库实例（备用）。
//   - rdb: 已初始化的 Redis 客户端，用于缓存数据。
//
// 返回值:
//   - *OAuthTokenRepo: 配置完成的令牌仓储实例指针。
func NewOAuthTokenRepo(db *gorm.DB, rdb *redis.Client) *OAuthTokenRepo {
	return &OAuthTokenRepo{
		db:    db,
		cache: bSdkCache.NewOAuthTokenCache(rdb),
		log:   xLog.WithName(xLog.NamedLOGC),
	}
}

func (r *OAuthTokenRepo) Store(ctx *gin.Context, token *bSdkModels.CacheOAuthToken) *xError.Error {
	if token == nil || token.AccessToken == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	if err := r.cache.SetAllStruct(ctx, token.AccessToken, token); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "写入令牌缓存失败", false, err)
	}

	return nil
}

func (r *OAuthTokenRepo) Get(ctx *gin.Context, accessToken string) (*bSdkModels.CacheOAuthToken, *xError.Error) {
	if accessToken == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	values, err := r.cache.GetAllStruct(ctx, accessToken)
	if err != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "读取令牌缓存失败", false, err)
	}

	return values, nil
}

func (r *OAuthTokenRepo) Delete(ctx *gin.Context, accessToken string) *xError.Error {
	if accessToken == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	if err := r.cache.Delete(ctx, accessToken); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "删除令牌缓存失败", false, err)
	}

	return nil
}
