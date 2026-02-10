package bSdkRepo

import (
	"encoding/json"

	xError "github.com/bamboo-services/bamboo-base-go/error"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	"github.com/gin-gonic/gin"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	bSdkCache "github.com/phalanx/beacon-sso-sdk/repository/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// IntrospectionRepo 业务层 Introspection 数据仓储层，负责管理令牌自省结果缓存。
//
// 该结构体专注于 Introspection 的缓存管理，提供读取、存储和删除操作。
type IntrospectionRepo struct {
	db    *gorm.DB
	cache *bSdkCache.IntrospectionCache
	log   *xLog.LogNamedLogger
}

// NewIntrospectionRepo 创建并初始化一个 Introspection 仓储实例。
//
// 参数:
//   - db: 已初始化的 GORM 数据库实例（备用）。
//   - rdb: 已初始化的 Redis 客户端，用于缓存数据。
//
// 返回值:
//   - *IntrospectionRepo: 配置完成的 Introspection 仓储实例指针。
func NewIntrospectionRepo(db *gorm.DB, rdb *redis.Client) *IntrospectionRepo {
	return &IntrospectionRepo{
		db:    db,
		cache: bSdkCache.NewIntrospectionCache(rdb),
		log:   xLog.WithName(xLog.NamedLOGC),
	}
}

// GetCache 从缓存中获取令牌自省结果
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - tokenType: 令牌类型（如 "access_token"、"refresh_token"）。
//   - token: 令牌值。
//
// 返回值:
//   - *bSdkModels.OAuthIntrospection: 缓存的令牌自省结果对象。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (r *IntrospectionRepo) GetCache(ctx *gin.Context, tokenType string, token string) (*bSdkModels.OAuthIntrospection, bool, error) {
	if tokenType == "" || token == "" {
		return nil, false, nil
	}

	cacheKey := tokenType + ":" + token
	cacheValue, exists, err := r.cache.GetAllStruct(ctx, cacheKey)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	introspection := &bSdkModels.OAuthIntrospection{
		Active:    cacheValue.Active,
		TokenType: cacheValue.TokenType,
		Exp:       cacheValue.Exp,
		Expiry:    cacheValue.Expiry,
		ExpiresIn: cacheValue.ExpiresIn,
		IsExpired: cacheValue.IsExpired,
	}

	if cacheValue.Raw != "" {
		var raw map[string]any
		if err := json.Unmarshal([]byte(cacheValue.Raw), &raw); err != nil {
			return nil, false, err
		}
		introspection.Raw = raw
	}

	return introspection, true, nil
}

// StoreCache 将令牌自省结果存储到缓存
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - tokenType: 令牌类型（如 "access_token"、"refresh_token"）。
//   - token: 令牌值。
//   - introspection: 要缓存的令牌自省结果对象。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (r *IntrospectionRepo) StoreCache(ctx *gin.Context, tokenType string, token string, introspection *bSdkModels.OAuthIntrospection) error {
	if tokenType == "" || token == "" || introspection == nil {
		return nil
	}

	cacheIntrospection := &bSdkModels.CacheBusinessIntrospection{
		Active:    introspection.Active,
		TokenType: introspection.TokenType,
		Exp:       introspection.Exp,
		Expiry:    introspection.Expiry,
		ExpiresIn: introspection.ExpiresIn,
		IsExpired: introspection.IsExpired,
	}

	if introspection.Raw != nil {
		rawJSON, err := json.Marshal(introspection.Raw)
		if err != nil {
			return err
		}
		cacheIntrospection.Raw = string(rawJSON)
	}

	cacheKey := tokenType + ":" + token
	return r.cache.SetAllStruct(ctx, cacheKey, cacheIntrospection)
}

// DeleteCache 删除令牌自省缓存
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - tokenType: 令牌类型（如 "access_token"、"refresh_token"）。
//   - token: 令牌值。
//
// 返回值:
//   - *xError.Error: 操作过程中发生的错误。
func (r *IntrospectionRepo) DeleteCache(ctx *gin.Context, tokenType string, token string) *xError.Error {
	if tokenType == "" || token == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌类型或令牌为空", false, nil)
	}

	cacheKey := tokenType + ":" + token
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "删除令牌自省缓存失败", false, err)
	}

	return nil
}
