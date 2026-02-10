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

// UserinfoRepo 业务层 Userinfo 数据仓储层，负责管理用户信息缓存。
//
// 该结构体专注于 Userinfo 的缓存管理，提供读取、存储和删除操作。
type UserinfoRepo struct {
	db    *gorm.DB
	cache *bSdkCache.UserinfoCache
	log   *xLog.LogNamedLogger
}

// NewUserinfoRepo 创建并初始化一个 Userinfo 仓储实例。
//
// 参数:
//   - db: 已初始化的 GORM 数据库实例（备用）。
//   - rdb: 已初始化的 Redis 客户端，用于缓存数据。
//
// 返回值:
//   - *UserinfoRepo: 配置完成的 Userinfo 仓储实例指针。
func NewUserinfoRepo(db *gorm.DB, rdb *redis.Client) *UserinfoRepo {
	return &UserinfoRepo{
		db:    db,
		cache: bSdkCache.NewUserinfoCache(rdb),
		log:   xLog.WithName(xLog.NamedLOGC),
	}
}

// GetCache 从缓存中获取用户信息
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//
// 返回值:
//   - *bSdkModels.OAuthUserinfo: 缓存的用户信息对象。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (r *UserinfoRepo) GetCache(ctx *gin.Context, accessToken string) (*bSdkModels.OAuthUserinfo, bool, error) {
	if accessToken == "" {
		return nil, false, nil
	}

	cacheValue, exists, err := r.cache.GetAllStruct(ctx, accessToken)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	userinfo := &bSdkModels.OAuthUserinfo{
		Sub:               cacheValue.Sub,
		Nickname:          cacheValue.Nickname,
		PreferredUsername: cacheValue.PreferredUsername,
		Email:             cacheValue.Email,
		Phone:             cacheValue.Phone,
	}

	if cacheValue.Raw != "" {
		var raw map[string]any
		if err := json.Unmarshal([]byte(cacheValue.Raw), &raw); err != nil {
			return nil, false, err
		}
		userinfo.Raw = raw
	}

	return userinfo, true, nil
}

// StoreCache 将用户信息存储到缓存
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - userinfo: 要缓存的用户信息对象。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (r *UserinfoRepo) StoreCache(ctx *gin.Context, accessToken string, userinfo *bSdkModels.OAuthUserinfo) error {
	if accessToken == "" || userinfo == nil {
		return nil
	}

	cacheUserinfo := &bSdkModels.CacheBusinessUserinfo{
		Sub:               userinfo.Sub,
		Nickname:          userinfo.Nickname,
		PreferredUsername: userinfo.PreferredUsername,
		Email:             userinfo.Email,
		Phone:             userinfo.Phone,
	}

	if userinfo.Raw != nil {
		rawJSON, err := json.Marshal(userinfo.Raw)
		if err != nil {
			return err
		}
		cacheUserinfo.Raw = string(rawJSON)
	}

	return r.cache.SetAllStruct(ctx, accessToken, cacheUserinfo)
}

// DeleteCache 删除用户信息缓存
//
// 参数:
//   - ctx: Gin 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//
// 返回值:
//   - *xError.Error: 操作过程中发生的错误。
func (r *UserinfoRepo) DeleteCache(ctx *gin.Context, accessToken string) *xError.Error {
	if accessToken == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	if err := r.cache.Delete(ctx, accessToken); err != nil {
		return xError.NewError(ctx, xError.OperationFailed, "删除用户信息缓存失败", false, err)
	}

	return nil
}
